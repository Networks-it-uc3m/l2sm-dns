// Copyright 2025 Alejandro de Cock Buning
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

package main

import (
	"context"
	"fmt"

	"github.com/Networks-it-uc3m/l2sm-dns/api/v1/dns"
	corednsmanager "github.com/Networks-it-uc3m/l2sm-dns/pkg/coredns-manager"
)

type server struct {
	dns.UnimplementedDnsServiceServer
	corednsmanager.CoreDNSManager
}

// CreateNetwork calls a method from mdclient to create a network
func (s *server) AddEntry(ctx context.Context, req *dns.AddEntryRequest) (*dns.AddEntryResponse, error) {

	dnsEntry := corednsmanager.DNSEntry{PodName: req.Entry.GetPodName(), Network: req.Entry.GetNetwork(), Scope: req.Entry.GetScope()}

	entryKey, err := corednsmanager.GenerateKey(dnsEntry)

	if err != nil {
		return &dns.AddEntryResponse{}, fmt.Errorf("could not generate entry key. err: %v", err)
	}

	err = s.CoreDNSManager.AddDNSEntry(context.TODO(), entryKey, req.GetEntry().GetIpAddress())

	if err != nil {
		return &dns.AddEntryResponse{}, fmt.Errorf("could not create entry. err: %v", err)
	}

	return &dns.AddEntryResponse{}, nil

}
