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
