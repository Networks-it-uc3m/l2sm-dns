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
	"os"

	"gopkg.in/yaml.v2"
)

// Config holds the configuration parameters for the DNS client.
type Config struct {
	ServerAddress string          `yaml:"serverAddress"`
	DNS           DNSEntryConfig  `yaml:"dns"`
	Server        DNSServerConfig `yaml:"server"`
}

// DNSEntryConfig holds the default DNS entry parameters.
type DNSEntryConfig struct {
	PodName   string `yaml:"podName"`
	IpAddress string `yaml:"ipAddress"`
	Network   string `yaml:"network"`
	Scope     string `yaml:"scope"`
}

type DNSServerConfig struct {
	DomPort      string `yaml:"domPort"`
	ServerDomain string `yaml:"serverDomain"`
	ServerPort   string `yaml:"serverPort"`
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
