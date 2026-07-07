# workstation-wizard: Drive-backed snapshot & sync — Design

**Date:** 2026-07-06
**Status:** Approved (design); pending spec review before planning

## Summary

Evolve the existing setup tool from a one-shot installer into an installed,
long-lived CLI utility named **`workstation-wizard`**. Beyond first-time setup it
lets you **snapshot** your current curated workstation state and **restore/sync**
it onto other machines, using your own **Google Drive** (private, hidden
app-data folder) as the sync hub. Authentication uses Google's OAuth **device
flow** so it works on headless Linux servers, with a **token-reuse** fallback for
CI/automation.

The tool binary continues to ship publicly (GitHub releases + checksum); only
your *config data* moves behind Google auth.

## Goals

- Install as `workstation-wizard`, callable from `$PATH`.
- `snapshot`: capture current curated state → push to Drive.
- `restore` / `sync`: pull a snapshot from Drive → apply it.
- Works on headless Linux servers (no local browser).
- No secret exposure; no config-injection/tamper risk.
- Keep the existing one-line bootstrap for install.

## Non-goals

- Capturing secret files (`.ssh`, credentials, tokens) — excluded by design.
- Service-account / fleet automation (possible later add-on, not in this scope).
- Client-side encryption (unnecessary once storage is private + authenticated).
- Replacing the public release-binary distribution channel.

## Decisions (from brainstorming)

| Decision | Choice |
|---|---|
| Command name | `workstation-wizard` (installed to `~/.local/bin`) |
| Storage backend | Google Drive, hidden `appDataFolder` |
| Access model | Private, authenticated as the user (read + write) |
| Auth flow | OAuth **device flow** (primary) + **token reuse** (fallback) |
| Snapshot scope | Curated `dotfiles/` + `app-configs/` + installed-package manifest |
| Secrets | Excluded from snapshots by design |
| Encryption | None (private storage makes it unnecessary) |
| Drive scope | `drive.appdata` — **verified** supported by device flow |

## Architecture

### 1. Naming & installation

- Built binary and installed command: **`workstation-wizard`**.
- Go module path unchanged (`github.com/BlakeMarterella/workstation-cc`).
- Cobra root command `Use:` becomes `workstation-wizard`; the `cmd/` entry
  point produces a binary of that name.
- Preflight (`lib/preflight.sh`) installs to
  `~/.local/bin/workstation-wizard` and ensures it is on `$PATH` (existing
  behavior, renamed target).
- Bootstrap one-liner (`install.sh` → preflight → download public release
  binary) is unchanged.

### 2. Command surface

```
workstation-wizard                          # interactive TUI (existing)
workstation-wizard install [--profile ...]  # first-time setup (existing);
                                            #   offers to restore if a snapshot exists
workstation-wizard login                    # OAuth device flow → cache refresh token
workstation-wizard logout                   # clear cached token
workstation-wizard snapshot [--message]     # capture current curated state → push to Drive
workstation-wizard restore [--from <id>]    # pull a snapshot from Drive → apply
workstation-wizard status                   # tracked files, drift vs. last snapshot, auth state
```

`restore` reuses the existing `LinkTree` / `LinkApps` + installer logic.
`sync` is an alias/behavior of `restore --from latest` (pull-and-apply latest).

### 3. Data model

- **Local state dir** (`~/.local/share/workstation-wizard/`): the working copy
  of your curated config — `dotfiles/`, `app-configs/`, and a generated
  `packages.lock`. This becomes *your* source of truth for personal config; the
  git repo's `dotfiles/` becomes shipped defaults/examples.
- **`packages.lock`**: the set of catalog packages currently installed, recorded
  via the existing `pkgmgr` abstraction so `restore` can re-install them.
- **`snapshot`**: tars the curated set + `packages.lock` → uploads to Drive as
  `snapshot-<timestamp>.tar.gz` and updates a `latest` pointer. Retains the last
  N (default 10); older snapshots pruned.
- **`restore`**: downloads (default `latest`), unpacks into local state,
  symlinks into `$HOME` (existing backup-to-`.bak` behavior — never overwrite
  silently), then installs any missing packages from `packages.lock`.
- **Drive layout**: all objects live in the hidden **`appDataFolder`** —
  invisible in the normal Drive UI and scoped so the app sees only its own files.
- **Secret-free by design**: snapshot captures only the curated tree + package
  manifest. No `.ssh`, credentials, or token files.

### 4. Authentication

- **Primary — OAuth device flow**: `login` prints a verification URL + user
  code; the user completes it on any device; the machine polls the token
  endpoint and caches a refresh token at
  `~/.config/workstation-wizard/token.json` (mode `0600`). Fully headless.
- **Fallback — token reuse**: run `login` once on a machine with a browser, then
  distribute the refresh token to servers/CI via a `WORKSTATION_WIZARD_TOKEN`
  env var or a `0600` file referenced in `config.sh`. Never committed to git.
- **Scopes**: `https://www.googleapis.com/auth/drive.appdata` + basic profile.
  Confirmed supported for the device flow per Google's limited-input-device
  documentation.

### 5. Security posture

- Private, Google-authenticated storage ⇒ no secret exposure, no
  config-injection/tamper risk, no client-side encryption or passphrase needed.
- The refresh token is the sensitive artifact: stored `0600`, never in git,
  cleared by `logout`.
- Snapshots exclude secrets, so even a Drive-access compromise cannot leak
  `.ssh`/credentials.
- Tool distribution stays public and checksum-verified (unchanged trust model).

### 6. New Go components

All follow the existing convention: small, single-purpose packages under
`wizard/internal/`, built test-first.

- `internal/auth/` — device-flow client + token cache (load/save/clear, env +
  file fallback).
- `internal/gdrive/` — Google Drive API wrapper (upload, download, list, prune)
  behind an interface so it can be faked in tests.
- `internal/snapshot/` — pack/unpack of the curated tree + `packages.lock`
  generation and application.
- `internal/cli/` — new Cobra commands (`login`, `logout`, `snapshot`,
  `restore`, `status`) wired to the above.

### 7. Error handling

- Fail-fast with clear messages (existing constraint): missing/expired token →
  tell the user to run `login`; no snapshot found → say so; partial restore →
  report what applied and what did not.
- Idempotent: re-running `restore` or `snapshot` is safe; symlinking preserves
  the existing `.bak` backup behavior.

## Testing

Test-driven throughout, matching current CI (`go test`, `gofmt -l`, `go vet`):

- `internal/snapshot/`: pack→unpack round-trips; `packages.lock` gen/apply
  against existing `pkgmgr` fakes.
- `internal/auth/`: token cache load/save/clear; env + file fallback precedence.
- `internal/gdrive/`: exercised via its interface with a fake implementation —
  no network in unit tests.
- `internal/cli/`: command wiring / flag parsing.

## Open items

None blocking. The one prior unknown (device-flow support for `drive.appdata`)
is **verified** against Google's documentation.

## Rollout notes

- Requires registering a Google OAuth client (client ID/secret shipped in the
  binary for an installed-app/device client) — a setup task in the plan.
- Docs (`README.md`, `CLAUDE.md`) to be updated to cover the new command name,
  subcommands, and the Drive/auth model as part of implementation.
