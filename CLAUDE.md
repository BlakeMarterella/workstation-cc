# CLAUDE.md

## Purpose

This is a workstation setup tool — the first thing cloned on a new machine. It installs utilities (CLI and GUI apps), manages dotfiles, and provides small scripts symlinked into `$PATH`. It supports Linux, macOS, and Windows.

## Architecture

### Top-level structure (intended)

```
install.sh / install.ps1     # Entry points per platform (thin; set WORKSTATION_ROOT)
config.sh                    # User-facing config: WORKSTATION_ROOT, profile defaults
wizard/                      # Self-contained Go module (CLI, TUI, all install logic)
packages/                    # Declarative package lists (embedded at build time)
profiles.yaml                # Installation profiles
dotfiles/                    # Repo-owned dotfiles, mirrors $HOME (symlinked by wizard)
app-configs/                 # App-specific configs + manifest.yaml (symlinked by wizard)
docs/                        # Documentation and specs
```

### Installation profiles

Profiles control what gets installed. A `slim` profile installs only essentials; `full` adds GUI apps, language runtimes, fonts, etc. Profiles are composable — `full` should include `slim`. Each profile maps to a set of package groups and script subsets.

### Cross-platform approach

- The primary shell language is Bash for Linux/macOS. Windows uses PowerShell.
- OS detection should be early and centralized (single function/file, not scattered `if [[ $OSTYPE ]]` blocks throughout).
- Package manager abstraction: `brew` (macOS), `apt`/`dnf` (Linux), `winget`/`scoop` (Windows). The abstraction layer lives in `lib/` and callers shouldn't need to know which manager is active.
- The Go module lives in `wizard/`. Because `//go:embed` cannot reach parent
  directories, `wizard/tools/gendata` copies root `packages/` + `profiles.yaml`
  into `wizard/` at build time (run via `go generate ./...`, also a GoReleaser
  before-hook). These copies are gitignored.

### Dotfiles

Dotfiles are **owned by this repo**, not a separate repo. `dotfiles/` mirrors
`$HOME`; `app-configs/` holds app-specific configs whose destinations are declared
in `app-configs/manifest.yaml` (with per-OS support). The wizard symlinks these
into place from the local checkout (located via `WORKSTATION_ROOT`, which the
entry scripts default to the directory they live in). Existing real files are
backed up to `<name>.bak` before linking — never overwritten silently.

### Scripts on `$PATH`

Small utility scripts live in `scripts/`. The installer symlinks this directory (or its contents) into a location already on `$PATH` (e.g. `~/.local/bin`). Scripts should be self-contained and not depend on each other unless clearly documented.

## Key design constraints

- **Idempotent**: every installer step must be safe to run multiple times.
- **Fail-fast with clear messages**: if a prerequisite is missing or a step fails, stop and tell the user what happened and what to do.
- **No silent mutations**: don't overwrite existing dotfiles without asking or backing them up.
- **Profiles are additive**: `slim` ⊂ `full`. Never design a profile that requires another profile to be un-installed.
