package pkgmgr

import "fmt"

// stub is a placeholder Manager for package managers that are recognized but
// not yet implemented. It keeps callers OS-agnostic today and gives a clear,
// fail-fast error instead of a silent no-op.
type stub struct {
	name string
}

func (s stub) Name() string { return s.name }

func (s stub) IsInstalled(string) (bool, error) {
	return false, fmt.Errorf("%s support not yet implemented", s.name)
}

func (s stub) Install(Package) error {
	return fmt.Errorf("%s support not yet implemented", s.name)
}

// NewApt returns a not-yet-implemented apt manager (Debian/Ubuntu).
func NewApt(Runner) Manager { return stub{name: "apt"} }

// NewDnf returns a not-yet-implemented dnf manager (Fedora/RHEL).
func NewDnf(Runner) Manager { return stub{name: "dnf"} }

// NewWinget returns a not-yet-implemented winget manager (Windows).
func NewWinget(Runner) Manager { return stub{name: "winget"} }
