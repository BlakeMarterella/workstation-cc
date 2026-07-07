# workstation-cc

Welcome to the Workstation Command Center! This is a setup utility to configure new machines. Clone this first and you will be off at blazing speed.

## What it does

- Installs CLI utilities and GUI applications
- Symlinks repo-owned dotfiles (`dotfiles/` and `app-configs/`) into place
- Applies installation profiles across macOS, Linux, and Windows

## Architecture

The real work is done by a small, fast **Go binary** (`workstation`). The
shell entry points only bootstrap it:

1. **`install.sh` / `install.ps1`** — thin entry point. Loads `config.sh`, runs
   the preflight (`lib/preflight.sh`: installs prerequisites and downloads a
   checksum-verified worker binary for this OS/arch), sets `WORKSTATION_ROOT` to
   the repo checkout directory, and hands off to the worker binary.
2. **`workstation` (Go, lives in `wizard/`)** — detects the platform, resolves
   an installation profile into packages, installs the missing ones via the
   host package manager, and symlinks the repo's `dotfiles/` and `app-configs/`
   directories into place.

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
| `--skip-dotfiles` | Don't symlink dotfiles/ or app-configs/. |
| `--root <path>` | Repo checkout to link dotfiles from. Defaults to `WORKSTATION_ROOT`. |

Profiles are additive and composable (`slim ⊂ full`) and are defined in
`profiles.yaml`. Packages are declared in `packages/*.yaml`, organized by
category, and embedded into the binary at build time.

## Dotfiles

Dotfiles are **owned by this repo** — no separate dotfiles repository or
external tool needed. Two directories are symlinked into your home:

- **`dotfiles/`** — mirrors `$HOME`. Every file here becomes a symlink at the
  corresponding path under `$HOME`.
- **`app-configs/`** — app-specific configs with per-OS destinations declared
  in `app-configs/manifest.yaml`.

The wizard reads `WORKSTATION_ROOT` (set automatically by the entry scripts to
the repo checkout directory) to find these directories. Existing real files are
backed up to `<name>.bak` before a symlink is created — they are never
overwritten silently.

## Development

The worker lives in `wizard/` as a standard Go module
(`github.com/BlakeMarterella/workstation-cc`). You only need the
[Go toolchain](https://go.dev/dl/) (1.26+) installed — everything else is a Go
dependency resolved by `go`.

### Layout

```
wizard/                       # Self-contained Go module
  cmd/workstation/            # main package — entry point
  internal/osdetect/          # OS + arch detection
  internal/pkgmgr/            # package-manager abstraction (brew, apt/dnf/winget stubs)
  internal/packages/          # loads the embedded packages/*.yaml catalog
  internal/profiles/          # profile resolution (slim ⊂ full)
  internal/dotfiles/          # dotfile symlink logic (LinkTree, LinkApps, manifest)
  internal/installer/         # builds and executes an install plan
  internal/tui/               # Bubble Tea interactive UI
  internal/cli/               # Cobra command wiring
  internal/ui/                # shared Lip Gloss styles
  tools/gendata/              # go generate tool: copies packages/ + profiles.yaml into wizard/
  packages/*.yaml             # copied at build time from root packages/ (gitignored)
  profiles.yaml               # copied at build time from root profiles.yaml (gitignored)
packages/*.yaml               # declarative package catalog (source of truth)
profiles.yaml                 # profile definitions (source of truth)
dotfiles/                     # repo-owned dotfiles, mirrors $HOME
app-configs/                  # app-specific configs + manifest.yaml
```

### Build from source

Because `//go:embed` cannot reach parent directories, you must run
`go generate ./...` first to copy `packages/` and `profiles.yaml` into
`wizard/` before building:

```sh
cd wizard
go generate ./...             # copies packages/ + profiles.yaml into wizard/
go build -o workstation ./cmd/workstation
./workstation --version
```

GoReleaser runs `go generate ./...` automatically as a before-hook, so release
builds do not require a manual step.

### Run from source

```sh
cd wizard
go generate ./...                                         # copies root packages/ + profiles.yaml in (re-run after editing them)
go run ./cmd/workstation --help
go run ./cmd/workstation install --profile slim --dry-run  # safe preview
go run ./cmd/workstation                                    # launch the TUI (needs a terminal)
```

`--dry-run` never changes your system, so it's the safe way to test package
resolution and output. Omit it (and pass `--yes`) only when you actually want
packages installed.

### Cross-compile

```sh
cd wizard
GOOS=linux  GOARCH=amd64 go build -o workstation-linux ./cmd/workstation
GOOS=darwin GOARCH=arm64 go build -o workstation-mac   ./cmd/workstation
```

Release builds are produced by [GoReleaser](https://goreleaser.com); preview the
full matrix locally with `goreleaser release --snapshot --clean` (output in
`dist/`).

### Test, format, vet

```sh
cd wizard
go test ./...          # run all unit tests
go test -count=1 ./... # run tests without the cache
gofmt -l .             # list unformatted files (should print nothing)
gofmt -w .             # format in place
go vet ./...           # static checks
```

These are the same checks CI runs (`.github/workflows/ci.yml`). New code follows
test-driven development — write the failing test first, then the implementation.

### Add packages or profiles

Package data and profiles are plain YAML. The source files live at the repo root:

- Add a package: edit the relevant `packages/<category>.yaml`. Give it a `name`,
  a `description`, and a `managers:` map of package-manager → package id (add
  `cask: true` for macOS GUI apps).
- Add or change a profile: edit `profiles.yaml`. A profile lists `groups:`
  (category names) and may `includes:` other profiles.

After editing, run `go generate ./...` (from `wizard/`) so the copies inside
`wizard/` are updated, then verify end-to-end with a dry run:

```sh
cd wizard && go generate ./... && go run ./cmd/workstation install --profile full --dry-run
```
