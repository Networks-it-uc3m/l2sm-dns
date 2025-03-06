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

package configmapmanager_test

import (
	"context"
	"testing"

	"github.com/Networks-it-uc3m/l2sm-dns/pkg/configmapmanager"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// ----------------------------------------------
// Helpers
// ----------------------------------------------
func createFakeScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	return scheme
}

func createConfigMap(name, namespace, corefileData string) *corev1.ConfigMap {
	data := make(map[string]string)
	if corefileData != "" {
		data["Corefile"] = corefileData
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}

func newDNSManager(t *testing.T, objs ...client.Object) configmapmanager.DNSManager {
	scheme := createFakeScheme()
	fclient := crfake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		Build()

	mgr, err := configmapmanager.NewDNSManager("test-namespace", "test-cm", nil, fclient)
	require.NoError(t, err)
	return mgr
}

// ----------------------------------------------
// GetConfigMap
// ----------------------------------------------
func TestGetConfigMapEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		objs           []client.Object
		expectErr      bool
		expectedErrMsg string
	}{
		{
			name:           "ConfigMap not found in cluster",
			objs:           []client.Object{}, // No ConfigMap at all
			expectErr:      true,
			expectedErrMsg: "not found",
		},
		{
			name: "ConfigMap exists",
			objs: []client.Object{
				createConfigMap("test-cm", "test-namespace", ".:53 {\n\thosts{}\n}"),
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc // pin
		t.Run(tc.name, func(t *testing.T) {
			mgr := newDNSManager(t, tc.objs...)
			cm, err := mgr.GetConfigMap(context.Background())
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
				require.Nil(t, cm)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cm)
			}
		})
	}
}

// ----------------------------------------------
// AddDNSEntryToConfigMap
// ----------------------------------------------
func TestAddDNSEntryToConfigMapEdgeCases(t *testing.T) {
	validCorefile := `.:53 {
  hosts {
  }
}`

	tests := []struct {
		name           string
		corefileData   string
		updatedData    map[string]string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			name:           "ConfigMap missing 'Corefile' key",
			corefileData:   "", // no Corefile key at all
			updatedData:    map[string]string{"1.2.3.4": "domain.com"},
			expectErr:      true,
			expectedErrMsg: "corefile not found in ConfigMap data",
		},
		{
			name:           "Invalid IP address in updatedData",
			corefileData:   validCorefile,
			updatedData:    map[string]string{"NOT_AN_IP": "domain.com"},
			expectErr:      true,
			expectedErrMsg: "invalid IP address",
		},
		{
			name: "Missing 'hosts' plugin in the server block",
			corefileData: `.:53 {
			  forward . /etc/resolv.conf
			}`,
			updatedData:    map[string]string{"1.2.3.4": "domain.com"},
			expectErr:      true,
			expectedErrMsg: "could not find 'hosts' plugin",
		},
		{
			name: "Missing server block for inter-domain port (env.GetInterDomainDomPort() not found)",
			corefileData: `example.org:53 {
			  hosts {
			  }
			}`,
			updatedData:    map[string]string{"1.2.3.4": "domain.com"},
			expectErr:      true,
			expectedErrMsg: "could not find inter-domain port",
		},
		{
			name:         "Successful add with valid data",
			corefileData: validCorefile,
			updatedData:  map[string]string{"1.2.3.4": "domain.com"},
			expectErr:    false,
		},
	}

	for _, tc := range tests {
		tc := tc // pin
		t.Run(tc.name, func(t *testing.T) {
			cm := createConfigMap("test-cm", "test-namespace", tc.corefileData)
			mgr := newDNSManager(t, cm)

			err := mgr.AddDNSEntryToConfigMap(context.Background(), tc.updatedData)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
				// optionally retrieve and confirm it contains the updated data
			}
		})
	}
}

