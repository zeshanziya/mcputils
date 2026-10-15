package scope

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolve_GlobalFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cwd := t.TempDir()

	sc, err := Resolve(true, "", cwd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !sc.IsGlobal {
		t.Fatalf("expected IsGlobal true, got %+v", sc)
	}
	expectedCatalog := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	if sc.CatalogPath != expectedCatalog {
		t.Fatalf("expected CatalogPath %s, got %s", expectedCatalog, sc.CatalogPath)
	}
	expectedState := filepath.Join(home, ".config", "mcputils", "state", "global.yaml")
	if sc.StatePath != expectedState {
		t.Fatalf("expected StatePath %s, got %s", expectedState, sc.StatePath)
	}
}

func TestResolve_ProjectAutoDetect_MCPJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
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
	expectedCatalog := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	if sc.CatalogPath != expectedCatalog {
		t.Fatalf("expected CatalogPath %s (same as global), got %s", expectedCatalog, sc.CatalogPath)
	}
	expectedState := filepath.Join(home, ".config", "mcputils", "state", stateFileName(root))
	if sc.StatePath != expectedState {
		t.Fatalf("expected StatePath %s, got %s", expectedState, sc.StatePath)
	}
}

func TestResolve_ProjectAutoDetect_ClaudeJSONProjectsKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cwd := t.TempDir()

	claudeJSON := `{"projects":{"` + cwd + `":{}}}`
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), []byte(claudeJSON), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	sc, err := Resolve(false, "", cwd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sc.IsGlobal {
		t.Fatalf("expected IsGlobal false, got %+v", sc)
	}
	if sc.ProjectRoot != cwd {
		t.Fatalf("expected ProjectRoot %s, got %s", cwd, sc.ProjectRoot)
	}
}

func TestResolve_ProjectNotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := t.TempDir()
	_, err := Resolve(false, "", root)
	if err == nil {
		t.Fatal("expected error when neither .mcp.json nor a matching ~/.claude.json projects key is found")
	}
}

func TestResolve_NoWalkUp_NestedCwdErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".mcp.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err := Resolve(false, "", nested)
	if err == nil {
		t.Fatal("expected error — auto-detect no longer walks up to find the parent's .mcp.json")
	}
}

func TestResolve_ProjectExplicitPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
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
	expectedCatalog := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	if sc.CatalogPath != expectedCatalog {
		t.Fatalf("expected CatalogPath %s, got %s", expectedCatalog, sc.CatalogPath)
	}
	if strings.Contains(sc.CatalogPath, ".mcputils") {
		t.Fatalf("expected no per-project .mcputils dir in CatalogPath, got %s", sc.CatalogPath)
	}
	expectedState := filepath.Join(home, ".config", "mcputils", "state", stateFileName(explicit))
	if sc.StatePath != expectedState {
		t.Fatalf("expected StatePath %s, got %s", expectedState, sc.StatePath)
	}
}

func TestResolve_ConflictingFlagsErrors(t *testing.T) {
	_, err := Resolve(true, "/x", "/irrelevant")
	if err == nil {
		t.Fatal("expected error for conflicting --global and --project flags")
	}
}

func TestResolve_CatalogPathIdenticalAcrossScopes(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	global, err := Resolve(true, "", "/irrelevant")
	if err != nil {
		t.Fatalf("global Resolve failed: %v", err)
	}
	projectA, err := Resolve(false, "/project/a", "/irrelevant")
	if err != nil {
		t.Fatalf("project A Resolve failed: %v", err)
	}
	projectB, err := Resolve(false, "/project/b", "/irrelevant")
	if err != nil {
		t.Fatalf("project B Resolve failed: %v", err)
	}

	if global.CatalogPath != projectA.CatalogPath || projectA.CatalogPath != projectB.CatalogPath {
		t.Fatalf("expected identical CatalogPath across scopes, got %s, %s, %s", global.CatalogPath, projectA.CatalogPath, projectB.CatalogPath)
	}
}

func TestResolve_StatePathDiffersPerProjectRoot(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	global, err := Resolve(true, "", "/irrelevant")
	if err != nil {
		t.Fatalf("global Resolve failed: %v", err)
	}
	projectA, err := Resolve(false, "/project/a", "/irrelevant")
	if err != nil {
		t.Fatalf("project A Resolve failed: %v", err)
	}
	projectB, err := Resolve(false, "/project/b", "/irrelevant")
	if err != nil {
		t.Fatalf("project B Resolve failed: %v", err)
	}

	if global.StatePath == projectA.StatePath || projectA.StatePath == projectB.StatePath || global.StatePath == projectB.StatePath {
		t.Fatalf("expected distinct StatePath per scope, got %s, %s, %s", global.StatePath, projectA.StatePath, projectB.StatePath)
	}
}

func TestStateFileName_Deterministic(t *testing.T) {
	a := stateFileName("/some/project")
	b := stateFileName("/some/project")
	if a != b {
		t.Fatalf("expected deterministic output, got %s vs %s", a, b)
	}
}

func TestStateFileName_DifferentRootsDifferentNames(t *testing.T) {
	a := stateFileName("/some/project")
	b := stateFileName("/some/other-project")
	if a == b {
		t.Fatalf("expected different output for different roots, got %s for both", a)
	}
}
