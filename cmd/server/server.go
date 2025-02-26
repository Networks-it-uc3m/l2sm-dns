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

package main

import (
	"context"
	"fmt"

	"github.com/Networks-it-uc3m/l2sm-dns/api/v1/dns"
	configmapmanager "github.com/Networks-it-uc3m/l2sm-dns/pkg/configmapmanager"
)

type server struct {
	dns.UnimplementedDnsServiceServer
	configmapmanager.CoreDNSManager
}

// CreateNetwork calls a method from mdclient to create a network
func (s *server) AddEntry(ctx context.Context, req *dns.AddEntryRequest) (*dns.AddEntryResponse, error) {

	dnsEntry := configmapmanager.DNSEntry{PodName: req.Entry.GetPodName(), Network: req.Entry.GetNetwork(), Scope: req.Entry.GetScope()}

	entryKey, err := configmapmanager.GenerateKey(dnsEntry)

	if err != nil {
		return &dns.AddEntryResponse{}, fmt.Errorf("could not generate entry key. err: %v", err)
	}

	err = s.CoreDNSManager.AddDNSEntry(context.TODO(), entryKey, req.GetEntry().GetIpAddress())

	if err != nil {
		return &dns.AddEntryResponse{}, fmt.Errorf("could not create entry. err: %v", err)
	}

	return &dns.AddEntryResponse{}, nil

}

func (s *server) AddServer(ctx context.Context, req *dns.AddServerRequest) (*dns.AddServerResponse, error) {

	err := s.CoreDNSManager.AddServerToConfigMap(ctx, req.Server.GetDomPort(), req.Server.GetServerDomain(), req.Server.GetServerPort())

	if err != nil {
		return &dns.AddServerResponse{}, fmt.Errorf("could not create server. err: %v", err)

	}
	return &dns.AddServerResponse{}, nil

}
