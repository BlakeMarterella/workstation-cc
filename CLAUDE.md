# CLAUDE.md

## Purpose

This is a workstation setup tool — the first thing cloned on a new machine. It installs utilities (CLI and GUI apps), manages dotfiles, and provides small scripts symlinked into `$PATH`. It supports Linux, macOS, and Windows.

## Architecture

### Top-level structure (intended)

```
install.sh / install.ps1     # Entry points per platform (shell/PowerShell)
config.sh                    # User-facing config: YADM_REPO, profile defaults, etc.
lib/                         # Shared logic (package managers, OS detection, yadm bootstrap)
packages/                    # Declarative package lists, organized by category
scripts/                     # Small standalone tools that go on $PATH
profiles/                    # Installation profiles (slim, full, dev, etc.)
os/                          # OS-specific overrides or installers
```

### Installation profiles

Profiles control what gets installed. A `slim` profile installs only essentials; `full` adds GUI apps, language runtimes, fonts, etc. Profiles are composable — `full` should include `slim`. Each profile maps to a set of package groups and script subsets.

### Cross-platform approach

- The primary shell language is Bash for Linux/macOS. Windows uses PowerShell.
- OS detection should be early and centralized (single function/file, not scattered `if [[ $OSTYPE ]]` blocks throughout).
- Package manager abstraction: `brew` (macOS), `apt`/`dnf` (Linux), `winget`/`scoop` (Windows). The abstraction layer lives in `lib/` and callers shouldn't need to know which manager is active.

### Dotfiles

Dotfiles are managed in a **separate repo** via [yadm](https://yadm.io). This repo does not contain dotfiles. `install.sh` installs yadm, then bootstraps the dotfiles repo by running `yadm clone <repo>`. The target repo URL is defined in `config.sh` as `YADM_REPO`.

### Scripts on `$PATH`

Small utility scripts live in `scripts/`. The installer symlinks this directory (or its contents) into a location already on `$PATH` (e.g. `~/.local/bin`). Scripts should be self-contained and not depend on each other unless clearly documented.

## Key design constraints

- **Idempotent**: every installer step must be safe to run multiple times.
- **Fail-fast with clear messages**: if a prerequisite is missing or a step fails, stop and tell the user what happened and what to do.
- **No silent mutations**: don't overwrite existing dotfiles without asking or backing them up.
- **Profiles are additive**: `slim` ⊂ `full`. Never design a profile that requires another profile to be un-installed.
