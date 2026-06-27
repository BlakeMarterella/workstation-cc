# workstation-cc

Welcome to the Workstation Command Center! This is a setup utility to configure new machines. Clone this first and you will be off at blazing speed.

## What it does

- Installs CLI utilities and GUI applications
- Bootstraps dotfiles via [yadm](https://yadm.io) from a separate dotfiles repo
- Adds small utility scripts to `$PATH`

## Architecture

The real work is done by a small, fast **Go worker binary** (`workstation`). The
shell entry points only bootstrap it:

1. **`install.sh` / `install.ps1`** — thin entry point. Loads `config.sh`, runs
   the preflight, then hands off to the worker.
2. **`lib/preflight.sh`** — installs prerequisites (and Homebrew on macOS), then
   downloads the correct release binary for your OS/arch from GitHub Releases and
   **verifies its SHA-256 checksum** before running it.
3. **`workstation` (Go)** — detects the platform, resolves an installation
   profile into packages, installs the missing ones via the host package manager,
   and bootstraps the yadm dotfiles repo.

The worker is a single static binary with no runtime dependency, built and
published by GoReleaser (`.goreleaser.yaml` + `.github/workflows/release.yml`).

## Usage

**macOS / Linux**
```sh
./install.sh                                   # interactive: launches the TUI in a terminal
./install.sh install --profile slim --dry-run  # preview, no changes
./install.sh install --profile full --yes      # apply the full profile
```

**Windows**
```powershell
.\install.ps1
.\install.ps1 install --profile slim --dry-run
```

Run with no subcommand in a terminal to get the interactive **TUI** (profile
picker + filterable package list). This works locally and over SSH. In a
non-interactive context (piped, CI), use the `install` subcommand with flags.

## The `install` command

| Flag | Description |
|---|---|
| `--profile <name>` | Profile to apply (`slim`, `full`). Default: `slim`. |
| `--dry-run` | Preview actions without changing anything. |
| `--yes` | Apply changes (required to mutate; fail-fast otherwise). |
| `--skip-dotfiles` | Don't bootstrap the yadm dotfiles repo. |
| `--dotfiles-repo <url>` | Override the dotfiles repository URL. |

Profiles are additive and composable (`slim ⊂ full`) and are defined in
`profiles.yaml`. Packages are declared in `packages/*.yaml`, organized by
category, and embedded into the binary at build time.

## Dotfiles

Dotfiles are managed via yadm in a separate repository. The worker installs yadm
(as part of the `core` group) and clones the dotfiles repo. An already-cloned
repo is left untouched. To change which dotfiles repo is used, update `YADM_REPO`
in `config.sh` (or set `WORKSTATION_YADM_REPO`).
