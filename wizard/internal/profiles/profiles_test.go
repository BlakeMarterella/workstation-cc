package profiles

import (
	"reflect"
	"testing"
)

const sampleYAML = `profiles:
  slim:
    description: Essentials only.
    groups:
      - core
  full:
    description: Everything.
    includes:
      - slim
    groups:
      - dev
      - gui
`

func TestParseSet(t *testing.T) {
	s, err := parseSet([]byte(sampleYAML))
	if err != nil {
		t.Fatalf("parseSet: %v", err)
	}
	if len(s.Profiles) != 2 {
		t.Fatalf("len(Profiles) = %d, want 2", len(s.Profiles))
	}
	if s.Profiles["slim"].Description != "Essentials only." {
		t.Errorf("slim description = %q", s.Profiles["slim"].Description)
	}
}

func TestResolveSlim(t *testing.T) {
	s, _ := parseSet([]byte(sampleYAML))
	got, err := s.Resolve("slim")
	if err != nil {
		t.Fatalf("Resolve(slim): %v", err)
	}
	want := []string{"core"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Resolve(slim) = %v, want %v", got, want)
	}
}

func TestResolveFullIncludesSlim(t *testing.T) {
	s, _ := parseSet([]byte(sampleYAML))
	got, err := s.Resolve("full")
	if err != nil {
		t.Fatalf("Resolve(full): %v", err)
	}
	// slim (core) must come first, then full's own groups, deduped.
	want := []string{"core", "dev", "gui"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Resolve(full) = %v, want %v", got, want)
	}
}

func TestResolveUnknownProfile(t *testing.T) {
	s, _ := parseSet([]byte(sampleYAML))
	if _, err := s.Resolve("nope"); err == nil {
		t.Error("expected error for unknown profile, got nil")
	}
}

func TestResolveDedupesOverlappingGroups(t *testing.T) {
	yaml := `profiles:
  a:
    groups: [core, dev]
  b:
    includes: [a]
    groups: [dev, gui]
`
	s, _ := parseSet([]byte(yaml))
	got, err := s.Resolve("b")
	if err != nil {
		t.Fatalf("Resolve(b): %v", err)
	}
	want := []string{"core", "dev", "gui"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Resolve(b) = %v, want %v", got, want)
	}
}

func TestResolveDetectsCycle(t *testing.T) {
	yaml := `profiles:
  a:
    includes: [b]
  b:
    includes: [a]
`
	s, _ := parseSet([]byte(yaml))
	if _, err := s.Resolve("a"); err == nil {
		t.Error("expected error for include cycle, got nil")
	}
}

func TestResolveUnknownInclude(t *testing.T) {
	yaml := `profiles:
  a:
    includes: [ghost]
`
	s, _ := parseSet([]byte(yaml))
	if _, err := s.Resolve("a"); err == nil {
		t.Error("expected error for unknown include, got nil")
	}
}
