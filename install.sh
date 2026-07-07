#!/usr/bin/env sh
# install.sh — workstation-cc entry point (macOS / Linux).
#
# This is intentionally thin. It loads config, runs the preflight (installs
# prerequisites, downloads the verified worker binary), points the worker at this
# checkout (WORKSTATION_ROOT) so it can symlink dotfiles/ and app-configs/, then
# hands off. All real logic is in Go.
#
# Usage:
#   ./install.sh                 # launch the worker (interactive TUI if a terminal)
#   ./install.sh install --help  # forward args straight to the worker
#
# Any arguments are passed through to the worker unchanged.

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

# Load user configuration if present (sets WORKSTATION_REPO, WORKSTATION_VERSION, etc.).
if [ -f "${SCRIPT_DIR}/config.sh" ]; then
	# shellcheck source=config.sh
	. "${SCRIPT_DIR}/config.sh"
fi

# shellcheck source=lib/preflight.sh
. "${SCRIPT_DIR}/lib/preflight.sh"

# Prepare prerequisites and download the worker; sets WORKSTATION_BIN.
workstation_preflight

# The wizard symlinks dotfiles/ and app-configs/ from a checkout of this repo.
# By default that is the directory this script lives in (the repo you cloned).
: "${WORKSTATION_ROOT:=${SCRIPT_DIR}}"
export WORKSTATION_ROOT

# Hand off to the worker (replacing this process so signals/exit codes pass through).
exec "${WORKSTATION_BIN}" "$@"
