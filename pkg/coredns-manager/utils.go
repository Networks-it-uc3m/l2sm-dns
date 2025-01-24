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
