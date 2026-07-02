#!/usr/bin/env sh
# config.sh — workstation-cc configuration
# Edit this file to customize your setup before running install.sh.

# GitHub repository that publishes the worker binary releases.
WORKSTATION_REPO="BlakeMarterella/workstation-cc"

# Release version to install ("latest" or a tag like "v1.2.3").
WORKSTATION_VERSION="latest"

# Where the worker binary is installed (should be on your PATH).
WORKSTATION_BIN_DIR="${HOME}/.local/bin"

# Checkout of this repo the worker links dotfiles from. Defaults to the directory
# install.sh lives in; override to link from a different clone.
# WORKSTATION_ROOT="${HOME}/git/workstation-cc"
