package pkgmgr

import "fmt"

// Brew is the Homebrew Manager for macOS (and Linuxbrew).
type Brew struct {
	run Runner
}

// NewBrew returns a Brew manager that executes commands via run.
func NewBrew(run Runner) *Brew {
	return &Brew{run: run}
}

// Name implements Manager.
func (b *Brew) Name() string { return "brew" }

// IsInstalled reports whether a formula or cask is present. A non-zero exit
// from `brew list` means "not installed", which is not an error.
func (b *Brew) IsInstalled(id string) (bool, error) {
	if err := b.run.RunSilent("brew", "list", "--versions", id); err != nil {
		return false, nil
	}
	return true, nil
}

// Install runs `brew install [--cask] <id>`.
func (b *Brew) Install(p Package) error {
	args := []string{"install"}
	if p.Cask {
		args = append(args, "--cask")
	}
	args = append(args, p.ID)
	if err := b.run.Run("brew", args...); err != nil {
		return fmt.Errorf("brew install %s: %w", p.ID, err)
	}
	return nil
}
