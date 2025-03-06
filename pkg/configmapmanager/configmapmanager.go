// Copyright 2025 ...
// Licensed under the Apache License, Version 2.0; see LICENSE.

package configmapmanager

import (
	"context"
	"fmt"
	"net"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Networks-it-uc3m/l2sm-dns/internal/env"
	"github.com/Networks-it-uc3m/l2sm-dns/pkg/corefile"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// DNSManager defines the interface for managing CoreDNS ConfigMaps.
type DNSManager interface {
	GetConfigMap(ctx context.Context) (*v1.ConfigMap, error)
	AddDNSEntryToConfigMap(ctx context.Context, updatedData map[string]string) error
	RemoveDNSRecords(ctx context.Context, removals map[string][]string) error
	ListDNSRecords(ctx context.Context) (map[string][]string, error)
	AddDNSEntry(ctx context.Context, dnsName, ipAddress string) error
	RemoveDNSEntry(ctx context.Context, key, ipAddress string) error
	AddServerToConfigMap(ctx context.Context, domainName, serverDomain, serverPort string) error
}

// ConfigMapClient is an abstraction over different ways to interact with a ConfigMap.
type ConfigMapClient interface {
	Get(ctx context.Context) (*v1.ConfigMap, error)
	Update(ctx context.Context, cfg *v1.ConfigMap) error
}

// --- Controller-runtime based client ---
type crConfigMapClient struct {
	client    client.Client
	namespace string
	name      string
}

func newCRConfigMapClient(namespace, name string, crClient client.Client) ConfigMapClient {
	return &crConfigMapClient{
		client:    crClient,
		namespace: namespace,
		name:      name,
	}
}

func (c *crConfigMapClient) Get(ctx context.Context) (*v1.ConfigMap, error) {
	cfg := &v1.ConfigMap{}
	key := client.ObjectKey{Namespace: c.namespace, Name: c.name}
	if err := c.client.Get(ctx, key, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *crConfigMapClient) Update(ctx context.Context, cfg *v1.ConfigMap) error {
	return c.client.Update(ctx, cfg)
}

// --- Clientset based client ---
type clientsetConfigMapClient struct {
	clientset *kubernetes.Clientset
	namespace string
	name      string
}

func newClientsetConfigMapClient(namespace, name string, k8sConfig *rest.Config) (ConfigMapClient, error) {
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}
	return &clientsetConfigMapClient{
		clientset: clientset,
		namespace: namespace,
		name:      name,
	}, nil
}

func (c *clientsetConfigMapClient) Get(ctx context.Context) (*v1.ConfigMap, error) {
	return c.clientset.CoreV1().ConfigMaps(c.namespace).Get(ctx, c.name, metav1.GetOptions{})
}

func (c *clientsetConfigMapClient) Update(ctx context.Context, cfg *v1.ConfigMap) error {
	_, err := c.clientset.CoreV1().ConfigMaps(c.namespace).Update(ctx, cfg, metav1.UpdateOptions{})
	return err
}

// --- CoreDNSManager Implementation ---
type coreDNSManager struct {
	cmClient  ConfigMapClient
	namespace string
	configMap string
}

// NewDNSManager is the factory function that creates a DNSManager.
// If crClient is provided (non-nil), it uses the controller-runtime client;
// otherwise, it falls back to using the standard Kubernetes clientset.
func NewDNSManager(namespace, configMap string, k8sConfig *rest.Config, crClient client.Client) (DNSManager, error) {
	var cmClient ConfigMapClient
	var err error
	if crClient != nil {
		cmClient = newCRConfigMapClient(namespace, configMap, crClient)
	} else {
		cmClient, err = newClientsetConfigMapClient(namespace, configMap, k8sConfig)
		if err != nil {
			return nil, err
		}
	}
	return &coreDNSManager{
		cmClient:  cmClient,
		namespace: namespace,
		configMap: configMap,
	}, nil
}

// GetConfigMap retrieves the CoreDNS ConfigMap using the configured client.
func (m *coreDNSManager) GetConfigMap(ctx context.Context) (*v1.ConfigMap, error) {
	return m.cmClient.Get(ctx)
}

func (m *coreDNSManager) AddDNSEntryToConfigMap(ctx context.Context, updatedData map[string]string) error {
	cfg, err := m.GetConfigMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ConfigMap: %w", err)
	}

	coreFileString, ok := cfg.Data["Corefile"]
	if !ok {
		return fmt.Errorf("corefile not found in ConfigMap data")
	}

	cf, err := corefile.New(coreFileString)
	if err != nil {
		return fmt.Errorf("could not parse existing corefile: %v", err)
	}

	interDomainServer, ok := cf.GetServer(env.GetInterDomainDomPort())
	if !ok {
		return fmt.Errorf("could not find inter-domain port '%v' in Corefile, check corefile syntax", env.GetInterDomainDomPort())
	}

	hostsPlugin, ok := interDomainServer.GetPlugin("hosts")
	if !ok {
		return fmt.Errorf("could not find 'hosts' plugin in the inter-domain server block")
	}

	// Convert updatedData to map[ip][]domain
	newEntries := map[string][]string{}
	for ip, domain := range updatedData {
		if net.ParseIP(ip) == nil {
			return fmt.Errorf("invalid IP address in updatedData: %q", ip)
		}
		newEntries[ip] = append(newEntries[ip], domain)
	}

	if err := hostsPlugin.AddHostsEntries(newEntries); err != nil {
		return fmt.Errorf("failed to add host entries: %v", err)
	}

	cfg.Data["Corefile"] = cf.ToString()
	return m.cmClient.Update(ctx, cfg)
}

