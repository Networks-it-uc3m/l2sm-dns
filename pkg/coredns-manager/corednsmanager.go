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

package corednsmanager

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"

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

func (c *CoreDNSManager) UpdateConfigMap(ctx context.Context, updatedData map[string]string) error {
	configMap, err := c.GetConfigMap(ctx)
	if err != nil {
		return err
	}

	corefile, ok := configMap.Data["Corefile"]
	if !ok {
		return fmt.Errorf("Corefile not found in ConfigMap data")
	}

	updatedCorefile, err := updateCorefileWithCustomHosts(corefile, updatedData)
	if err != nil {
		return err
	}

	// Update the Corefile field.
	configMap.Data["Corefile"] = updatedCorefile

	// In production, update the ConfigMap in the cluster.
	_, err = c.clientset.CoreV1().ConfigMaps(c.namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap: %v", err)
	}

	return nil
}

// func (c *CoreDNSManager) UpdateConfigMap(ctx context.Context, updatedData map[string]string) error {
// 	configMap, err := c.GetConfigMap(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	corefile, ok := configMap.Data["Corefile"]
// 	if !ok {
// 		return fmt.Errorf("Corefile not found in ConfigMap data")
// 	}

// 	lines := strings.Split(corefile, "\n")

// 	// Look for an existing hosts block (assumed to start with "hosts {" and end with "}").
// 	hostsStart, hostsEnd := -1, -1
// 	inHostsBlock := false
// 	for i, line := range lines {
// 		trim := strings.TrimSpace(line)
// 		if strings.HasPrefix(trim, "hosts") && strings.HasSuffix(trim, "{") {
// 			hostsStart = i
// 			inHostsBlock = true
// 		} else if inHostsBlock && trim == "}" {
// 			hostsEnd = i
// 			inHostsBlock = false
// 			break
// 		}
// 	}

// 	if hostsStart != -1 && hostsEnd != -1 {
// 		// Process existing hosts block.
// 		// Build a set of existing mapping lines to avoid duplicates.
// 		existingMappings := make(map[string]bool)
// 		for i := hostsStart + 1; i < hostsEnd; i++ {
// 			trim := strings.TrimSpace(lines[i])
// 			// Skip empty lines and non-mapping directives (like fallthrough).
// 			if trim == "" || trim == "fallthrough" {
// 				continue
// 			}
// 			// Assume mapping lines are in the format "IP domain".
// 			existingMappings[trim] = true
// 		}

// 		// Prepare new mapping lines.
// 		var newMappings []string
// 		for ip, domain := range updatedData {
// 			mapping := fmt.Sprintf("%s %s", ip, domain)
// 			if !existingMappings[mapping] {
// 				// Use an indentation matching the hosts block (3 spaces in this example).
// 				newMappings = append(newMappings, fmt.Sprintf("   %s", mapping))
// 			}
// 		}

// 		// Insert new mappings before the closing brace of the hosts block.
// 		updatedBlock := append(lines[hostsStart+1:hostsEnd], newMappings...)
// 		// Rebuild the hosts block.
// 		newHostsBlock := append([]string{lines[hostsStart]}, append(updatedBlock, lines[hostsEnd])...)
// 		// Replace the original hosts block lines.
// 		newLines := append(lines[:hostsStart], append(newHostsBlock, lines[hostsEnd+1:]...)...)
// 		corefile = strings.Join(newLines, "\n")
// 	} else {
// 		// No existing hosts block found.
// 		// Create a new hosts block with our mappings and a default fallthrough directive.
// 		var hostsBlock []string
// 		hostsBlock = append(hostsBlock, "   hosts {")
// 		for ip, domain := range updatedData {
// 			hostsBlock = append(hostsBlock, fmt.Sprintf("      %s %s", ip, domain))
// 		}
// 		hostsBlock = append(hostsBlock, "      fallthrough")
// 		hostsBlock = append(hostsBlock, "   }")

// 		// Insert the new hosts block right after the .:53 { line.
// 		insertIndex := -1
// 		for i, line := range lines {
// 			if strings.HasPrefix(strings.TrimSpace(line), ".:53 {") {
// 				insertIndex = i + 1
// 				break
// 			}
// 		}
// 		if insertIndex == -1 {
// 			return fmt.Errorf("could not find .:53 block in Corefile")
// 		}

