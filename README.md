# workstation-cc

Welcome to the Workstation Command Center! This is a setup utility to configure new machines. Clone this first and you will be off at blazing speed.

## What it does

- Installs CLI utilities and GUI applications
- Bootstraps dotfiles via [yadm](https://yadm.io) from a separate dotfiles repo
- Adds small utility scripts to `$PATH`

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
| `--dry-run` | Preview changes without applying them |

## Dotfiles

Dotfiles are managed via yadm in a separate repository. `install.sh` installs yadm and bootstraps the dotfiles repo automatically. To change which dotfiles repo is used, update `YADM_REPO` in `config.sh`.
