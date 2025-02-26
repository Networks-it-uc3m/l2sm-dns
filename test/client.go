// ======================
// File: client.go
// ======================
//
// This client connects to a DNS gRPC server and sends an AddEntry request.
// Usage example:
//   go run client.go --test-add-entry --config ./config.yaml --pod my-pod --ip 10.0.0.2 --network custom-network --scope local
//
// Copyright 2024
// Licensed under the Apache License, Version 2.0
// (see the LICENSE file for details).

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/Networks-it-uc3m/l2sm-dns/api/v1/dns"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Command-line flags.
	testAddEntry := flag.Bool("test-add-entry", false, "Simulate adding a DNS entry")
	testAddServer := flag.Bool("test-add-server", false, "Simulate adding a server")

	configPath := flag.String("config", "./config.yaml", "Path to YAML config file")
	// Allow overriding default DNS entry parameters from config.
	podName := flag.String("pod", "", "Pod name for the DNS entry")
	ipAddress := flag.String("ip", "", "IP address for the DNS entry")
	network := flag.String("network", "", "Network for the DNS entry")
	scope := flag.String("scope", "", "Scope for the DNS entry (default: global)")

	flag.Parse()

	// Load configuration from file.
	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override config values if flags are provided.
	if *podName != "" {
		cfg.DNS.PodName = *podName
	}
	if *ipAddress != "" {
		cfg.DNS.IpAddress = *ipAddress
	}
	if *network != "" {
		cfg.DNS.Network = *network
	}
	if *scope != "" {
		cfg.DNS.Scope = *scope
	}
	if cfg.DNS.Scope == "" {
		cfg.DNS.Scope = "global"
	}

	// Create a gRPC connection.
	conn, err := grpc.NewClient(cfg.ServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server at %s: %v", cfg.ServerAddress, err)
	}
	defer conn.Close()

	// Create a DNS service client.
	client := dns.NewDnsServiceClient(conn)

	// If the test-add-entry flag is provided, build and send the AddEntry request.
	if *testAddEntry {
		fmt.Println("Sending AddEntry request...")
		req := &dns.AddEntryRequest{
			Entry: &dns.DNSEntry{
				PodName:   cfg.DNS.PodName,
				IpAddress: cfg.DNS.IpAddress,
				Network:   cfg.DNS.Network,
				Scope:     cfg.DNS.Scope,
			},
		}
		// Wrap the call in a context with timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := client.AddEntry(ctx, req)
		if err != nil {
			log.Fatalf("Failed to add DNS entry: %v", err)
		}
		fmt.Printf("AddEntry response: %s\n", resp.GetMessage())
	}
	if *testAddServer {
		fmt.Println("Sending AddServer request...")
		req := &dns.AddServerRequest{
			Server: &dns.Server{
				DomPort:      cfg.Server.DomPort,
				ServerDomain: cfg.Server.ServerDomain,
				ServerPort:   cfg.Server.ServerPort,
			},
		}
		// Wrap the call in a context with timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := client.AddServer(ctx, req)
		if err != nil {
			log.Fatalf("Failed to add DNS server: %v", err)
		}
		fmt.Printf("AddEntry response: %s\n", resp.GetMessage())
	}
}
