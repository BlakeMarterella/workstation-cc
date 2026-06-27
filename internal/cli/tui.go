package cli

import (
	"fmt"
	"io"

	"github.com/BlakeMarterella/workstation-cc/internal/osdetect"
	"github.com/BlakeMarterella/workstation-cc/internal/packages"
	"github.com/BlakeMarterella/workstation-cc/internal/pkgmgr"
	"github.com/BlakeMarterella/workstation-cc/internal/profiles"
	"github.com/BlakeMarterella/workstation-cc/internal/tui"
)

// runTUI launches the interactive selector, then applies the chosen profile via
// the same install path used non-interactively. A TUI confirmation counts as
// --yes; cancelling makes no changes.
func runTUI(out io.Writer) error {
	info, err := osdetect.Detect()
	if err != nil {
		return err
	}
	mgr, err := pkgmgr.For(info, pkgmgr.ExecRunner{})
	if err != nil {
		return err
	}
	cat, err := packages.Load()
	if err != nil {
		return err
	}
	set, err := profiles.Load()
	if err != nil {
		return err
	}

	res, err := tui.Run(cat, set, mgr.Name())
	if err != nil {
		return err
	}
	if !res.Confirmed {
		fmt.Fprintln(out, "Cancelled — no changes made.")
		return nil
	}

	return runInstall(out, installOptions{
		profile:      res.Profile,
		assumeYes:    true,
		dotfilesRepo: dotfilesRepoDefault(),
	})
}
