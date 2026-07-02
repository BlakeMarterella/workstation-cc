// Package cli wires up the Cobra command tree for the workstation worker.
package cli

import (
	"fmt"

	"github.com/BlakeMarterella/workstation-cc/internal/tui"
	"github.com/spf13/cobra"
)

// NewRootCmd builds the root command. version is injected from main so the
// build can stamp it via ldflags.
func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "workstation",
		Short: "Bootstrap and manage a workstation (packages, dotfiles, profiles)",
		Long: "workstation is the worker binary for workstation-cc. It installs " +
			"utilities, symlinks repo-owned dotfiles, and applies installation " +
			"profiles across macOS, Linux, and Windows.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: false,
		// With no subcommand on an interactive terminal, launch the TUI.
		// Otherwise (piped/headless) fall back to printing help.
		RunE: func(cmd *cobra.Command, _ []string) error {
			if tui.IsInteractive() {
				return runTUI(cmd.OutOrStdout())
			}
			return cmd.Help()
		},
	}

	// Render `--version` as a clean single line.
	root.SetVersionTemplate(fmt.Sprintf("workstation %s\n", version))

	root.AddCommand(newInstallCmd())

	return root
}
