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
	"log"
	"net"
	"path/filepath"

	"github.com/Networks-it-uc3m/l2sm-dns/api/v1/dns"
	corednsmanager "github.com/Networks-it-uc3m/l2sm-dns/pkg/coredns-manager"
	"google.golang.org/grpc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		// If in-cluster config is not available, try the local kubeconfig
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
		if err != nil {
			log.Fatalf("could not create config from either in-cluster or kubeconfig: %v", err)
		}
	}

	corednsmanager, err := corednsmanager.NewCoreDNSManager("default", "coredns", k8sConfig)

	dns.RegisterDnsServiceServer(grpcServer, &server{dns.UnimplementedDnsServiceServer{}, *corednsmanager})

	log.Printf("Server listening at %v", lis.Addr())

	// Start serving requests
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
