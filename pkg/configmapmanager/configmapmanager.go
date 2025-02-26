// Copyright 2025 Alejandro de Cock Buning; Ivan Vidal; Francisco Valera; Diego R. Lopez.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configmapmanager

import (
	"context"
	"fmt"
	"net"

	"github.com/Networks-it-uc3m/l2sm-dns/internal/env"
	"github.com/Networks-it-uc3m/l2sm-dns/pkg/corefile"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CoreDNSManager manages CoreDNS ConfigMaps and DNS entries.
type CoreDNSManager struct {
	clientset *kubernetes.Clientset
	namespace string
	configMap string
}

// NewCoreDNSManager creates a new instance of CoreDNSManager.
func NewCoreDNSManager(namespace, configMap string, k8sConfig *rest.Config) (*CoreDNSManager, error) {

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %v", err)
	}

	return &CoreDNSManager{
		clientset: clientset,
		namespace: namespace,
		configMap: configMap,
	}, nil
}

// GetConfigMap retrieves the CoreDNS ConfigMap.
func (c *CoreDNSManager) GetConfigMap(ctx context.Context) (*v1.ConfigMap, error) {
	return c.clientset.CoreV1().ConfigMaps(c.namespace).Get(ctx, c.configMap, metav1.GetOptions{})
}

func (c *CoreDNSManager) AddDNSEntryToConfigMap(ctx context.Context, updatedData map[string]string) error {
	cfg, err := c.GetConfigMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ConfigMap: %w", err)
	}

	coreFileString, ok := cfg.Data["Corefile"]
	if !ok {
		return fmt.Errorf("corefile not found in ConfigMap data")
	}

	// Parse the existing Corefile into a structure we can manipulate.
	cf, err := corefile.New(coreFileString)
	if err != nil {
		return fmt.Errorf("could not parse existing corefile: %v", err)
	}

	// Retrieve the server block for your inter-domain port.
	interDomainServer, ok := cf.GetServer(env.GetInterDomainDomPort())
	if !ok {
		return fmt.Errorf("could not find inter-domain port '%v' in Corefile, check corefile syntax", env.GetInterDomainDomPort())
	}

	// Get the "hosts" plugin inside the inter-domain server block.
	hostsPlugin, ok := interDomainServer.GetPlugin("hosts")
	if !ok {
		return fmt.Errorf("could not find 'hosts' plugin in the inter-domain server block")
	}

	// Convert updatedData from a single domain per IP into a map of IP->[]domains
	// so we can pass it into the plugin’s AddHostsEntries method.
	newEntries := map[string][]string{}
	for ip, domain := range updatedData {
		// Basic sanity checks, for example ensuring IP is valid:
		if net.ParseIP(ip) == nil {
			return fmt.Errorf("invalid IP address in updatedData: %q", ip)
		}
		newEntries[ip] = append(newEntries[ip], domain)
	}

	// Merge these new mappings into the existing "hosts" plugin.
	if err := hostsPlugin.AddHostsEntries(newEntries); err != nil {
		return fmt.Errorf("failed to add host entries: %v", err)
	}

	// Serialize the modified Corefile back to string.
	updatedCoreFileString := cf.ToString()
	cfg.Data["Corefile"] = updatedCoreFileString

	_, err = c.clientset.CoreV1().ConfigMaps(c.namespace).
		Update(ctx, cfg, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap with new Corefile: %v", err)
	}

	return nil
}

// RemoveDNSRecords removes one or more domains from the “hosts” plugin records
// in the inter-domain server block. If an IP loses all domains, that IP entry
// is removed entirely.
func (c *CoreDNSManager) RemoveDNSRecords(ctx context.Context, removals map[string][]string) error {
	cfg, err := c.GetConfigMap(ctx)
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

	// Remove the specified domains for each IP in "hosts".
	if err := hostsPlugin.RemoveHostsEntries(removals); err != nil {
		return fmt.Errorf("failed to remove host entries: %v", err)
	}

	cfg.Data["Corefile"] = cf.ToString()
	_, err = c.clientset.CoreV1().ConfigMaps(c.namespace).
		Update(ctx, cfg, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap after removing entries: %v", err)
	}

	return nil
}

// ListDNSRecords returns the current IP->domains mapping from the inter-domain server’s “hosts” plugin.
func (c *CoreDNSManager) ListDNSRecords(ctx context.Context) (map[string][]string, error) {
	cfg, err := c.GetConfigMap(ctx)
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

// AddDNSEntry adds a new DNS entry to the CoreDNS ConfigMap.
func (c *CoreDNSManager) AddDNSEntry(ctx context.Context, dnsName, ipAddress string) error {
	updatedData := map[string]string{
		ipAddress: dnsName,
	}
	return c.AddDNSEntryToConfigMap(ctx, updatedData)
}

// RemoveDNSEntry removes a DNS entry from the CoreDNS ConfigMap.
func (c *CoreDNSManager) RemoveDNSEntry(ctx context.Context, key string) error {
	configMap, err := c.GetConfigMap(ctx)
	if err != nil {
		return err
	}

	// Delete the entry from the ConfigMap.
	delete(configMap.Data, key)

	_, err = c.clientset.CoreV1().ConfigMaps(c.namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove DNS entry: %v", err)
	}

	return nil
}

func (c *CoreDNSManager) AddServerToConfigMap(ctx context.Context, domainName, serverDomain, serverPort string) error {

	cfg, err := c.GetConfigMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ConfigMap: %w", err)
	}

	coreFileString, ok := cfg.Data["Corefile"]
	if !ok {
		return fmt.Errorf("corefile not found in ConfigMap data")
	}

	// Parse the existing Corefile into a structure we can manipulate.
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

	err = cf.AddServer(newServer)
	if err != nil {
		return fmt.Errorf("failed to add server in corefile: %v", err)
	}
	// Serialize the modified Corefile back to string.
	updatedCoreFileString := cf.ToString()
	cfg.Data["Corefile"] = updatedCoreFileString

	_, err = c.clientset.CoreV1().ConfigMaps(c.namespace).
		Update(ctx, cfg, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap with new Corefile: %v", err)
	}

	return nil
}
