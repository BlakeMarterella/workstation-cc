// Package profiles loads installation profiles and resolves them into a flat,
// deduplicated list of package groups. Profiles are additive and composable:
// a profile may `include` others, and `full` includes `slim` (never the
// reverse).
package profiles

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// Profile is a named selection of package groups, optionally composed from
// other profiles via Includes.
type Profile struct {
	Description string   `yaml:"description"`
	Includes    []string `yaml:"includes"`
	Groups      []string `yaml:"groups"`
}

// Set is the collection of all defined profiles.
type Set struct {
	Profiles map[string]Profile `yaml:"profiles"`
}

// Names returns the defined profile names, sorted.
func (s *Set) Names() []string {
	names := make([]string, 0, len(s.Profiles))
	for n := range s.Profiles {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Resolve returns the flattened, deduplicated list of package groups for a
// profile, expanding includes depth-first. Included groups precede the
// profile's own groups. Returns an error for unknown profiles/includes or
// include cycles.
func (s *Set) Resolve(name string) ([]string, error) {
	var out []string
	seen := map[string]bool{}     // group dedupe
	visiting := map[string]bool{} // cycle detection along current path

	var walk func(string) error
	walk = func(n string) error {
		p, ok := s.Profiles[n]
		if !ok {
			return fmt.Errorf("unknown profile %q", n)
		}
		if visiting[n] {
			return fmt.Errorf("include cycle detected at profile %q", n)
		}
		visiting[n] = true

		for _, inc := range p.Includes {
			if err := walk(inc); err != nil {
				return err
			}
		}
		for _, g := range p.Groups {
			if !seen[g] {
				seen[g] = true
				out = append(out, g)
			}
		}

		visiting[n] = false
		return nil
	}

	if err := walk(name); err != nil {
		return nil, err
	}
	return out, nil
}

// parseSet decodes a profiles YAML document.
func parseSet(data []byte) (*Set, error) {
	var s Set
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse profiles: %w", err)
	}
	if len(s.Profiles) == 0 {
		return nil, fmt.Errorf("no profiles defined")
	}
	return &s, nil
}
