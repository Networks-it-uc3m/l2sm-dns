package corefile

import "strings"

type Option struct {
	Name string
	Args []string
}

func (o *Option) ToString() (out string) {
	str := strings.Join(append([]string{o.Name}, escapeArgs(o.Args)...), " ")
	return str
}

func (o *Option) FindMatch(def []*Option) (*Option, bool) {
NextOption:
	for _, oDef := range def {
		if oDef.Name != o.Name {
			continue
		}
		for i, arg := range oDef.Args {
			if arg == "*" {
				continue
			}
			if arg == "***" {
				return oDef, true
			}
			if i >= len(o.Args) || arg != o.Args[i] {
				continue NextOption
			}
		}
		if len(oDef.Args) != len(o.Args) {
			continue
		}
		return oDef, true
	}
	return nil, false
}
