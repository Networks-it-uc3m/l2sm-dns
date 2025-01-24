package corednsmanager

import (
	"context"
	"fmt"

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

// UpdateConfigMap updates the CoreDNS ConfigMap with new DNS entries.
func (c *CoreDNSManager) UpdateConfigMap(ctx context.Context, updatedData map[string]string) error {
	configMap, err := c.GetConfigMap(ctx)
	if err != nil {
		return err
	}

	// Update the ConfigMap's data.
	for key, value := range updatedData {
		configMap.Data[key] = value
	}

	_, err = c.clientset.CoreV1().ConfigMaps(c.namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap: %v", err)
	}

	return nil
}

// AddDNSEntry adds a new DNS entry to the CoreDNS ConfigMap.
func (c *CoreDNSManager) AddDNSEntry(ctx context.Context, key, value string) error {
	updatedData := map[string]string{
		key: value,
	}
	return c.UpdateConfigMap(ctx, updatedData)
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
