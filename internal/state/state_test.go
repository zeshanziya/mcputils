package state

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func TestSetActive_AddsOnce(t *testing.T) {
	st := &State{}
	st.SetActive("x", true)
	st.SetActive("x", true)

	count := 0
	for _, v := range st.ActiveProfiles {
		if v == "x" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected 'x' exactly once, got %v", st.ActiveProfiles)
	}
}

func TestSetActive_Removes(t *testing.T) {
	st := &State{}
	st.SetActive("x", true)
	st.SetActive("x", false)

	if contains(st.ActiveProfiles, "x") {
		t.Fatalf("expected 'x' to be removed, got %v", st.ActiveProfiles)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.yaml")

	st, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if st.ActiveProfiles != nil {
		t.Fatalf("expected nil ActiveProfiles, got %v", st.ActiveProfiles)
	}
}

func TestSaveThenLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "state.yaml")

	st := &State{}
	st.SetActive("work", true)
	if err := st.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !loaded.IsActive("work") {
		t.Fatalf("expected 'work' active after round trip, got %v", loaded.ActiveProfiles)
	}
}

func TestLoad_InvalidYAMLErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := writeFile(path, "active_profiles: [this is not valid: yaml"); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
