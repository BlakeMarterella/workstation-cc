package installer

import (
	"io"
	"testing"

	"github.com/BlakeMarterella/workstation-cc/internal/packages"
	"github.com/BlakeMarterella/workstation-cc/internal/pkgmgr"
)

func testCatalog() *packages.Catalog {
	return &packages.Catalog{Groups: map[string]packages.Group{
		"core": {Name: "core", Packages: []packages.Package{
			{Name: "git", Managers: map[string]string{"brew": "git", "apt": "git"}},
			{Name: "ripgrep", Managers: map[string]string{"brew": "ripgrep", "apt": "ripgrep"}},
		}},
		"gui": {Name: "gui", Packages: []packages.Package{
			{Name: "firefox", Cask: true, Managers: map[string]string{"brew": "firefox"}},
		}},
	}}
}

func TestBuildPlanResolvesForManager(t *testing.T) {
	cat := testCatalog()
	items, err := BuildPlan(cat, []string{"core"}, "brew")
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Name != "git" || items[0].Pkg.ID != "git" || !items[0].Available {
		t.Errorf("items[0] = %+v, want git/available", items[0])
	}
}

func TestBuildPlanMarksUnavailable(t *testing.T) {
	cat := testCatalog()
	// firefox is gui/cask, only mapped for brew — unavailable on apt.
	items, err := BuildPlan(cat, []string{"gui"}, "apt")
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].Available {
		t.Errorf("firefox should be unavailable on apt")
	}
}

func TestBuildPlanDedupesAcrossGroups(t *testing.T) {
	cat := &packages.Catalog{Groups: map[string]packages.Group{
		"a": {Name: "a", Packages: []packages.Package{{Name: "git", Managers: map[string]string{"brew": "git"}}}},
		"b": {Name: "b", Packages: []packages.Package{{Name: "git", Managers: map[string]string{"brew": "git"}}}},
	}}
	items, err := BuildPlan(cat, []string{"a", "b"}, "brew")
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d, want 1 (deduped)", len(items))
	}
}

func TestBuildPlanUnknownGroup(t *testing.T) {
	if _, err := BuildPlan(testCatalog(), []string{"ghost"}, "brew"); err == nil {
		t.Error("expected error for unknown group, got nil")
	}
}

// fakeManager records Install calls and reports preset installed state.
type fakeManager struct {
	installed    map[string]bool
	installCalls []string
}

func (f *fakeManager) Name() string { return "fake" }
func (f *fakeManager) IsInstalled(id string) (bool, error) {
	return f.installed[id], nil
}
func (f *fakeManager) Install(p pkgmgr.Package) error {
	f.installCalls = append(f.installCalls, p.ID)
	return nil
}

func TestExecuteInstallsMissingSkipsPresent(t *testing.T) {
	items := []Item{
		{Name: "git", Pkg: pkgmgr.Package{ID: "git"}, Available: true},
		{Name: "ripgrep", Pkg: pkgmgr.Package{ID: "ripgrep"}, Available: true},
		{Name: "firefox", Available: false},
	}
	mgr := &fakeManager{installed: map[string]bool{"git": true}}

	res, err := Execute(items, mgr, false, io.Discard)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(mgr.installCalls) != 1 || mgr.installCalls[0] != "ripgrep" {
		t.Errorf("install calls = %v, want [ripgrep]", mgr.installCalls)
	}
	if len(res.Installed) != 1 || res.Installed[0] != "ripgrep" {
		t.Errorf("Installed = %v, want [ripgrep]", res.Installed)
	}
	if len(res.Skipped) != 1 || res.Skipped[0] != "git" {
		t.Errorf("Skipped = %v, want [git]", res.Skipped)
	}
	if len(res.Unavailable) != 1 || res.Unavailable[0] != "firefox" {
		t.Errorf("Unavailable = %v, want [firefox]", res.Unavailable)
	}
}

func TestExecuteDryRunDoesNotInstall(t *testing.T) {
	items := []Item{
		{Name: "ripgrep", Pkg: pkgmgr.Package{ID: "ripgrep"}, Available: true},
	}
	mgr := &fakeManager{installed: map[string]bool{}}

	res, err := Execute(items, mgr, true, io.Discard)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(mgr.installCalls) != 0 {
		t.Errorf("dry-run made install calls: %v", mgr.installCalls)
	}
	if len(res.Installed) != 1 {
		t.Errorf("Installed (would-install) = %v, want [ripgrep]", res.Installed)
	}
}
