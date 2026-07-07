// Command workstation is the worker binary for workstation-cc.
//
// It performs the real bootstrap work: OS detection, package-manager
// abstraction, profile resolution, package installation, and dotfile
// symlinking. The thin install.sh/preflight scripts download and exec this
// binary.
package main

import (
	"os"

	"github.com/BlakeMarterella/workstation-cc/internal/cli"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := cli.NewRootCmd(version).Execute(); err != nil {
		os.Exit(1)
	}
}
