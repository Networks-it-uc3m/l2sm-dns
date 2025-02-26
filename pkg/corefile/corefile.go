// Copyright 2025 Alejandro T. de Cock Buning
//
// Portions of this file are derived from the CoreDNS corefile-migration project:
//   https://github.com/coredns/corefile-migration/blob/master/migration/corefile/corefile.go
//
// Modifications have been made to the original file.
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

import (
	"fmt"
	"strings"

	"github.com/coredns/caddy/caddyfile"
)

const indent = 4

type Corefile struct {
	Servers []*Server
}

func New(s string) (*Corefile, error) {
	c := Corefile{}
	cc := caddyfile.NewDispenser("migration", strings.NewReader(s))
	depth := 0
	var cSvr *Server
	var cPlg *Plugin
	for cc.Next() {
		if cc.Val() == "{" {
			depth += 1
			continue
		} else if cc.Val() == "}" {
			depth -= 1
			continue
		}
		val := cc.Val()
		args := cc.RemainingArgs()
		switch depth {
		case 0:
			c.Servers = append(c.Servers,
				&Server{
					DomPorts: append([]string{val}, args...),
				})
			cSvr = c.Servers[len(c.Servers)-1]
		case 1:
			cSvr.Plugins = append(cSvr.Plugins,
				&Plugin{
					Name: val,
					Args: args,
				})
			cPlg = cSvr.Plugins[len(cSvr.Plugins)-1]
		case 2:
			cPlg.Options = append(cPlg.Options,
				&Option{
					Name: val,
					Args: args,
				})
		}
	}
	return &c, nil
}

func (c *Corefile) ToString() (out string) {
	strs := []string{}
	for _, s := range c.Servers {
		strs = append(strs, s.ToString())
	}
	return strings.Join(strs, "\n")
}

// escapeArgs returns the arguments list escaping and wrapping any argument containing whitespace in quotes
func escapeArgs(args []string) []string {
	var escapedArgs []string
	for _, a := range args {
		// if there is white space, wrap argument with quotes
		if len(strings.Fields(a)) > 1 {
			// escape quotes
			a = strings.Replace(a, "\"", "\\\"", -1)
			// wrap with quotes
			a = "\"" + a + "\""
		}
		escapedArgs = append(escapedArgs, a)
	}
	return escapedArgs
}

// FindServer returns the first *Server that exactly matches the given domPorts.
func (c *Corefile) GetServer(domPorts ...string) (*Server, bool) {
	for _, s := range c.Servers {
		if len(s.DomPorts) == len(domPorts) {
			match := true
			for i, dp := range s.DomPorts {
				if dp != domPorts[i] {
					match = false
					break
				}
			}
			if match {
				return s, true
			}
		}
	}
	return nil, false
}

// uniqueStrings helps deduplicate domain entries for a single IP.
func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, val := range in {
		if _, found := seen[val]; !found {
			seen[val] = struct{}{}
			out = append(out, val)
		}
	}
	return out
}

// removeStrings removes each element of removeList from base.
func removeStrings(base, removeList []string) []string {
	rmSet := make(map[string]struct{}, len(removeList))
	for _, rm := range removeList {
		rmSet[rm] = struct{}{}
	}
	var out []string
	for _, val := range base {
		if _, toRemove := rmSet[val]; !toRemove {
			out = append(out, val)
		}
	}
	return out
}

func (c *Corefile) AddServer(server Server) error {
	// Check if a server with the same DomPorts already exists.
	existing, ok := c.GetServer(server.DomPorts...)
	if ok {
		// Update existing server by merging its plugins with the new server's plugins.
		for _, newPlg := range server.Plugins {
			if existingPlg, found := existing.GetPlugin(newPlg.Name); found {
				// Here we choose to update the plugin's args and options.
				existingPlg.Args = newPlg.Args
				existingPlg.Options = newPlg.Options
			} else {
				// Plugin not present in the existing server, so add it.
				existing.Plugins = append(existing.Plugins, newPlg)
			}
		}
		return nil
	}
	fmt.Println(server.ToString())
	// If no matching server is found, add the new server.
	c.Servers = append(c.Servers, &server)
	return nil
}
