#!/usr/bin/env sh
# lib/preflight.sh — installs the prerequisites needed to run the worker binary
# and downloads (with checksum verification) the correct release binary for this
# host. Sourced by install.sh; defines workstation_preflight, which sets
# WORKSTATION_BIN to the path of the ready-to-run worker.
#
# POSIX sh only — no bashisms — so it runs on a bare machine.

# Defaults (overridable via config.sh / environment).
: "${WORKSTATION_REPO:=BlakeMarterella/workstation-cc}"
: "${WORKSTATION_VERSION:=latest}"
: "${WORKSTATION_BIN_DIR:=${HOME}/.local/bin}"

workstation_die() {
	echo "[preflight] ERROR: $*" >&2
	exit 1
}

workstation_log() {
	echo "[preflight] $*"
}

# Detect os/arch and set WS_OS / WS_ARCH using the same names GoReleaser uses.
workstation_detect_platform() {
	WS_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
	case "$WS_OS" in
		darwin | linux) ;;
		*) workstation_die "unsupported OS: $WS_OS (use install.ps1 on Windows)" ;;
	esac

	WS_ARCH=$(uname -m)
	case "$WS_ARCH" in
		x86_64 | amd64) WS_ARCH=amd64 ;;
		arm64 | aarch64) WS_ARCH=arm64 ;;
		*) workstation_die "unsupported architecture: $WS_ARCH" ;;
	esac
}

# Ensure the tools the rest of preflight (and the worker) need are present.
workstation_ensure_prereqs() {
	command -v curl >/dev/null 2>&1 || workstation_die "curl is required but not installed"

	# The worker shells out to the host package manager; on macOS that's
	# Homebrew. Install it here so the worker has something to delegate to.
	if [ "$WS_OS" = "darwin" ] && ! command -v brew >/dev/null 2>&1; then
		workstation_log "Homebrew not found — installing it"
		/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" \
			|| workstation_die "Homebrew installation failed"
	fi
}

# Echo the sha256 of a file in "<hash>  <file>"-comparable bare-hash form.
workstation_sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | awk '{print $1}'
	else
		workstation_die "no sha256 tool (sha256sum/shasum) available"
	fi
}

# Build the base release-download URL for the configured version.
workstation_release_base() {
	if [ "$WORKSTATION_VERSION" = "latest" ]; then
		echo "https://github.com/${WORKSTATION_REPO}/releases/latest/download"
	else
		echo "https://github.com/${WORKSTATION_REPO}/releases/download/${WORKSTATION_VERSION}"
	fi
}

# Download the worker binary, verify its checksum, install it, and set
# WORKSTATION_BIN. Idempotent: re-running overwrites with a freshly verified copy.
workstation_download() {
	asset="workstation_${WS_OS}_${WS_ARCH}"
	base=$(workstation_release_base)
	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT

	workstation_log "Downloading ${asset} (${WORKSTATION_VERSION})"
	curl -fsSL "${base}/${asset}" -o "${tmp}/${asset}" \
		|| workstation_die "failed to download worker binary"
	curl -fsSL "${base}/checksums.txt" -o "${tmp}/checksums.txt" \
		|| workstation_die "failed to download checksums"

	expected=$(awk -v a="$asset" '$2 == a {print $1}' "${tmp}/checksums.txt")
	[ -n "$expected" ] || workstation_die "no checksum listed for ${asset}"
	actual=$(workstation_sha256 "${tmp}/${asset}")
	[ "$expected" = "$actual" ] \
		|| workstation_die "checksum mismatch for ${asset} (expected ${expected}, got ${actual})"

	mkdir -p "$WORKSTATION_BIN_DIR"
	install -m 0755 "${tmp}/${asset}" "${WORKSTATION_BIN_DIR}/workstation"
	WORKSTATION_BIN="${WORKSTATION_BIN_DIR}/workstation"
	workstation_log "Installed worker to ${WORKSTATION_BIN}"
}

# Top-level entry: prepare everything and set WORKSTATION_BIN.
workstation_preflight() {
	workstation_detect_platform
	workstation_ensure_prereqs
	workstation_download
}
