package dotfiles

import (
	"errors"
	"reflect"
	"testing"
)

// fakeEnv is a scriptable Env for testing the bootstrap decision logic without
// touching the real yadm binary or filesystem.
type fakeEnv struct {
	hasYadm  bool
	statusOK bool // whether `yadm status` reports an existing repo
	cloneErr error
	runCalls [][]string
}

func (f *fakeEnv) HasCommand(name string) bool { return name == "yadm" && f.hasYadm }

func (f *fakeEnv) RunOK(name string, args ...string) bool {
	if name == "yadm" && len(args) == 1 && args[0] == "status" {
		return f.statusOK
	}
	return false
}

func (f *fakeEnv) Run(name string, args ...string) error {
	f.runCalls = append(f.runCalls, append([]string{name}, args...))
	return f.cloneErr
}

const repo = "https://github.com/example/dotfiles"

func TestCloneFresh(t *testing.T) {
	env := &fakeEnv{hasYadm: true, statusOK: false}
	b := New(env)

	status, err := b.Clone(repo)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	if status != StatusCloned {
		t.Errorf("status = %v, want StatusCloned", status)
	}
	want := []string{"yadm", "clone", repo}
	if len(env.runCalls) != 1 || !reflect.DeepEqual(env.runCalls[0], want) {
		t.Errorf("run calls = %v, want one %v", env.runCalls, want)
	}
}

func TestCloneSkipsWhenAlreadyPresent(t *testing.T) {
	env := &fakeEnv{hasYadm: true, statusOK: true}
	b := New(env)

	status, err := b.Clone(repo)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	if status != StatusAlreadyPresent {
		t.Errorf("status = %v, want StatusAlreadyPresent", status)
	}
	if len(env.runCalls) != 0 {
		t.Errorf("expected no clone, got %v", env.runCalls)
	}
}

func TestCloneErrorsWhenYadmMissing(t *testing.T) {
	env := &fakeEnv{hasYadm: false}
	b := New(env)

	if _, err := b.Clone(repo); err == nil {
		t.Error("expected error when yadm missing, got nil")
	}
}

func TestClonePropagatesCloneError(t *testing.T) {
	sentinel := errors.New("network down")
	env := &fakeEnv{hasYadm: true, statusOK: false, cloneErr: sentinel}
	b := New(env)

	if _, err := b.Clone(repo); !errors.Is(err, sentinel) {
		t.Errorf("err = %v, want wrap of %v", err, sentinel)
	}
}

func TestCloneEmptyRepo(t *testing.T) {
	env := &fakeEnv{hasYadm: true}
	b := New(env)

	if _, err := b.Clone(""); err == nil {
		t.Error("expected error for empty repo URL, got nil")
	}
}