// ----------------------------------------------
// RemoveDNSRecords
// ----------------------------------------------
func TestRemoveDNSRecordsEdgeCases(t *testing.T) {
	validCorefile := `.:53 {
  hosts {
    1.2.3.4 domain.com
    2.3.4.5 another.com
  }
}`

	tests := []struct {
		name           string
		corefileData   string
		removals       map[string][]string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			name:           "No 'Corefile' key in ConfigMap",
			corefileData:   "",
			removals:       map[string][]string{"1.2.3.4": {"domain.com"}},
			expectErr:      true,
			expectedErrMsg: "corefile not found in ConfigMap data",
		},
		{
			name: "No server block for inter-domain port",
			corefileData: `example.org:53 {
			  hosts {
			    1.2.3.4 domain.com
			  }
			}`,
			removals:       map[string][]string{"1.2.3.4": {"domain.com"}},
			expectErr:      true,
			expectedErrMsg: "could not find inter-domain server",
		},
		{
			name: "Missing 'hosts' plugin in valid server block",
			corefileData: `.:53 {
			  forward . /etc/resolv.conf
			}`,
			removals:       map[string][]string{"1.2.3.4": {"domain.com"}},
			expectErr:      true,
			expectedErrMsg: "could not find 'hosts' plugin",
		},
		{
			name:           "Invalid IP in removals",
			corefileData:   validCorefile,
			removals:       map[string][]string{"NOT_AN_IP": {"domain.com"}},
			expectErr:      true,
			expectedErrMsg: "invalid IP address in removals",
		},
		{
			name:         "Successful removal",
			corefileData: validCorefile,
			removals:     map[string][]string{"1.2.3.4": {"domain.com"}},
			expectErr:    false,
		},
	}

	for _, tc := range tests {
		tc := tc // pin
		t.Run(tc.name, func(t *testing.T) {
			cm := createConfigMap("test-cm", "test-namespace", tc.corefileData)
			mgr := newDNSManager(t, cm)

			err := mgr.RemoveDNSRecords(context.Background(), tc.removals)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
				// Optionally re-check the content of the CM to ensure the removal was successful
			}
		})
	}
}

// ----------------------------------------------
// ListDNSRecords
// ----------------------------------------------
func TestListDNSRecordsEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		corefileData   string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			name:           "No 'Corefile' key in ConfigMap",
			corefileData:   "",
			expectErr:      true,
			expectedErrMsg: "corefile not found in ConfigMap data",
		},
		{
			name: "Missing inter-domain server block",
			corefileData: `example.org:53 {
				hosts {
				  1.2.3.4 domain.com
				}
			}`,
			expectErr:      true,
			expectedErrMsg: "could not find inter-domain server",
		},
		{
			name: "Missing 'hosts' plugin in the correct server block",
			corefileData: `.:53 {
				forward . /etc/resolv.conf
			}`,
			expectErr:      true,
			expectedErrMsg: "could not find 'hosts' plugin",
		},
		{
			name: "Successful listing",
			corefileData: `.:53 {
				hosts {
				  1.2.3.4 domain.com
				  2.3.4.5 another.com
				}
			}`,
			expectErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc // pin
		t.Run(tc.name, func(t *testing.T) {
			cm := createConfigMap("test-cm", "test-namespace", tc.corefileData)
			mgr := newDNSManager(t, cm)

			records, err := mgr.ListDNSRecords(context.Background())
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
				require.Nil(t, records)
			} else {
				require.NoError(t, err)
				require.NotNil(t, records)
				// Optionally verify the content in 'records'
			}
		})
	}
}

