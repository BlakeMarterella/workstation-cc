// Package tui provides the interactive terminal UI: an profile picker followed
// by a filterable preview of the packages that profile will install. It is a
// selector — it returns the user's choice and the caller runs the (tested)
// install path. Long-running installs never happen inside the event loop.
//
// The TUI renders over ANSI and therefore works locally and over SSH; callers
// must only launch it when attached to an interactive terminal (see IsInteractive).
package tui

import (
	"fmt"
	"os"

	"github.com/BlakeMarterella/workstation-cc/internal/installer"
	"github.com/BlakeMarterella/workstation-cc/internal/packages"
	"github.com/BlakeMarterella/workstation-cc/internal/profiles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// IsInteractive reports whether stdin and stdout are attached to a terminal.
// The TUI must only be launched when this is true.
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// Result is what the user chose in the TUI.
type Result struct {
	Profile   string
	Confirmed bool
}

type viewState int

const (
	stateProfile viewState = iota
	statePackages
)

type profileItem struct{ name, desc string }

func (p profileItem) Title() string       { return p.name }
func (p profileItem) Description() string { return p.desc }
func (p profileItem) FilterValue() string { return p.name }

type pkgItem struct{ name, detail string }

func (p pkgItem) Title() string       { return p.name }
func (p pkgItem) Description() string { return p.detail }
func (p pkgItem) FilterValue() string { return p.name }

type model struct {
	state    viewState
	profiles list.Model
	pkgs     list.Model
	cat      *packages.Catalog
	set      *profiles.Set
	manager  string
	chosen   string
	result   Result
	width    int
	height   int
}

func newModel(cat *packages.Catalog, set *profiles.Set, manager string) model {
	items := make([]list.Item, 0, len(set.Names()))
	for _, name := range set.Names() {
		items = append(items, profileItem{name: name, desc: set.Profiles[name].Description})
	}

	pl := list.New(items, list.NewDefaultDelegate(), 0, 0)
	pl.Title = "Choose a profile"
	pl.SetShowStatusBar(false)

	return model{
		state:    stateProfile,
		profiles: pl,
		cat:      cat,
		set:      set,
		manager:  manager,
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.profiles.SetSize(msg.Width, msg.Height-1)
		// The package list is created lazily (loadPackages) using the current
		// width/height; only resize it once it exists.
		if m.state == statePackages {
			m.pkgs.SetSize(msg.Width, msg.Height-1)
		}
		return m, nil

	case tea.KeyMsg:
		// Never intercept keys while a list's filter input is active.
		switch m.state {
		case stateProfile:
			if m.profiles.FilterState() != list.Filtering {
				switch msg.String() {
				case "ctrl+c", "q":
					return m, tea.Quit
				case "enter":
					if it, ok := m.profiles.SelectedItem().(profileItem); ok {
						m.chosen = it.name
						m.loadPackages(it.name)
						m.state = statePackages
					}
					return m, nil
				}
			}
		case statePackages:
			if m.pkgs.FilterState() != list.Filtering {
				switch msg.String() {
				case "ctrl+c", "q":
					return m, tea.Quit
				case "esc":
					m.state = stateProfile
					return m, nil
				case "enter":
					m.result = Result{Profile: m.chosen, Confirmed: true}
					return m, tea.Quit
				}
			}
		}
	}

	var cmd tea.Cmd
	if m.state == stateProfile {
		m.profiles, cmd = m.profiles.Update(msg)
	} else {
		m.pkgs, cmd = m.pkgs.Update(msg)
	}
	return m, cmd
}

// loadPackages resolves the chosen profile into a filterable package list.
func (m *model) loadPackages(profile string) {
	var items []list.Item
	if groups, err := m.set.Resolve(profile); err == nil {
		if plan, err := installer.BuildPlan(m.cat, groups, m.manager); err == nil {
			for _, it := range plan {
				detail := fmt.Sprintf("%s: %s", m.manager, it.Pkg.ID)
				if !it.Available {
					detail = "unavailable for " + m.manager
				}
				items = append(items, pkgItem{name: it.Name, detail: detail})
			}
		}
	}

	pl := list.New(items, list.NewDefaultDelegate(), m.width, m.height-1)
	pl.Title = fmt.Sprintf("Packages in %q  (press / to filter, enter to install, esc to go back)", profile)
	pl.SetShowStatusBar(true)
	m.pkgs = pl
}

func (m model) View() string {
	switch m.state {
	case statePackages:
		return m.pkgs.View()
	default:
		return m.profiles.View()
	}
}

// Run launches the interactive selector and returns the user's choice.
func Run(cat *packages.Catalog, set *profiles.Set, manager string) (Result, error) {
	p := tea.NewProgram(newModel(cat, set, manager), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return Result{}, err
	}
	if m, ok := final.(model); ok {
		return m.result, nil
	}
	return Result{}, nil
}
