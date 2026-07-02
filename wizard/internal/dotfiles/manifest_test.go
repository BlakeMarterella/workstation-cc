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
	m, _ := ParseManifest([]byte(sampleManifest))
	got, ok := m.DestFor("vscode", "darwin", "/Users/u")
	want := "/Users/u/Library/Application Support/Code/User"
	if !ok || got != want {
		t.Errorf("DestFor(vscode,darwin) = %q,%v want %q,true", got, ok, want)
	}
}

func TestResolveMissingOSSkips(t *testing.T) {
	m, _ := ParseManifest([]byte(sampleManifest))
	if _, ok := m.DestFor("vscode", "windows", "C:/Users/u"); ok {
		t.Error("DestFor(vscode,windows) = ok, want not-ok (no windows dest)")
	}
}

func TestLinkAppsSkipsMissingOS(t *testing.T) {
	fs := newFakeFS()
	fs.addFile("/repo/app-configs/vscode/settings.json", "{}")
	m, _ := ParseManifest([]byte(sampleManifest))
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