// ----------------------------------------------
// AddDNSEntry
// ----------------------------------------------
func TestAddDNSEntryEdgeCases(t *testing.T) {
	validCorefile := `.:53 {
  hosts {
  }
}`

	tests := []struct {
		name           string
		corefileData   string
		dnsName        string
		ipAddress      string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			name:           "Invalid IP address",
			corefileData:   validCorefile,
			dnsName:        "domain.com",
			ipAddress:      "NOT_AN_IP",
			expectErr:      true,
			expectedErrMsg: "invalid IP address",
		},
		{
			name:           "No 'Corefile' in data",
			corefileData:   "",
			dnsName:        "domain.com",
			ipAddress:      "1.2.3.4",
			expectErr:      true,
			expectedErrMsg: "corefile not found",
		},
		{
			name:           "Missing server block for inter-domain port",
			corefileData:   `example.org:53 { hosts {} }`,
			dnsName:        "domain.com",
			ipAddress:      "1.2.3.4",
			expectErr:      true,
			expectedErrMsg: "could not find inter-domain port",
		},
		{
			name: "Missing 'hosts' plugin in correct server block",
			corefileData: `.:53 {
			  forward . /etc/resolv.conf
			}`,
			dnsName:        "domain.com",
			ipAddress:      "1.2.3.4",
			expectErr:      true,
			expectedErrMsg: "could not find 'hosts' plugin",
		},
		{
			name:         "Successful add",
			corefileData: validCorefile,
			dnsName:      "domain.com",
			ipAddress:    "1.2.3.4",
			expectErr:    false,
		},
	}

	for _, tc := range tests {
		tc := tc // pin
		t.Run(tc.name, func(t *testing.T) {
			cm := createConfigMap("test-cm", "test-namespace", tc.corefileData)
			mgr := newDNSManager(t, cm)

			err := mgr.AddDNSEntry(context.Background(), tc.dnsName, tc.ipAddress)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ----------------------------------------------
// RemoveDNSEntry
// ----------------------------------------------
func TestRemoveDNSEntryEdgeCases(t *testing.T) {
	validCorefile := `.:53 {
  hosts {
    1.2.3.4 domain.com
  }
}`

	tests := []struct {
		name           string
		corefileData   string
		key            string
		ipAddress      string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			name:           "No 'Corefile' key in ConfigMap",
			corefileData:   "",
			key:            "domain.com",
			ipAddress:      "1.2.3.4",
			expectErr:      true,
			expectedErrMsg: "corefile not found",
		},
		{
			name: "Missing server block for inter-domain port",
			corefileData: `example.org:53 {
			  hosts {
			    1.2.3.4 domain.com
			  }
			}`,
			key:            "domain.com",
			ipAddress:      "1.2.3.4",
			expectErr:      true,
			expectedErrMsg: "could not find inter-domain port",
		},
		{
			name: "No 'hosts' plugin in that server block",
			corefileData: `.:53 {
			  forward . /etc/resolv.conf
			}`,
			key:            "domain.com",
			ipAddress:      "1.2.3.4",
			expectErr:      true,
			expectedErrMsg: "could not find 'hosts' plugin",
		},
		{
			name:         "Successful removal",
			corefileData: validCorefile,
			key:          "domain.com",
			ipAddress:    "1.2.3.4",
			expectErr:    false,
		},
	}

	for _, tc := range tests {
		tc := tc // pin
		t.Run(tc.name, func(t *testing.T) {
			cm := createConfigMap("test-cm", "test-namespace", tc.corefileData)
			mgr := newDNSManager(t, cm)

			err := mgr.RemoveDNSEntry(context.Background(), tc.key, tc.ipAddress)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
				// optionally get the configmap again and ensure the entry is gone
			}
		})
	}
}

// ----------------------------------------------
// AddServerToConfigMap
// ----------------------------------------------
func TestAddServerToConfigMapEdgeCases(t *testing.T) {
	validCorefile := `.:53 {
  hosts {
  }
}`

	tests := []struct {
		name           string
		corefileData   string
		domainName     string
		serverDomain   string
		serverPort     string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			name:           "No 'Corefile' key in ConfigMap",
			corefileData:   "",
			domainName:     "example.org",
			serverDomain:   "server.domain",
			serverPort:     "53",
			expectErr:      true,
			expectedErrMsg: "corefile not found in ConfigMap data",
		},
		{
			name:         "Add a new server block successfully",
			corefileData: validCorefile,
			domainName:   "mydomain.org:8053",
			serverDomain: "server.domain",
			serverPort:   "53",
			expectErr:    false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cm := createConfigMap("test-cm", "test-namespace", tc.corefileData)
			mgr := newDNSManager(t, cm)

			err := mgr.AddServerToConfigMap(context.Background(), tc.domainName, tc.serverDomain, tc.serverPort)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
				// Optionally confirm new server block is in the Corefile
			}
		})
	}
}
