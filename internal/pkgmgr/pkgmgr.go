// Package pkgmgr abstracts the host package manager (brew, apt, dnf, winget)
// behind a single Manager interface. Callers select a Manager once (based on
// osdetect) and never branch on the underlying tool again.
package pkgmgr

import (
	"os"
	"os/exec"
)

// Package is a single installable unit, already resolved to the active
// manager's identifier.
type Package struct {
	// ID is the manager-specific package name (e.g. "ripgrep", "fd-find").
	ID string
	// Cask marks a GUI application installed via `brew install --cask`.
	// Ignored by managers that have no cask concept.
	Cask bool
}

// Manager installs packages with a specific underlying tool.
type Manager interface {
	// Name returns the manager's identifier (e.g. "brew").
	Name() string
	// IsInstalled reports whether a package is already present.
	IsInstalled(id string) (bool, error)
	// Install installs a package. Implementations must be idempotent enough
	// that re-installing an existing package is harmless.
	Install(p Package) error
}

// Runner executes external commands. It exists so managers can be tested
// without spawning real processes.
type Runner interface {
	// Run executes a command with stdio attached so the user sees progress.
	Run(name string, args ...string) error
	// RunSilent executes a command discarding its output. Used for queries
	// (e.g. "is this installed?") that should not clutter the terminal.
	RunSilent(name string, args ...string) error
}

// ExecRunner is the production Runner. It streams child stdio to the parent so
// package-manager progress is visible to the user.
type ExecRunner struct{}

// Run implements Runner using os/exec, streaming child stdio to the parent.
func (ExecRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunSilent implements Runner, discarding the command's output.
func (ExecRunner) RunSilent(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}
