// Package assets embeds declarative data files (package catalog, profiles) into
// the binary so the worker needs no external files at runtime.
package assets

import "embed"

// PackagesFS holds the declarative package catalog (packages/*.yaml).
//
//go:embed packages/*.yaml
var PackagesFS embed.FS

// ProfilesFS holds the profile definitions (profiles.yaml).
//
//go:embed profiles.yaml
var ProfilesFS embed.FS
