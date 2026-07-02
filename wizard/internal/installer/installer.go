// Package installer turns a resolved set of package groups into a concrete,
// ordered install plan and executes it against a pkgmgr.Manager. Planning is
// kept separate from execution so it can be unit-tested and previewed
// (dry-run) without side effects.
package installer

import (
	"fmt"
	"io"

	"github.com/BlakeMarterella/workstation-cc/internal/packages"
	"github.com/BlakeMarterella/workstation-cc/internal/pkgmgr"
)

// Item is one planned package, resolved for the active manager.
type Item struct {
	Name string         // display name from the catalog
	Pkg  pkgmgr.Package // manager-specific package (valid only when Available)
	// Available reports whether the active manager has an identifier for this
	// package. Unavailable items are reported, not installed.
	Available bool
}

// BuildPlan expands the given groups (already resolved from a profile) into a
// deduplicated, ordered list of Items for the named manager.
func BuildPlan(cat *packages.Catalog, groups []string, manager string) ([]Item, error) {
	var items []Item
	seen := map[string]bool{}

	for _, gname := range groups {
		g, ok := cat.Group(gname)
		if !ok {
			return nil, fmt.Errorf("unknown package group %q", gname)
		}
		for _, p := range g.Packages {
			if seen[p.Name] {
				continue
			}
			seen[p.Name] = true

			pkg, ok := p.For(manager)
			items = append(items, Item{Name: p.Name, Pkg: pkg, Available: ok})
		}
	}
	return items, nil
}

// Result summarizes what Execute did (or, in dry-run, would do).
type Result struct {
	Installed   []string // installed now (or would install in dry-run)
	Skipped     []string // already present
	Unavailable []string // no mapping for the active manager
}

// Execute installs every available, missing item via mgr. Present items are
// skipped (idempotent) and unavailable items are reported. When dryRun is true
// no changes are made; planned installs are still reported. Progress is written
// to out.
func Execute(items []Item, mgr pkgmgr.Manager, dryRun bool, out io.Writer) (Result, error) {
	var res Result

	for _, it := range items {
		if !it.Available {
			res.Unavailable = append(res.Unavailable, it.Name)
			fmt.Fprintf(out, "  - %s: unavailable for %s, skipping\n", it.Name, mgr.Name())
			continue
		}

		installed, err := mgr.IsInstalled(it.Pkg.ID)
		if err != nil {
			return res, fmt.Errorf("checking %s: %w", it.Name, err)
		}
		if installed {
			res.Skipped = append(res.Skipped, it.Name)
			fmt.Fprintf(out, "  - %s: already installed\n", it.Name)
			continue
		}

		if dryRun {
			res.Installed = append(res.Installed, it.Name)
			fmt.Fprintf(out, "  - %s: would install (%s %s)\n", it.Name, mgr.Name(), it.Pkg.ID)
			continue
		}

		fmt.Fprintf(out, "  - %s: installing...\n", it.Name)
		if err := mgr.Install(it.Pkg); err != nil {
			return res, fmt.Errorf("installing %s: %w", it.Name, err)
		}
		res.Installed = append(res.Installed, it.Name)
	}

	return res, nil
}