func (m *coreDNSManager) RemoveDNSRecords(ctx context.Context, removals map[string][]string) error {
	cfg, err := m.GetConfigMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ConfigMap: %w", err)
	}

	coreFileString, ok := cfg.Data["Corefile"]
	if !ok {
		return fmt.Errorf("corefile not found in ConfigMap data")
	}

	cf, err := corefile.New(coreFileString)
	if err != nil {
		return fmt.Errorf("could not parse existing corefile: %v", err)
	}

	interDomainServer, ok := cf.GetServer(env.GetInterDomainDomPort())
	if !ok {
		return fmt.Errorf("could not find inter-domain server '%v' in Corefile", env.GetInterDomainDomPort())
	}

	hostsPlugin, ok := interDomainServer.GetPlugin("hosts")
	if !ok {
		return fmt.Errorf("could not find 'hosts' plugin in inter-domain server block")
	}

	// Validate IPs
	for ip := range removals {
		if net.ParseIP(ip) == nil {
			return fmt.Errorf("invalid IP address in removals: %q", ip)
		}
	}

	if err := hostsPlugin.RemoveHostsEntries(removals); err != nil {
		return fmt.Errorf("failed to remove host entries: %v", err)
	}

	cfg.Data["Corefile"] = cf.ToString()
	return m.cmClient.Update(ctx, cfg)
}

func (m *coreDNSManager) ListDNSRecords(ctx context.Context) (map[string][]string, error) {
	cfg, err := m.GetConfigMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get ConfigMap: %w", err)
	}

	coreFileString, ok := cfg.Data["Corefile"]
	if !ok {
		return nil, fmt.Errorf("corefile not found in ConfigMap data")
	}

	cf, err := corefile.New(coreFileString)
	if err != nil {
		return nil, fmt.Errorf("could not parse existing corefile: %v", err)
	}

	interDomainServer, ok := cf.GetServer(env.GetInterDomainDomPort())
	if !ok {
		return nil, fmt.Errorf("could not find inter-domain server '%v' in Corefile", env.GetInterDomainDomPort())
	}

	hostsPlugin, ok := interDomainServer.GetPlugin("hosts")
	if !ok {
		return nil, fmt.Errorf("could not find 'hosts' plugin in inter-domain server block")
	}

	return hostsPlugin.ListHostsEntries()
}

func (m *coreDNSManager) AddDNSEntry(ctx context.Context, dnsName, ipAddress string) error {
	updatedData := map[string]string{
		ipAddress: dnsName,
	}
	return m.AddDNSEntryToConfigMap(ctx, updatedData)
}

func (m *coreDNSManager) RemoveDNSEntry(ctx context.Context, key, ipAddress string) error {

	deletedEntries := make(map[string][]string)

	deletedEntries[ipAddress] = []string{key}

	cfg, err := m.GetConfigMap(ctx)
	if err != nil {
		return err
	}
	coreFileString, ok := cfg.Data["Corefile"]
	if !ok {
		return fmt.Errorf("corefile not found in ConfigMap data")
	}

	cf, err := corefile.New(coreFileString)
	if err != nil {
		return fmt.Errorf("could not parse existing corefile: %v", err)
	}

	interDomainServer, ok := cf.GetServer(env.GetInterDomainDomPort())
	if !ok {
		return fmt.Errorf("could not find inter-domain port '%v' in Corefile, check corefile syntax", env.GetInterDomainDomPort())
	}

	hostsPlugin, ok := interDomainServer.GetPlugin("hosts")
	if !ok {
		return fmt.Errorf("could not find 'hosts' plugin in the inter-domain server block")
	}

	if err := hostsPlugin.RemoveHostsEntries(deletedEntries); err != nil {
		return fmt.Errorf("failed to add host entries: %v", err)
	}

	cfg.Data["Corefile"] = cf.ToString()

	return m.cmClient.Update(ctx, cfg)
}

func (m *coreDNSManager) AddServerToConfigMap(ctx context.Context, domainName, serverDomain, serverPort string) error {
	cfg, err := m.GetConfigMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ConfigMap: %w", err)
	}

	coreFileString, ok := cfg.Data["Corefile"]
	if !ok {
		return fmt.Errorf("corefile not found in ConfigMap data")
	}

	cf, err := corefile.New(coreFileString)
	if err != nil {
		return fmt.Errorf("could not parse existing corefile: %v", err)
	}

	forwardPlugin := corefile.Plugin{
		Name: "forward",
		Args: []string{".", fmt.Sprintf("%s:%s", serverDomain, serverPort)},
	}
	newServer := corefile.Server{
		DomPorts: []string{domainName},
		Plugins:  []*corefile.Plugin{&forwardPlugin},
	}

	if err := cf.AddServer(newServer); err != nil {
		return fmt.Errorf("failed to add server in corefile: %v", err)
	}

	cfg.Data["Corefile"] = cf.ToString()
	return m.cmClient.Update(ctx, cfg)
}
