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

## Development

The worker lives in this repo as a standard Go module
(`github.com/BlakeMarterella/workstation-cc`). You only need the
[Go toolchain](https://go.dev/dl/) (1.26+) installed — everything else is a Go
dependency resolved by `go`.

### Layout

```
cmd/workstation/      # main package — entry point
internal/osdetect/    # OS + arch detection
internal/pkgmgr/      # package-manager abstraction (brew, apt/dnf/winget stubs)
internal/packages/    # loads the embedded packages/*.yaml catalog
internal/profiles/    # profile resolution (slim ⊂ full)
internal/dotfiles/    # yadm bootstrap
internal/installer/   # builds and executes an install plan
internal/tui/         # Bubble Tea interactive UI
internal/cli/         # Cobra command wiring
internal/ui/          # shared Lip Gloss styles
packages/*.yaml       # declarative package catalog (embedded at build time)
profiles.yaml         # profile definitions (embedded at build time)
```

### Run from source

No build step needed while iterating — `go run` compiles and runs in one step:

```sh
go run ./cmd/workstation --help
go run ./cmd/workstation install --profile slim --dry-run   # safe preview
go run ./cmd/workstation                                    # launch the TUI (needs a terminal)
```

`--dry-run` never changes your system, so it's the safe way to test package
resolution and output. Omit it (and pass `--yes`) only when you actually want
packages installed.

### Build a binary

```sh
go build -o workstation ./cmd/workstation   # produces ./workstation (gitignored)
./workstation --version
```

Cross-compile for another platform by setting `GOOS`/`GOARCH`:

```sh
GOOS=linux  GOARCH=amd64 go build -o workstation-linux ./cmd/workstation
GOOS=darwin GOARCH=arm64 go build -o workstation-mac   ./cmd/workstation
```

Release builds are produced by [GoReleaser](https://goreleaser.com); preview the
full matrix locally with `goreleaser release --snapshot --clean` (output in
`dist/`).

### Test, format, vet

```sh
go test ./...        # run all unit tests
go test -count=1 ./... # run tests without the cache
gofmt -l .           # list unformatted files (should print nothing)
gofmt -w .           # format in place
go vet ./...         # static checks
```

These are the same checks CI runs (`.github/workflows/ci.yml`). New code follows
test-driven development — write the failing test first, then the implementation.

### Add packages or profiles

Package data and profiles are plain YAML embedded into the binary via `go:embed`,
so **rebuild after editing** for changes to take effect:

- Add a package: edit the relevant `packages/<category>.yaml`. Give it a `name`,
  a `description`, and a `managers:` map of package-manager → package id (add
  `cask: true` for macOS GUI apps).
- Add or change a profile: edit `profiles.yaml`. A profile lists `groups:`
  (category names) and may `includes:` other profiles.

Verify a change end-to-end with a dry run, e.g. `go run ./cmd/workstation install
--profile full --dry-run`.

## Dotfiles

Dotfiles are managed via yadm in a separate repository. The worker installs yadm
(as part of the `core` group) and clones the dotfiles repo. An already-cloned
repo is left untouched. To change which dotfiles repo is used, update `YADM_REPO`
in `config.sh` (or set `WORKSTATION_YADM_REPO`).
