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

package corefile

import "strings"

type Server struct {
	DomPorts []string
	Plugins  []*Plugin
}

func (s *Server) ToString() (out string) {
	str := strings.Join(escapeArgs(s.DomPorts), " ")
	strs := []string{}
	for _, p := range s.Plugins {
		strs = append(strs, strings.Repeat(" ", indent)+p.ToString())
	}
	if len(strs) > 0 {
		str += " {\n" + strings.Join(strs, "\n") + "\n}\n"
	}
	return str
}

func (s *Server) FindMatch(def []*Server) (*Server, bool) {
NextServer:
	for _, sDef := range def {
		for i, dp := range sDef.DomPorts {
			if dp == "*" {
				continue
			}
			if dp == "***" {
				return sDef, true
			}
			if i >= len(s.DomPorts) || dp != s.DomPorts[i] {
				continue NextServer
			}
		}
		if len(sDef.DomPorts) != len(s.DomPorts) {
			continue
		}
		return sDef, true
	}
	return nil, false
}

// GetPlugin returns the first *Plugin in s.Plugins whose Name matches pluginName.
// If no match is found, it returns (nil, false).
func (s *Server) GetPlugin(pluginName string) (*Plugin, bool) {
	for _, p := range s.Plugins {
		if p.Name == pluginName {
			return p, true
		}
	}
	return nil, false
}
