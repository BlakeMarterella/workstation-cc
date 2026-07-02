// Package dotfiles bootstraps a yadm-managed dotfiles repository. It mirrors
// the original install.sh behavior (install yadm, then `yadm clone <repo>`) but
// is idempotent: an already-cloned repo is left untouched.
package dotfiles

import (
	"fmt"
	"os"
	"os/exec"
)

// Status reports what Clone did.
type Status int

const (
	// StatusCloned means the repository was freshly cloned.
	StatusCloned Status = iota
	// StatusAlreadyPresent means a yadm repo already existed and was left alone.
	StatusAlreadyPresent
)

func (s Status) String() string {
	switch s {
	case StatusCloned:
		return "cloned"
	case StatusAlreadyPresent:
		return "already present"
	default:
		return "unknown"
	}
}

// Env abstracts the commands dotfiles needs, so the bootstrap logic is testable
// without invoking the real yadm binary.
type Env interface {
	// HasCommand reports whether a command is available on PATH.
	HasCommand(name string) bool
	// RunOK runs a command silently and reports whether it exited zero.
	RunOK(name string, args ...string) bool
	// Run runs a command with stdio attached to the parent.
	Run(name string, args ...string) error
}

// Bootstrapper performs the dotfiles bootstrap against an Env.
type Bootstrapper struct {
	env Env
}

// New returns a Bootstrapper using env.
func New(env Env) *Bootstrapper {
	return &Bootstrapper{env: env}
}

// Clone ensures the yadm dotfiles repo is present. It is idempotent: if a repo
// is already initialized it returns StatusAlreadyPresent without touching it
// (no silent mutation of existing dotfiles).
func (b *Bootstrapper) Clone(repo string) (Status, error) {
	if repo == "" {
		return 0, fmt.Errorf("no dotfiles repo configured")
	}
	if !b.env.HasCommand("yadm") {
		return 0, fmt.Errorf("yadm is not installed; install it before bootstrapping dotfiles")
	}
	if b.env.RunOK("yadm", "status") {
		return StatusAlreadyPresent, nil
	}
	if err := b.env.Run("yadm", "clone", repo); err != nil {
		return 0, fmt.Errorf("yadm clone %s: %w", repo, err)
	}
	return StatusCloned, nil
}

// OSEnv is the production Env backed by os/exec.
type OSEnv struct{}

// HasCommand implements Env.
func (OSEnv) HasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// RunOK implements Env.
func (OSEnv) RunOK(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	return cmd.Run() == nil
}

// Run implements Env.
func (OSEnv) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
