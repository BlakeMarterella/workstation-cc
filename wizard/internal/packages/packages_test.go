package packages

import (
	"testing"
	"testing/fstest"
)

const coreYAML = `category: core
description: Essential CLI tools.
packages:
  - name: ripgrep
    description: Fast recursive grep
    managers:
      brew: ripgrep
      apt: ripgrep
      dnf: ripgrep
  - name: fd
    description: Fast find alternative
    managers:
      brew: fd
      apt: fd-find
`

const guiYAML = `category: gui
description: GUI apps.
packages:
  - name: firefox
    cask: true
    managers:
      brew: firefox
`

func TestParseGroup(t *testing.T) {
	g, err := parseGroup([]byte(coreYAML))
	if err != nil {
		t.Fatalf("parseGroup: %v", err)
	}
	if g.Name != "core" {
		t.Errorf("Name = %q, want core", g.Name)
	}
	if len(g.Packages) != 2 {
		t.Fatalf("len(Packages) = %d, want 2", len(g.Packages))
	}
	if g.Packages[0].Name != "ripgrep" {
		t.Errorf("Packages[0].Name = %q, want ripgrep", g.Packages[0].Name)
	}
	if g.Packages[0].Managers["apt"] != "ripgrep" {
		t.Errorf("ripgrep apt id = %q, want ripgrep", g.Packages[0].Managers["apt"])
	}
}

func TestParseGroupMissingCategory(t *testing.T) {
	if _, err := parseGroup([]byte("packages: []\n")); err == nil {
		t.Error("expected error for missing category, got nil")
	}
}

func TestLoadFS(t *testing.T) {
	fsys := fstest.MapFS{
		"core.yaml": {Data: []byte(coreYAML)},
		"gui.yaml":  {Data: []byte(guiYAML)},
	}
	cat, err := LoadFS(fsys)
	if err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	if len(cat.Groups) != 2 {
		t.Fatalf("len(Groups) = %d, want 2", len(cat.Groups))
	}
	core, ok := cat.Group("core")
	if !ok {
		t.Fatal("Group(core) not found")
	}
	if len(core.Packages) != 2 {
		t.Errorf("core packages = %d, want 2", len(core.Packages))
	}
}

func TestPackageForManager(t *testing.T) {
	p := Package{
		Name: "firefox",
		Cask: true,
		Managers: map[string]string{
			"brew": "firefox",
		},
	}

	got, ok := p.For("brew")
	if !ok {
		t.Fatal("For(brew): ok = false, want true")
	}
	if got.ID != "firefox" {
		t.Errorf("ID = %q, want firefox", got.ID)
	}
	if !got.Cask {
		t.Error("Cask = false, want true")
	}

	if _, ok := p.For("winget"); ok {
		t.Error("For(winget): ok = true, want false (no mapping)")
	}
}