// 		// Insert the hosts block lines.
// 		newLines := append(lines[:insertIndex], append(hostsBlock, lines[insertIndex:]...)...)
// 		corefile = strings.Join(newLines, "\n")
// 	}

// 	// Update the Corefile field in the ConfigMap.
// 	configMap.Data["Corefile"] = corefile

// 	// Update the ConfigMap in the cluster.
// 	_, err = c.clientset.CoreV1().ConfigMaps(c.namespace).Update(ctx, configMap, metav1.UpdateOptions{})
// 	if err != nil {
// 		return fmt.Errorf("failed to update ConfigMap: %v", err)
// 	}

//		return nil
//	}
//
// AddDNSEntry adds a new DNS entry to the CoreDNS ConfigMap.
func (c *CoreDNSManager) AddDNSEntry(ctx context.Context, dnsName, ipAddress string) error {
	updatedData := map[string]string{
		ipAddress: dnsName,
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

func updateCorefileWithCustomHosts(corefile string, customHosts map[string]string) (string, error) {
	// Split into lines to do a simple block parse.
	lines := strings.Split(corefile, "\n")
	var resultLines []string

	inHostsBlock := false
	hostsBlockIndent := ""
	insertedCustom := false
	braceCount := 0

	// Process each line.
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inHostsBlock {
			// Look for the start of a hosts block.
			// (A proper parser would be more robust; here we simply check for a line like "hosts {")
			if strings.HasPrefix(trimmed, "hosts") && strings.Contains(trimmed, "{") {
				inHostsBlock = true
				hostsBlockIndent = line[:len(line)-len(strings.TrimLeft(line, " \t"))]
				resultLines = append(resultLines, line)
				// Initialize our brace counter (number of "{" minus "}")
				braceCount = strings.Count(line, "{") - strings.Count(line, "}")
				continue
			} else {
				resultLines = append(resultLines, line)
			}
		} else {
			// We are inside the hosts block.
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")
			// If we reached the closing brace for the hosts block...
			if braceCount == 0 {
				// If we have not yet inserted our custom entries, do so now.
				if !insertedCustom {
					customLines := generateCustomHostsLines(hostsBlockIndent+"    ", customHosts)
					resultLines = append(resultLines, customLines...)
					insertedCustom = true
				}
				resultLines = append(resultLines, line)
				inHostsBlock = false
				continue
			}

			// If we see a directive like "fallthrough", insert our custom entries immediately before it.
			if strings.HasPrefix(trimmed, "fallthrough") && !insertedCustom {
				customLines := generateCustomHostsLines(hostsBlockIndent+"    ", customHosts)
				resultLines = append(resultLines, customLines...)
				insertedCustom = true
			}

			// Check if this line looks like an IP mapping (i.e. first token is an IP).
			fields := strings.Fields(trimmed)
			if len(fields) > 1 {
				if net.ParseIP(fields[0]) != nil {
					// If the IP exists in our custom map, skip this line so that our custom entry wins.
					if _, exists := customHosts[fields[0]]; exists {
						continue
					}
				}
			}
			// Otherwise, preserve the original line.
			resultLines = append(resultLines, line)
		}
	}

	// If no hosts block was found (or custom entries were not inserted), append a new hosts block.
	if !insertedCustom {
		resultLines = append(resultLines, "hosts {")
		customLines := generateCustomHostsLines("    ", customHosts)
		resultLines = append(resultLines, customLines...)
		resultLines = append(resultLines, "}")
	}

	return strings.Join(resultLines, "\n"), nil
}

// generateCustomHostsLines returns the custom hosts entries as lines,
// sorted by IP (so that duplicate keys – if provided in order – yield the last one).
func generateCustomHostsLines(indent string, customHosts map[string]string) []string {
	ips := make([]string, 0, len(customHosts))
	for ip := range customHosts {
		ips = append(ips, ip)
	}
	sort.Strings(ips)
	var lines []string
	for _, ip := range ips {
		// Each line is "indent IP hostname"
		line := fmt.Sprintf("%s%s %s", indent, ip, customHosts[ip])
		lines = append(lines, line)
	}
	return lines
}
