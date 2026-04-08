# workstation-cc

Welcome to the Workstation Command Center! This is a setup utility to configure new machines. Clone this first and you will be off at blazing speed.

## What it does

- Installs CLI utilities and GUI applications
- Symlinks dotfiles into `$HOME`
- Adds small utility scripts to `$PATH`

## Dotfiles

| File | Description | Version | Updated |
|---|---|---|---|
| [`.vimrc`](dotfiles/.vimrc) | Vim configuration | 0.1.0 | 2026-04-08 |

## Usage

**macOS / Linux**
```sh
./install.sh          # full install
./install.sh --slim   # essentials only
```

**Windows**
```powershell
.\install.ps1
.\install.ps1 --slim
```

## Options

| Flag | Description |
|---|---|
| `--slim` | Install essentials only |
| `--full` | Install everything (default) |
| `--dotfiles-only` | Symlink dotfiles, skip packages |
| `--dry-run` | Preview changes without applying them |