#!/usr/bin/env sh
# install.sh — workstation-cc entry point (macOS / Linux).
#
# This is intentionally thin. It loads config, runs the preflight (which installs
# prerequisites and downloads the verified worker binary), then hands off to the
# worker. All real logic lives in the Go worker.
#
# Usage:
#   ./install.sh                 # launch the worker (interactive TUI if a terminal)
#   ./install.sh install --help  # forward args straight to the worker
#
# Any arguments are passed through to the worker unchanged.

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

# Load user configuration if present (sets WORKSTATION_REPO, YADM_REPO, etc.).
if [ -f "${SCRIPT_DIR}/config.sh" ]; then
	# shellcheck source=config.sh
	. "${SCRIPT_DIR}/config.sh"
fi

# shellcheck source=lib/preflight.sh
. "${SCRIPT_DIR}/lib/preflight.sh"

# Prepare prerequisites and download the worker; sets WORKSTATION_BIN.
workstation_preflight

# Pass the configured dotfiles repo to the worker via its env override.
if [ -n "${YADM_REPO:-}" ]; then
	export WORKSTATION_YADM_REPO="${YADM_REPO}"
fi

# Hand off to the worker (replacing this process so signals/exit codes pass through).
exec "${WORKSTATION_BIN}" "$@"
