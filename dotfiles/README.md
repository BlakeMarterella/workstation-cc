# dotfiles/

Files here mirror your `$HOME` directory. The wizard symlinks each file into
`$HOME` at the same relative path (e.g. `dotfiles/.vimrc` → `~/.vimrc`,
`dotfiles/.config/git/config` → `~/.config/git/config`).

Existing real files at a destination are backed up to `<name>.bak` before the
symlink is created — nothing is overwritten silently.

This `README.md` is documentation for the directory and is deliberately not
linked into `$HOME`.
