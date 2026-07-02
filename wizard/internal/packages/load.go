package packages

import (
	"fmt"
	"io/fs"

	assets "github.com/BlakeMarterella/workstation-cc"
)

// Load builds the Catalog from the binary's embedded package data.
func Load() (*Catalog, error) {
	sub, err := fs.Sub(assets.PackagesFS, "packages")
	if err != nil {
		return nil, fmt.Errorf("open embedded package data: %w", err)
	}
	return LoadFS(sub)
}
