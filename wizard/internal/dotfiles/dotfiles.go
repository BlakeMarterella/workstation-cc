// Package dotfiles symlinks repo-owned dotfiles and app configs into place. It
// replaces the previous yadm bootstrap: this repo now owns its dotfiles, and the
// wizard links them from the local checkout. Linking is idempotent and never
// overwrites an existing real file silently — conflicts are backed up first.
package dotfiles

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Action reports what LinkTree/LinkEntry did for one destination.
type Action int

const (
	// ActionLinked means a new symlink was created.
	ActionLinked Action = iota
	// ActionAlreadyLinked means the correct symlink already existed (no-op).
	ActionAlreadyLinked
	// ActionBackedUp means an existing real file was moved to <name>.bak, then linked.
	ActionBackedUp
	// ActionSkipped means nothing was done (e.g. dry-run, or unresolved destination).
	ActionSkipped
)

func (a Action) String() string {
	switch a {
	case ActionLinked:
		return "linked"
	case ActionAlreadyLinked:
		return "already linked"
	case ActionBackedUp:
		return "backed up + linked"
	case ActionSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// Result describes one linking outcome.
type Result struct {
	Path   string // source file (inside the checkout)
	Dest   string // destination path
	Action Action
	Note   string // optional human-readable detail (e.g. backup path)
}

// FS is the filesystem seam so linking is testable without touching real $HOME.
type FS interface {
	WalkFiles(root string, fn func(path string) error) error // visit regular files only
	Readlink(name string) (string, error)                    // "" err if not a symlink
	Lstat(name string) (exists bool, isSymlink bool, err error)
	MkdirAll(dir string) error
	Symlink(oldname, newname string) error
	Rename(oldpath, newpath string) error
}

// Linker creates symlinks from a checkout into home.
type Linker struct {
	fs     FS
	home   string
	dryRun bool
}

// NewLinker returns a Linker targeting home. When dryRun is true no changes are made.
func NewLinker(fsys FS, home string, dryRun bool) *Linker {
	return &Linker{fs: fsys, home: home, dryRun: dryRun}
}

// LinkTree walks srcRoot and links every regular file into home at the same
// relative path (srcRoot/.vimrc -> home/.vimrc).
func (l *Linker) LinkTree(srcRoot string) ([]Result, error) {
	var results []Result
	err := l.fs.WalkFiles(srcRoot, func(path string) error {
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(l.home, rel)
		res, err := l.LinkOne(path, dest)
		if err != nil {
			return err
		}
		results = append(results, res)
		return nil
	})
	return results, err
}

// LinkOne symlinks src -> dest, backing up any conflicting real file first.
func (l *Linker) LinkOne(src, dest string) (Result, error) {
	res := Result{Path: src, Dest: dest}

	// Already the correct symlink? No-op.
	if target, err := l.fs.Readlink(dest); err == nil && target == src {
		res.Action = ActionAlreadyLinked
		return res, nil
	}

	exists, isSymlink, err := l.fs.Lstat(dest)
	if err != nil {
		return res, err
	}

	if l.dryRun {
		res.Action = ActionSkipped
		if exists && !isSymlink {
			res.Note = "would back up existing file, then link"
		} else {
			res.Note = "would link"
		}
		return res, nil
	}

	if err := l.fs.MkdirAll(filepath.Dir(dest)); err != nil {
		return res, err
	}

	if exists {
		// A real file or a wrong/foreign symlink: back it up, never overwrite silently.
		backup := dest + ".bak"
		if err := l.fs.Rename(dest, backup); err != nil {
			return res, fmt.Errorf("back up %s: %w", dest, err)
		}
		res.Action = ActionBackedUp
		res.Note = backup
	} else {
		res.Action = ActionLinked
	}

	if err := l.fs.Symlink(src, dest); err != nil {
		return res, fmt.Errorf("symlink %s -> %s: %w", dest, src, err)
	}
	return res, nil
}

// OSFS is the production FS backed by the os package.
type OSFS struct{}

// WalkFiles implements FS.
func (OSFS) WalkFiles(root string, fn func(path string) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		return fn(path)
	})
}

// Readlink implements FS.
func (OSFS) Readlink(name string) (string, error) { return os.Readlink(name) }

// Lstat implements FS.
func (OSFS) Lstat(name string) (bool, bool, error) {
	fi, err := os.Lstat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, fi.Mode()&os.ModeSymlink != 0, nil
}

// MkdirAll implements FS.
func (OSFS) MkdirAll(dir string) error { return os.MkdirAll(dir, 0o755) }

// Symlink implements FS.
func (OSFS) Symlink(oldname, newname string) error { return os.Symlink(oldname, newname) }

// Rename implements FS.
func (OSFS) Rename(oldpath, newpath string) error { return os.Rename(oldpath, newpath) }
