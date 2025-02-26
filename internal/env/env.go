// Copyright 2024 Universidad Carlos III de Madrid
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

package env

import (
	"os"
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetConfigMapNS() string {
	return getEnv("CONFIGMAP_NS", "default")
}

func GetConfigMapName() string {
	return getEnv("CONFIGMAP_NAME", "l2sm-coredns-config")
}

func GetServerPort() string {
	return getEnv("SERVER_PORT", "8081")
}

func GetInterDomainDomPort() string {
	return getEnv("INTER_DOMAIN_DOM_PORT", ".:53")
}
