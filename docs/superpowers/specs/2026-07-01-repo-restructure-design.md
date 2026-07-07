# Repo Restructure: Encapsulate the Wizard, Own the Dotfiles

**Date:** 2026-07-01
**Status:** Approved (design)

## Goal

Reorganize the repository so all installer logic (the Go CLI) lives in a single
self-contained directory, and the workstation *content* the installer acts on
(packages, dotfiles, app configs, docs) sits at the root as first-class,
editable inputs. As part of this, the repo stops delegating dotfiles to a
separate yadm repo and instead **owns** its dotfiles, which the wizard applies
by symlinking from a local checkout.

## Target Layout

```
workstation-cc/
├── install.sh                 # thin entry: ensure checkout → run wizard
├── install.ps1                # same for Windows
├── config.sh                  # user-facing config (repo, version, toggles)
├── README.md
├── CLAUDE.md
├── docs/
├── packages/                  # declarative catalog (core.yaml, dev.yaml, gui.yaml)
├── profiles.yaml              # profile definitions (stays beside packages/)
├── dotfiles/                  # repo-owned dotfiles, mirrors $HOME
├── app-configs/               # app-specific configs + manifest.yaml
│   └── manifest.yaml          # maps each entry → destination (per-OS aware)
└── wizard/                    # ALL Go code, one self-contained module
    ├── go.mod / go.sum
    ├── cmd/workstation/
    ├── internal/
    ├── embed.go
    ├── generate.go            # //go:generate copy of ../packages, ../profiles.yaml
    ├── packages/              # gitignored build-time copy (embedded)
    └── profiles.yaml          # gitignored build-time copy (embedded)
```

## Key Decisions

1. **Data files: build-step copy.** `packages/` and `profiles.yaml` stay
   editable at the root. Go's `//go:embed` cannot reach outside the package's
   directory tree, so a copy step brings them into `wizard/` at build time. The
   copied files are gitignored; `embed.go` embeds them unchanged, so the binary
   keeps its zero-runtime-deps catalog.

2. **Dotfiles: repo-owned, not yadm.** This repo now owns dotfiles. The separate
   yadm repo model is removed. `dotfiles/` mirrors `$HOME`; `app-configs/` holds
   app-specific configs whose destinations are declared in a manifest.

3. **Runtime source: run from a checkout.** Because you cannot symlink to a file
   baked inside a binary, `install.sh` ensures a git checkout of this repo exists
   locally and runs the wizard against it, so `dotfiles/` and `app-configs/` are
   real files the wizard can link.

4. **Full scope:** this effort includes both the directory reorg AND the wizard's
   dotfiles rewrite (yadm removal + symlink logic), plus doc updates.

## Components

### wizard/ (Go module)

- The Go module path stays `github.com/BlakeMarterella/workstation-cc` (module
  path is independent of on-disk location), so internal import paths do not
  churn. Only `go.mod`/`go.sum` and the source tree relocate under `wizard/`.
- `.goreleaser.yaml` builds from `wizard/` (`main: ./cmd/workstation`, with the
  build working directory set to `wizard/`), and its `before` hook runs the data
  copy so embeds resolve on CI.
- `generate.go` declares `//go:generate` to copy `../packages` and
  `../profiles.yaml` into `wizard/` for local `go generate ./...` builds.

### Data copy step

- One mechanism, invoked from two places: local `go generate` and the goreleaser
  `before` hook. Implemented as a tiny Go program or a shell snippet — chosen
  during planning; must be idempotent and overwrite the gitignored copies.
- `.gitignore` adds `wizard/packages/` and `wizard/profiles.yaml`.

### install.sh / install.ps1 (clone-then-run)

- Ensure a checkout at a known dir (default `~/.local/share/workstation-cc`,
  overridable): `git clone` if absent, `git pull --ff-only` if present.
- Obtain the wizard binary (download from Releases, or build from the checkout).
- Export `WORKSTATION_ROOT` (the checkout path) and hand off to the wizard so it
  can locate `dotfiles/` and `app-configs/`.
- Remain thin and idempotent; fail fast with clear messages.

### internal/dotfiles (rewrite)

- Remove all yadm code (`Clone`, yadm command checks, `YADM_REPO`).
- New responsibility: given the checkout root, symlink `dotfiles/` entries into
  `$HOME` and `app-configs/` entries into their manifest-declared destinations.
- **Idempotent:** a correct existing symlink is a no-op.
- **No silent mutation:** an existing real file (not our symlink) is backed up to
  `<name>.bak` (or the user is prompted) before linking — honoring the CLAUDE.md
  "no silent mutations" constraint.
- `--dry-run` prints every link/back-up it would perform.
- Keep the testable `Env`-style seam (filesystem operations behind an interface)
  so the linker is unit-testable without touching the real `$HOME`.

### app-configs/manifest.yaml

- Maps each entry to a destination path. Supports per-OS destinations, e.g.:

  ```yaml
  entries:
    - name: nvim
      dest:
        default: ~/.config/nvim
    - name: vscode-settings
      dest:
        darwin: ~/Library/Application Support/Code/User
        linux: ~/.config/Code/User
  ```

- Destinations expand `~` and may be OS-keyed; missing OS key → entry skipped
  with a notice (not an error).

### CLI surface changes

- `internal/cli/install.go`: drop `--dotfiles-repo` / `--skip-dotfiles` yadm
  wording; the dotfiles step now links from the checkout. Keep a `--skip-dotfiles`
  flag (now meaning "don't symlink"). Remove `WORKSTATION_YADM_REPO`.

### Docs

- `CLAUDE.md`: update the Dotfiles section (repo-owns-dotfiles, no yadm), the
  top-level structure (add `wizard/`, `dotfiles/`, `app-configs/`), and the
  cross-platform/entry-point notes (clone-then-run, build-step data copy).
- `README.md`: update install flow and layout to match.
- `config.sh`: replace `YADM_REPO` with checkout-location / dotfiles toggle;
  keep `WORKSTATION_REPO` / `WORKSTATION_VERSION`.

## Data Flow (install)

```
install.sh
  → ensure checkout at WORKSTATION_ROOT (clone/pull)
  → obtain wizard binary
  → exec wizard install --profile … (WORKSTATION_ROOT in env)
        → detect OS, select pkg manager
        → resolve profile → package groups → install missing (embedded catalog)
        → link dotfiles/ + app-configs/ from WORKSTATION_ROOT (backup on conflict)
        → print summary
```

## Error Handling

- Missing `git` for clone-then-run → fail fast with install guidance.
- Existing non-symlink at a dotfile destination → back up, never overwrite silently.
- Manifest entry with no destination for the current OS → skip with a notice.
- All destructive steps are gated behind `--yes`; `--dry-run` previews.

## Testing

- `internal/dotfiles`: table-driven tests over a fake filesystem seam covering:
  fresh link, already-correct link (idempotent), conflicting real file (backup),
  broken/foreign symlink, dry-run (no writes).
- Manifest parsing: per-OS resolution, `~` expansion, missing-OS skip.
- Existing package/profile/installer tests continue to pass after the move.
- A build smoke check that `go generate ./... && go build ./...` works from a
  clean tree (data copy populates the embedded files).

## Out of Scope

- Migrating existing dotfile *content* from the old yadm repo (user does this
  separately by dropping files into `dotfiles/`).
- Windows symlink parity beyond what `install.ps1` reasonably supports (documented
  if limited).

## Naming

- Directory is `wizard/` (chosen over `src/`).
