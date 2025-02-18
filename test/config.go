package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config holds the configuration parameters for the DNS client.
type Config struct {
	ServerAddress string         `yaml:"serverAddress"`
	DNS           DNSEntryConfig `yaml:"dns"`
}

// DNSEntryConfig holds the default DNS entry parameters.
type DNSEntryConfig struct {
	PodName   string `yaml:"podName"`
	IpAddress string `yaml:"ipAddress"`
	Network   string `yaml:"network"`
	Scope     string `yaml:"scope"`
}

// LoadConfig reads a YAML configuration file and unmarshals it into a Config struct.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
