# app-configs/

App-specific configuration whose destination is NOT a simple `$HOME`-relative
path (e.g. `~/.config/nvim`, `~/Library/Application Support/...`). Each top-level
entry is mapped to its destination by `manifest.yaml`, which supports per-OS
destinations. The wizard symlinks each entry to its resolved destination.
