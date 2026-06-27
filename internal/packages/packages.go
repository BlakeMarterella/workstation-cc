// Package packages loads the declarative package catalog. Package data lives
// in YAML files (one per category/group) that are embedded into the binary at
// build time, so the catalog is available with no external files at runtime.
package packages

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/BlakeMarterella/workstation-cc/internal/pkgmgr"
	"gopkg.in/yaml.v3"
)

// Package is one installable item with per-manager identifiers.
type Package struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Cask        bool              `yaml:"cask"`
	Managers    map[string]string `yaml:"managers"`
}

// For resolves the package to a pkgmgr.Package for the named manager. ok is
// false when the package has no identifier for that manager (unavailable).
func (p Package) For(manager string) (pkgmgr.Package, bool) {
	id, ok := p.Managers[manager]
	if !ok || id == "" {
		return pkgmgr.Package{}, false
	}
	return pkgmgr.Package{ID: id, Cask: p.Cask}, true
}

// Group is a named collection of packages (a category, e.g. "core", "dev").
type Group struct {
	Name        string    `yaml:"category"`
	Description string    `yaml:"description"`
	Packages    []Package `yaml:"packages"`
}

// Catalog is the full set of groups, keyed by name.
type Catalog struct {
	Groups map[string]Group
}

// Group returns a group by name.
func (c *Catalog) Group(name string) (Group, bool) {
	g, ok := c.Groups[name]
	return g, ok
}

// GroupNames returns the catalog's group names, sorted.
func (c *Catalog) GroupNames() []string {
	names := make([]string, 0, len(c.Groups))
	for n := range c.Groups {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// parseGroup decodes a single category YAML document.
func parseGroup(data []byte) (Group, error) {
	var g Group
	if err := yaml.Unmarshal(data, &g); err != nil {
		return Group{}, fmt.Errorf("parse group: %w", err)
	}
	if strings.TrimSpace(g.Name) == "" {
		return Group{}, fmt.Errorf("group is missing required 'category' field")
	}
	return g, nil
}

// LoadFS builds a Catalog from every *.yaml file in fsys.
func LoadFS(fsys fs.FS) (*Catalog, error) {
	cat := &Catalog{Groups: map[string]Group{}}

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read package data: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		g, err := parseGroup(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path.Base(e.Name()), err)
		}
		if _, dup := cat.Groups[g.Name]; dup {
			return nil, fmt.Errorf("duplicate group %q (in %s)", g.Name, e.Name())
		}
		cat.Groups[g.Name] = g
	}

	return cat, nil
}
