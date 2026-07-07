package dotfiles

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLinkTreeFreshLink(t *testing.T) {
	fs := newFakeFS()
	fs.addFile("/repo/dotfiles/.vimrc", "set number")
	l := NewLinker(fs, "/home/u", false)

	results, err := l.LinkTree("/repo/dotfiles")
	if err != nil {
		t.Fatalf("LinkTree: %v", err)
	}
	if len(results) != 1 || results[0].Action != ActionLinked {
		t.Fatalf("results = %+v, want one ActionLinked", results)
	}
	target, ok := fs.symlinks["/home/u/.vimrc"]
	if !ok || target != "/repo/dotfiles/.vimrc" {
		t.Errorf("symlink = %q (present=%v), want /repo/dotfiles/.vimrc", target, ok)
	}
}

func TestLinkTreeSkipsReadme(t *testing.T) {
	fs := newFakeFS()
	fs.addFile("/repo/dotfiles/.vimrc", "set number")
	fs.addFile("/repo/dotfiles/README.md", "docs about this dir")
	l := NewLinker(fs, "/home/u", false)

	results, err := l.LinkTree("/repo/dotfiles")
	if err != nil {
		t.Fatalf("LinkTree: %v", err)
	}
	// README.md is documentation, not a home-dir file: it must never be linked.
	if _, linked := fs.symlinks["/home/u/README.md"]; linked {
		t.Error("README.md was linked into home; it should be skipped")
	}
	if _, linked := fs.symlinks["/home/u/.vimrc"]; !linked {
		t.Error(".vimrc should still be linked")
	}
	var sawReadmeSkip bool
	for _, r := range results {
		if r.Path == "/repo/dotfiles/README.md" {
			if r.Action != ActionSkipped {
				t.Errorf("README.md action = %v, want ActionSkipped", r.Action)
			}
			sawReadmeSkip = true
		}
	}
	if !sawReadmeSkip {
		t.Error("expected a skipped Result for README.md")
	}
}

// fakeFS is an in-memory FS for testing the linker.
type fakeFS struct {
	files    map[string]string // path -> contents (regular files)
	symlinks map[string]string // path -> target
}

func newFakeFS() *fakeFS {
	return &fakeFS{files: map[string]string{}, symlinks: map[string]string{}}
}

func (f *fakeFS) addFile(path, contents string) { f.files[path] = contents }

func (f *fakeFS) WalkFiles(root string, fn func(path string) error) error {
	for p := range f.files {
		rel, err := filepath.Rel(root, p)
		// Skip files outside root: filepath.Rel returns a path starting with ".."
		// when p is not under root. Use strings.HasPrefix to avoid a panic when
		// rel is shorter than 2 chars (e.g. rel == "."), which would make rel[:2] panic.
		if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
			continue
		}
		if err := fn(p); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeFS) Readlink(name string) (string, error) {
	if t, ok := f.symlinks[name]; ok {
		return t, nil
	}
	return "", errNotSymlink
}

func (f *fakeFS) Lstat(name string) (bool, bool, error) {
	if _, ok := f.symlinks[name]; ok {
		return true, true, nil
	}
	if _, ok := f.files[name]; ok {
		return true, false, nil
	}
	return false, false, nil
}

func (f *fakeFS) MkdirAll(string) error { return nil }

func (f *fakeFS) Symlink(oldname, newname string) error {
	f.symlinks[newname] = oldname
	return nil
}

func (f *fakeFS) Rename(oldpath, newpath string) error {
	if c, ok := f.files[oldpath]; ok {
		f.files[newpath] = c
		delete(f.files, oldpath)
		return nil
	}
	if t, ok := f.symlinks[oldpath]; ok {
		f.symlinks[newpath] = t
		delete(f.symlinks, oldpath)
		return nil
	}
	return errNotSymlink
}

var errNotSymlink = fmtErrorf("not a symlink")

func fmtErrorf(s string) error { return &simpleErr{s} }

type simpleErr struct{ s string }

func (e *simpleErr) Error() string { return e.s }

func TestLinkOneIdempotent(t *testing.T) {
	fs := newFakeFS()
	fs.addFile("/repo/dotfiles/.vimrc", "x")
	fs.symlinks["/home/u/.vimrc"] = "/repo/dotfiles/.vimrc" // already linked
	l := NewLinker(fs, "/home/u", false)

	res, err := l.LinkOne("/repo/dotfiles/.vimrc", "/home/u/.vimrc")
	if err != nil {
		t.Fatalf("LinkOne: %v", err)
	}
	if res.Action != ActionAlreadyLinked {
		t.Errorf("action = %v, want ActionAlreadyLinked", res.Action)
	}
}

func TestLinkOneBacksUpConflict(t *testing.T) {
	fs := newFakeFS()
	fs.addFile("/repo/dotfiles/.vimrc", "new")
	fs.addFile("/home/u/.vimrc", "existing user file") // real file in the way
	l := NewLinker(fs, "/home/u", false)

	res, err := l.LinkOne("/repo/dotfiles/.vimrc", "/home/u/.vimrc")
	if err != nil {
		t.Fatalf("LinkOne: %v", err)
	}
	if res.Action != ActionBackedUp {
		t.Errorf("action = %v, want ActionBackedUp", res.Action)
	}
	if fs.files["/home/u/.vimrc.bak"] != "existing user file" {
		t.Errorf("backup contents = %q, want preserved original", fs.files["/home/u/.vimrc.bak"])
	}
	if fs.symlinks["/home/u/.vimrc"] != "/repo/dotfiles/.vimrc" {
		t.Errorf("symlink not created after backup")
	}
}

func TestLinkOneDryRunMakesNoChanges(t *testing.T) {
	fs := newFakeFS()
	fs.addFile("/repo/dotfiles/.vimrc", "x")
	l := NewLinker(fs, "/home/u", true) // dry-run

	res, err := l.LinkOne("/repo/dotfiles/.vimrc", "/home/u/.vimrc")
	if err != nil {
		t.Fatalf("LinkOne: %v", err)
	}
	if res.Action != ActionSkipped {
		t.Errorf("action = %v, want ActionSkipped", res.Action)
	}
	if len(fs.symlinks) != 0 {
		t.Errorf("dry-run created symlinks: %v", fs.symlinks)
	}
}
