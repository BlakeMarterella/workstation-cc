package dotfiles

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Entry maps an app-configs/ subdirectory to its install destination(s).
type Entry struct {
	Name string            `yaml:"name"`
	Dest map[string]string `yaml:"dest"` // keys: "default" or a GOOS (darwin/linux/windows)
}

// Manifest is the parsed app-configs/manifest.yaml.
type Manifest struct {
	Entries []Entry `yaml:"entries"`
}

// ParseManifest parses manifest.yaml bytes. An empty/whitespace body yields an
// empty manifest (no entries), which is valid.
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &m, nil
}

// DestFor resolves the destination for name under goos, expanding a leading ~ to
// home. It prefers an exact goos key, then "default". Returns ("", false) when
// neither is present.
func (m *Manifest) DestFor(name, goos, home string) (string, bool) {
	for _, e := range m.Entries {
		if e.Name != name {
			continue
		}
		raw, ok := e.Dest[goos]
		if !ok {
			raw, ok = e.Dest["default"]
		}
		if !ok {
			return "", false
		}
		return expandHome(raw, home), true
	}
	return "", false
}

func expandHome(p, home string) string {
	if p == "~" {
		return home
	}
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(home, p[2:])
	}
	return p
}

// LinkApps links each manifest entry's directory (appsRoot/<name>) to its
// resolved destination for goos. Entries with no destination for goos are
// recorded as ActionSkipped rather than failing.
func (l *Linker) LinkApps(appsRoot string, m *Manifest, goos string) ([]Result, error) {
	var results []Result
	for _, e := range m.Entries {
		src := filepath.Join(appsRoot, e.Name)
		dest, ok := m.DestFor(e.Name, goos, l.home)
		if !ok {
			results = append(results, Result{
				Path:   src,
				Action: ActionSkipped,
				Note:   fmt.Sprintf("no destination for %s", goos),
			})
			continue
		}
		res, err := l.LinkOne(src, dest)
		if err != nil {
			return results, err
		}
		results = append(results, res)
	}
	return results, nil
}
