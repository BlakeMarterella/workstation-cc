package pkgmgr

import (
	"fmt"

	"github.com/BlakeMarterella/workstation-cc/internal/osdetect"
)

// For selects the appropriate Manager for the detected platform. This is the
// only place that maps OS/distro to a concrete package manager; callers
// downstream work purely against the Manager interface.
func For(info osdetect.Info, run Runner) (Manager, error) {
	switch info.OS {
	case osdetect.Darwin:
		return NewBrew(run), nil
	case osdetect.Linux:
		switch info.LinuxID {
		case "debian", "ubuntu", "linuxmint", "pop":
			return NewApt(run), nil
		case "fedora", "rhel", "centos", "rocky", "almalinux":
			return NewDnf(run), nil
		default:
			return nil, fmt.Errorf("no known package manager for Linux distro %q", info.LinuxID)
		}
	case osdetect.Windows:
		return NewWinget(run), nil
	default:
		return nil, fmt.Errorf("no package manager for OS %q", info.OS)
	}
}
