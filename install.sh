#!/usr/bin/env bash
# install.sh — workstation-cc bootstrap script
# Sets up a macOS workstation via yadm dotfiles.
#
# Usage:
#   ./install.sh            # Dry-run (default) — prints what would happen
#   ./install.sh --execute  # Actually apply changes
#   ./install.sh --help     # Show this help

set -euo pipefail

# ---------------------------------------------------------------------------
# Defaults
# ---------------------------------------------------------------------------

DRY_RUN=true
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default dotfiles repo — may be overridden by config.sh
YADM_REPO="https://github.com/BlakeMarterella/workstation-dotfiles"

# ---------------------------------------------------------------------------
# Load optional config
# ---------------------------------------------------------------------------

if [[ -f "${SCRIPT_DIR}/config.sh" ]]; then
    # shellcheck source=config.sh
    source "${SCRIPT_DIR}/config.sh"
fi

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

log() {
    echo "[workstation-cc] $*"
}

dry_log() {
    echo "[DRY RUN] $*"
}

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Bootstrap a macOS workstation with yadm dotfiles.

Options:
  --execute   Actually apply changes (default is dry-run)
  --help      Show this help and exit

Dry-run mode is ON by default. Pass --execute to make real changes.

Config:
  Edit config.sh in the same directory to override defaults.
  Current YADM_REPO: ${YADM_REPO}
EOF
}

run() {
    # run <description> <cmd> [args...]
    local description="$1"
    shift

    if [[ "${DRY_RUN}" == true ]]; then
        dry_log "${description}"
        dry_log "  Would run: $*"
    else
        log "${description}"
        "$@"
    fi
}

# ---------------------------------------------------------------------------
# OS detection
# ---------------------------------------------------------------------------

detect_os() {
    local os
    os="$(uname -s)"

    case "${os}" in
        Darwin)
            log "Detected macOS — OK."
            ;;
        *)
            echo "ERROR: This script only supports macOS. Detected OS: ${os}" >&2
            exit 1
            ;;
    esac
}

# ---------------------------------------------------------------------------
# Steps
# ---------------------------------------------------------------------------

check_homebrew() {
    if command -v brew &>/dev/null; then
        log "Homebrew is already installed — skipping."
        return
    fi

    run \
        "Install Homebrew (package manager for macOS)" \
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
}

install_yadm() {
    if command -v yadm &>/dev/null; then
        log "yadm is already installed — skipping."
        return
    fi

    run \
        "Install yadm via Homebrew" \
        brew install yadm
}

clone_dotfiles() {
    if yadm status &>/dev/null 2>&1; then
        log "yadm repository already initialised — skipping clone."
        return
    fi

    run \
        "Clone dotfiles from ${YADM_REPO}" \
        yadm clone "${YADM_REPO}"
}

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------

parse_args() {
    for arg in "$@"; do
        case "${arg}" in
            --execute)
                DRY_RUN=false
                ;;
            --help|-h)
                usage
                exit 0
                ;;
            *)
                echo "Unknown option: ${arg}" >&2
                usage >&2
                exit 1
                ;;
        esac
    done
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

main() {
    parse_args "$@"

    if [[ "${DRY_RUN}" == true ]]; then
        log "Running in DRY-RUN mode. Pass --execute to apply changes."
        echo
    fi

    detect_os
    check_homebrew
    install_yadm
    clone_dotfiles

    echo
    if [[ "${DRY_RUN}" == true ]]; then
        log "Dry-run complete. Re-run with --execute to apply the above changes."
    else
        log "Bootstrap complete."
    fi
}

main "$@"
