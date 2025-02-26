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
