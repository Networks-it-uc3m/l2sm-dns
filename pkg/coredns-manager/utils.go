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
	"fmt"
)

type DNSEntry struct {
	PodName string
	Network string
	Scope   string
}

func GenerateKey(dnsEntry DNSEntry) (string, error) {
	if dnsEntry.PodName == "" || dnsEntry.Network == "" || dnsEntry.Scope == "" {
		return "", fmt.Errorf("input entry has fields missing. All fields must be filled, received: %v", dnsEntry)
	}
	return fmt.Sprintf("%s.%s.%s.l2sm", dnsEntry.PodName, dnsEntry.Network, dnsEntry.Scope), nil
}
