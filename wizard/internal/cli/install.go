package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BlakeMarterella/workstation-cc/internal/dotfiles"
	"github.com/BlakeMarterella/workstation-cc/internal/installer"
	"github.com/BlakeMarterella/workstation-cc/internal/osdetect"
	"github.com/BlakeMarterella/workstation-cc/internal/packages"
	"github.com/BlakeMarterella/workstation-cc/internal/pkgmgr"
	"github.com/BlakeMarterella/workstation-cc/internal/profiles"
	"github.com/BlakeMarterella/workstation-cc/internal/ui"
	"github.com/spf13/cobra"
)

type installOptions struct {
	profile      string
	dryRun       bool
	assumeYes    bool
	skipDotfiles bool
	root         string // repo checkout to link dotfiles/ and app-configs/ from
}

func newInstallCmd() *cobra.Command {
	opts := installOptions{}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install packages for a profile and symlink dotfiles",
		Long: "Install resolves a profile into a set of packages, installs the " +
			"missing ones via the host package manager, and symlinks the repo's " +
			"dotfiles/ and app-configs/ into place. It is non-interactive and " +
			"idempotent: use --dry-run to preview, and --yes to apply changes.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInstall(cmd.OutOrStdout(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVar(&opts.profile, "profile", "slim", "installation profile to apply")
	f.BoolVar(&opts.dryRun, "dry-run", false, "preview actions without changing anything")
	f.BoolVar(&opts.assumeYes, "yes", false, "apply changes without prompting (required to mutate)")
	f.BoolVar(&opts.skipDotfiles, "skip-dotfiles", false, "do not symlink dotfiles/ or app-configs/")
	f.StringVar(&opts.root, "root", os.Getenv("WORKSTATION_ROOT"), "repo checkout to link dotfiles from")

	return cmd
}

func runInstall(out io.Writer, opts installOptions) error {
	// Fail fast rather than silently mutating the system.
	if !opts.dryRun && !opts.assumeYes {
		return fmt.Errorf("refusing to make changes without --yes (or pass --dry-run to preview)")
	}

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
	groups, err := set.Resolve(opts.profile)
	if err != nil {
		return err
	}

	items, err := installer.BuildPlan(cat, groups, mgr.Name())
	if err != nil {
		return err
	}

	mode := "applying"
	if opts.dryRun {
		mode = "dry-run"
	}
	fmt.Fprintln(out, ui.Header(fmt.Sprintf("Profile %q on %s/%s via %s (%s)",
		opts.profile, info.OS, info.Arch, mgr.Name(), mode)))

	res, err := installer.Execute(items, mgr, opts.dryRun, out)
	if err != nil {
		return err
	}

	if !opts.skipDotfiles {
		if err := linkDotfiles(out, opts, string(info.OS)); err != nil {
			return err
		}
	}

	fmt.Fprintln(out, ui.Header("Summary"))
	fmt.Fprintln(out, ui.Success.Render(fmt.Sprintf("  installed: %d", len(res.Installed))))
	fmt.Fprintln(out, ui.Faint.Render(fmt.Sprintf("  skipped (present): %d", len(res.Skipped))))
	if len(res.Unavailable) > 0 {
		fmt.Fprintln(out, ui.Warn.Render(fmt.Sprintf("  unavailable: %d (%v)", len(res.Unavailable), res.Unavailable)))
	}
	return nil
}

func linkDotfiles(out io.Writer, opts installOptions, goos string) error {
	if opts.root == "" {
		return fmt.Errorf("no checkout root: set WORKSTATION_ROOT or pass --root " +
			"(install.sh does this automatically)")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	l := dotfiles.NewLinker(dotfiles.OSFS{}, home, opts.dryRun)

	fmt.Fprintln(out, ui.Header("Dotfiles"))
	dfResults, err := l.LinkTree(filepath.Join(opts.root, "dotfiles"))
	if err != nil {
		return err
	}
	printLinkResults(out, dfResults)

	fmt.Fprintln(out, ui.Header("App configs"))
	manifestPath := filepath.Join(opts.root, "app-configs", "manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", manifestPath, err)
	}
	m, err := dotfiles.ParseManifest(data)
	if err != nil {
		return err
	}
	appResults, err := l.LinkApps(filepath.Join(opts.root, "app-configs"), m, goos)
	if err != nil {
		return err
	}
	printLinkResults(out, appResults)
	return nil
}

func printLinkResults(out io.Writer, results []dotfiles.Result) {
	for _, r := range results {
		line := fmt.Sprintf("  - %s: %s", r.Dest, r.Action)
		if r.Dest == "" {
			line = fmt.Sprintf("  - %s: %s", r.Path, r.Action)
		}
		if r.Note != "" {
			line += " (" + r.Note + ")"
		}
		fmt.Fprintln(out, line)
	}
}
