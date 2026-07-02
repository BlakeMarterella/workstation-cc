package dotfiles

import "testing"

const sampleManifest = `
entries:
  - name: nvim
    dest:
      default: ~/.config/nvim
  - name: vscode
    dest:
      darwin: ~/Library/Application Support/Code/User
      linux: ~/.config/Code/User
`

func TestParseEmptyManifest(t *testing.T) {
	// Test that empty input yields an empty manifest without error
	m, err := ParseManifest([]byte(""))
	if err != nil {
		t.Fatalf("ParseManifest(empty): %v", err)
	}
	if len(m.Entries) != 0 {
		t.Errorf("ParseManifest(empty) len(Entries) = %d, want 0", len(m.Entries))
	}

	// Test that whitespace-only input also yields an empty manifest
	m, err = ParseManifest([]byte("   "))
	if err != nil {
		t.Fatalf("ParseManifest(whitespace): %v", err)
	}
	if len(m.Entries) != 0 {
		t.Errorf("ParseManifest(whitespace) len(Entries) = %d, want 0", len(m.Entries))
	}
}

func TestParseAndResolveDefault(t *testing.T) {
	m, err := ParseManifest([]byte(sampleManifest))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	got, ok := m.DestFor("nvim", "linux", "/home/u")
	if !ok || got != "/home/u/.config/nvim" {
		t.Errorf("DestFor(nvim,linux) = %q,%v want /home/u/.config/nvim,true", got, ok)
	}
}

func TestResolvePerOS(t *testing.T) {
	m, err := ParseManifest([]byte(sampleManifest))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	got, ok := m.DestFor("vscode", "darwin", "/Users/u")
	want := "/Users/u/Library/Application Support/Code/User"
	if !ok || got != want {
		t.Errorf("DestFor(vscode,darwin) = %q,%v want %q,true", got, ok, want)
	}
}

func TestResolveMissingOSSkips(t *testing.T) {
	m, err := ParseManifest([]byte(sampleManifest))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if _, ok := m.DestFor("vscode", "windows", "C:/Users/u"); ok {
		t.Error("DestFor(vscode,windows) = ok, want not-ok (no windows dest)")
	}
}

func TestLinkAppsSkipsMissingOS(t *testing.T) {
	fs := newFakeFS()
	// nvim has a default dest, so it resolves on all OSes; pre-populate source so LinkOne has real entry
	fs.addFile("/repo/app-configs/nvim/init.lua", "")
	// vscode has no windows dest, so it should be ActionSkipped on windows
	fs.addFile("/repo/app-configs/vscode/settings.json", "{}")
	m, err := ParseManifest([]byte(sampleManifest))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	l := NewLinker(fs, "C:/Users/u", true)

	results, err := l.LinkApps("/repo/app-configs", m, "windows")
	if err != nil {
		t.Fatalf("LinkApps: %v", err)
	}
	for _, r := range results {
		if r.Path == "/repo/app-configs/vscode" && r.Action != ActionSkipped {
			t.Errorf("vscode on windows = %v, want ActionSkipped", r.Action)
		}
	}
}
