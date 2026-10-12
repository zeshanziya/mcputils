package scope

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolve_GlobalFlag(t *testing.T) {
	cwd := t.TempDir()
	sc, err := Resolve(true, "", cwd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !sc.IsGlobal {
		t.Fatalf("expected IsGlobal true, got %+v", sc)
	}
	home, _ := os.UserHomeDir()
	expectedDir := filepath.Join(home, ".config", "mcputils")
	if !strings.HasPrefix(sc.CatalogPath, expectedDir) || !strings.HasPrefix(sc.StatePath, expectedDir) {
		t.Fatalf("expected paths under %s, got %+v", expectedDir, sc)
	}
}

func TestResolve_ProjectAutoDetect(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".mcp.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	sc, err := Resolve(false, "", root)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sc.IsGlobal {
		t.Fatalf("expected IsGlobal false, got %+v", sc)
	}
	if sc.ProjectRoot != root {
		t.Fatalf("expected ProjectRoot %s, got %s", root, sc.ProjectRoot)
	}
}

func TestResolve_ProjectNotFound(t *testing.T) {
	root := t.TempDir()
	_, err := Resolve(false, "", root)
	if err == nil {
		t.Fatal("expected error when no .mcp.json found")
	}
}

func TestResolve_NestedCwd(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".mcp.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	sc, err := Resolve(false, "", nested)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sc.ProjectRoot != root {
		t.Fatalf("expected ProjectRoot %s, got %s", root, sc.ProjectRoot)
	}
}

func TestResolve_ProjectExplicitPath(t *testing.T) {
	explicit := "/some/path"
	sc, err := Resolve(false, explicit, "/irrelevant/cwd")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sc.IsGlobal {
		t.Fatalf("expected IsGlobal false, got %+v", sc)
	}
	if sc.ProjectRoot != explicit {
		t.Fatalf("expected ProjectRoot %s, got %s", explicit, sc.ProjectRoot)
	}
	expectedDir := filepath.Join(explicit, ".mcputils")
	if !strings.HasPrefix(sc.CatalogPath, expectedDir) || !strings.HasPrefix(sc.StatePath, expectedDir) {
		t.Fatalf("expected paths under %s, got %+v", expectedDir, sc)
	}
}

func TestResolve_ConflictingFlagsErrors(t *testing.T) {
	_, err := Resolve(true, "/x", "/irrelevant")
	if err == nil {
		t.Fatal("expected error for conflicting --global and --project flags")
	}
}
