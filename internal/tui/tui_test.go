package tui

import (
	"testing"

	"github.com/BlakeMarterella/workstation-cc/internal/packages"
	"github.com/BlakeMarterella/workstation-cc/internal/profiles"
	tea "github.com/charmbracelet/bubbletea"
)

func testModel() model {
	cat := &packages.Catalog{Groups: map[string]packages.Group{
		"core": {Name: "core", Packages: []packages.Package{
			{Name: "git", Managers: map[string]string{"brew": "git"}},
		}},
	}}
	set := &profiles.Set{Profiles: map[string]profiles.Profile{
		"slim": {Description: "Essentials", Groups: []string{"core"}},
	}}
	m := newModel(cat, set, "brew")
	// Give the lists a size, as a real terminal would.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return updated.(model)
}

func enter(m model) model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return updated.(model)
}

func TestEnterSelectsProfileAndShowsPackages(t *testing.T) {
	m := enter(testModel())
	if m.state != statePackages {
		t.Fatalf("state = %v, want statePackages", m.state)
	}
	if m.chosen != "slim" {
		t.Errorf("chosen = %q, want slim", m.chosen)
	}
	if len(m.pkgs.Items()) != 1 {
		t.Errorf("package items = %d, want 1", len(m.pkgs.Items()))
	}
}

func TestEnterOnPackagesConfirms(t *testing.T) {
	m := enter(enter(testModel())) // select profile, then confirm
	if !m.result.Confirmed {
		t.Error("result.Confirmed = false, want true")
	}
	if m.result.Profile != "slim" {
		t.Errorf("result.Profile = %q, want slim", m.result.Profile)
	}
}

func TestEscFromPackagesGoesBack(t *testing.T) {
	m := enter(testModel()) // now on packages
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if m.state != stateProfile {
		t.Errorf("state = %v, want stateProfile", m.state)
	}
	if m.result.Confirmed {
		t.Error("esc should not confirm")
	}
}
