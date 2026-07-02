package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/BlakeMarterella/workstation-cc/internal/dotfiles"
	"github.com/BlakeMarterella/workstation-cc/internal/installer"
	"github.com/BlakeMarterella/workstation-cc/internal/osdetect"
	"github.com/BlakeMarterella/workstation-cc/internal/packages"
	"github.com/BlakeMarterella/workstation-cc/internal/pkgmgr"
	"github.com/BlakeMarterella/workstation-cc/internal/profiles"
	"github.com/BlakeMarterella/workstation-cc/internal/ui"
	"github.com/spf13/cobra"
)

// defaultDotfilesRepo mirrors YADM_REPO from the original config.sh. It can be
// overridden by the WORKSTATION_YADM_REPO env var or the --dotfiles-repo flag.
const defaultDotfilesRepo = "https://github.com/BlakeMarterella/workstation-dotfiles"

type installOptions struct {
	profile      string
	dryRun       bool
	assumeYes    bool
	skipDotfiles bool
	dotfilesRepo string
}

func newInstallCmd() *cobra.Command {
	opts := installOptions{}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install packages for a profile and bootstrap dotfiles",
		Long: "Install resolves a profile into a set of packages, installs the " +
			"missing ones via the host package manager, and bootstraps the yadm " +
			"dotfiles repo. It is non-interactive and idempotent: use --dry-run to " +
			"preview, and --yes to apply changes.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInstall(cmd.OutOrStdout(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVar(&opts.profile, "profile", "slim", "installation profile to apply")
	f.BoolVar(&opts.dryRun, "dry-run", false, "preview actions without changing anything")
	f.BoolVar(&opts.assumeYes, "yes", false, "apply changes without prompting (required to mutate)")
	f.BoolVar(&opts.skipDotfiles, "skip-dotfiles", false, "do not bootstrap the yadm dotfiles repo")
	f.StringVar(&opts.dotfilesRepo, "dotfiles-repo", dotfilesRepoDefault(), "yadm dotfiles repository URL")

	return cmd
}

func dotfilesRepoDefault() string {
	if v := os.Getenv("WORKSTATION_YADM_REPO"); v != "" {
		return v
	}
	return defaultDotfilesRepo
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
		if err := bootstrapDotfiles(out, opts); err != nil {
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

func bootstrapDotfiles(out io.Writer, opts installOptions) error {
	fmt.Fprintln(out, ui.Header("Dotfiles (yadm)"))
	if opts.dryRun {
		fmt.Fprintf(out, "  - would clone %s\n", opts.dotfilesRepo)
		return nil
	}
	status, err := dotfiles.New(dotfiles.OSEnv{}).Clone(opts.dotfilesRepo)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "  - %s: %s\n", opts.dotfilesRepo, status)
	return nil
}
