package pkgmgr

import (
	"errors"
	"reflect"
	"testing"
)

// fakeRunner records the commands it is asked to run and returns canned
// results. It stands in for the os/exec boundary so we can assert on the
// exact commands the manager builds.
type fakeRunner struct {
	calls       [][]string
	silentCalls [][]string
	runErr      error
	okOnRun     bool // when false, Run/RunSilent return runErr
}

func (f *fakeRunner) Run(name string, args ...string) error {
	f.calls = append(f.calls, append([]string{name}, args...))
	if f.okOnRun {
		return nil
	}
	return f.runErr
}

func (f *fakeRunner) RunSilent(name string, args ...string) error {
	f.silentCalls = append(f.silentCalls, append([]string{name}, args...))
	if f.okOnRun {
		return nil
	}
	return f.runErr
}

func (f *fakeRunner) lastCall() []string {
	if len(f.calls) == 0 {
		return nil
	}
	return f.calls[len(f.calls)-1]
}

func TestBrewName(t *testing.T) {
	b := NewBrew(&fakeRunner{okOnRun: true})
	if b.Name() != "brew" {
		t.Errorf("Name() = %q, want %q", b.Name(), "brew")
	}
}

func TestBrewInstallFormula(t *testing.T) {
	r := &fakeRunner{okOnRun: true}
	b := NewBrew(r)

	if err := b.Install(Package{ID: "ripgrep"}); err != nil {
		t.Fatalf("Install: unexpected error: %v", err)
	}

	want := []string{"brew", "install", "ripgrep"}
	if got := r.lastCall(); !reflect.DeepEqual(got, want) {
		t.Errorf("command = %v, want %v", got, want)
	}
}

func TestBrewInstallCask(t *testing.T) {
	r := &fakeRunner{okOnRun: true}
	b := NewBrew(r)

	if err := b.Install(Package{ID: "firefox", Cask: true}); err != nil {
		t.Fatalf("Install: unexpected error: %v", err)
	}

	want := []string{"brew", "install", "--cask", "firefox"}
	if got := r.lastCall(); !reflect.DeepEqual(got, want) {
		t.Errorf("command = %v, want %v", got, want)
	}
}

func TestBrewInstallPropagatesError(t *testing.T) {
	sentinel := errors.New("boom")
	r := &fakeRunner{okOnRun: false, runErr: sentinel}
	b := NewBrew(r)

	if err := b.Install(Package{ID: "ripgrep"}); !errors.Is(err, sentinel) {
		t.Errorf("Install error = %v, want wrap of %v", err, sentinel)
	}
}

func TestBrewIsInstalledTrue(t *testing.T) {
	r := &fakeRunner{okOnRun: true}
	b := NewBrew(r)

	got, err := b.IsInstalled("ripgrep")
	if err != nil {
		t.Fatalf("IsInstalled: unexpected error: %v", err)
	}
	if !got {
		t.Errorf("IsInstalled = false, want true")
	}
	// The check must run silently (no streamed output to the user).
	want := []string{"brew", "list", "--versions", "ripgrep"}
	if len(r.silentCalls) != 1 || !reflect.DeepEqual(r.silentCalls[0], want) {
		t.Errorf("silent calls = %v, want one %v", r.silentCalls, want)
	}
	if len(r.calls) != 0 {
		t.Errorf("IsInstalled streamed output via Run: %v", r.calls)
	}
}

func TestBrewIsInstalledFalse(t *testing.T) {
	// A non-zero exit from `brew list` means the package is absent, which is
	// not an error condition for IsInstalled.
	r := &fakeRunner{okOnRun: false, runErr: errors.New("exit status 1")}
	b := NewBrew(r)

	got, err := b.IsInstalled("ripgrep")
	if err != nil {
		t.Fatalf("IsInstalled: unexpected error: %v", err)
	}
	if got {
		t.Errorf("IsInstalled = true, want false")
	}
}
