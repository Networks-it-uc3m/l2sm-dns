package corefile

import (
	"fmt"
	"strings"
)

type Plugin struct {
	Name    string
	Args    []string
	Options []*Option
}

// ListHostsEntries collects and returns a map of IP -> []domains from the hosts plugin options.
func (p *Plugin) ListHostsEntries() (map[string][]string, error) {
	if p.Name != "hosts" {
		return nil, fmt.Errorf("plugin %s is not 'hosts'", p.Name)
	}

	result := make(map[string][]string)
	for _, opt := range p.Options {
		// Each Option typically is:  <ip> <domain1> <domain2> ...
		ip := opt.Name
		if ip == "" {
			continue
		}
		// The rest of the Args are domain names
		result[ip] = append(result[ip], opt.Args...)
	}

	return result, nil
}

// ReplaceHostsEntries takes a map of ip -> []domains and replaces the plugin’s entire set of host entries.
func (p *Plugin) ReplaceHostsEntries(entries map[string][]string) error {
	if p.Name != "hosts" {
		return fmt.Errorf("plugin %s is not 'hosts'", p.Name)
	}

	var newOptions []*Option
	for ip, domains := range entries {
		newOptions = append(newOptions, &Option{
			Name: ip,
			Args: domains,
		})
	}
	p.Options = newOptions
	return nil
}

// AddHostsEntries merges the given map of ip -> []domains into the existing host entries.
func (p *Plugin) AddHostsEntries(entries map[string][]string) error {
	existing, err := p.ListHostsEntries()
	if err != nil {
		return err
	}

	for ip, newDomains := range entries {
		existing[ip] = append(existing[ip], newDomains...)
		// optionally deduplicate
		existing[ip] = uniqueStrings(existing[ip])
	}
	return p.ReplaceHostsEntries(existing)
}

// RemoveHostsEntries removes the specified domains from the specified IPs. If domains for an IP become empty, that IP is removed.
func (p *Plugin) RemoveHostsEntries(entries map[string][]string) error {
	existing, err := p.ListHostsEntries()
	if err != nil {
		return err
	}

	for ip, rmDomains := range entries {
		if _, found := existing[ip]; !found {
			continue
		}
		existing[ip] = removeStrings(existing[ip], rmDomains)
		if len(existing[ip]) == 0 {
			delete(existing, ip)
		}
	}

	return p.ReplaceHostsEntries(existing)
}

func (p *Plugin) ToString() (out string) {
	str := strings.Join(append([]string{p.Name}, escapeArgs(p.Args)...), " ")
	strs := []string{}
	for _, o := range p.Options {
		strs = append(strs, strings.Repeat(" ", indent*2)+o.ToString())
	}
	if len(strs) > 0 {
		str += " {\n" + strings.Join(strs, "\n") + "\n" + strings.Repeat(" ", indent*1) + "}"
	}
	return str
}

func (p *Plugin) FindMatch(def []*Plugin) (*Plugin, bool) {
NextPlugin:
	for _, pDef := range def {
		if pDef.Name != p.Name {
			continue
		}
		for i, arg := range pDef.Args {
			if arg == "*" {
				continue
			}
			if arg == "***" {
				return pDef, true
			}
			if i >= len(p.Args) || arg != p.Args[i] {
				continue NextPlugin
			}
		}
		if len(pDef.Args) != len(p.Args) {
			continue
		}
		return pDef, true
	}
	return nil, false
}
