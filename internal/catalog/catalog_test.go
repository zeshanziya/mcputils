package catalog

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoad_EmptyFileMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.yaml")

	cat, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cat.Servers) != 0 || len(cat.Profiles) != 0 {
		t.Fatalf("expected empty catalog, got %+v", cat)
	}
}

func TestLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.yaml")

	cat := &Catalog{
		Servers: []Server{
			{Name: "fs", Command: "fs-server", Args: []string{"--root", "/"}, Env: map[string]string{"X": "1"}},
		},
		Profiles: []Profile{
			{Name: "work", Servers: []string{"fs"}},
		},
	}
	if err := cat.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !reflect.DeepEqual(cat, loaded) {
		t.Fatalf("expected %+v, got %+v", cat, loaded)
	}
}

func TestFindServer_Found(t *testing.T) {
	cat := &Catalog{Servers: []Server{{Name: "fs", Command: "fs-server"}}}
	s, ok := cat.FindServer("fs")
	if !ok || s.Name != "fs" {
		t.Fatalf("expected to find fs, got %+v, %v", s, ok)
	}
}

func TestFindServer_NotFound(t *testing.T) {
	cat := &Catalog{Servers: []Server{{Name: "fs", Command: "fs-server"}}}
	s, ok := cat.FindServer("nope")
	if ok || s != nil {
		t.Fatalf("expected not found, got %+v, %v", s, ok)
	}
}

func TestLoad_DuplicateServerNameErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.yaml")
	content := `
servers:
  - name: fs
    command: fs-server
  - name: fs
    command: other-server
`
	writeFile(t, path, content)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate server name")
	}
}

func TestLoad_DuplicateProfileNameErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.yaml")
	content := `
profiles:
  - name: work
    servers: []
  - name: work
    servers: []
`
	writeFile(t, path, content)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate profile name")
	}
}

func TestRemoveServer_ReferencedByProfileErrors(t *testing.T) {
	cat := &Catalog{
		Servers:  []Server{{Name: "fs", Command: "fs-server"}},
		Profiles: []Profile{{Name: "work", Servers: []string{"fs"}}},
	}

	err := cat.RemoveServer("fs")
	if err == nil {
		t.Fatal("expected error when server is referenced by a profile")
	}
	if !strings.Contains(err.Error(), "work") {
		t.Fatalf("expected error to name 'work', got %v", err)
	}
	if len(cat.Servers) != 1 {
		t.Fatalf("expected catalog unchanged, got %+v", cat.Servers)
	}
}

func TestRemoveServer_UnreferencedSucceeds(t *testing.T) {
	cat := &Catalog{
		Servers:  []Server{{Name: "fs", Command: "fs-server"}},
		Profiles: []Profile{{Name: "work", Servers: []string{}}},
	}

	if err := cat.RemoveServer("fs"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, ok := cat.FindServer("fs"); ok {
		t.Fatalf("expected fs to be removed, got %+v", cat.Servers)
	}
}

func TestServer_IsHTTP_TypeFieldSet(t *testing.T) {
	s := Server{Type: "http"}
	if !s.IsHTTP() {
		t.Fatal("expected IsHTTP true when Type is http")
	}
}

func TestServer_IsHTTP_URLSetNoType(t *testing.T) {
	s := Server{URL: "https://example.com/mcp"}
	if !s.IsHTTP() {
		t.Fatal("expected IsHTTP true when URL is set")
	}
}

func TestServer_IsHTTP_StdioServerFalse(t *testing.T) {
	s := Server{Command: "fs-server"}
	if s.IsHTTP() {
		t.Fatal("expected IsHTTP false for a stdio server")
	}
}

func TestLoad_RoundTrip_HTTPServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.yaml")

	cat := &Catalog{
		Servers: []Server{
			{Name: "aihq", Type: "http", URL: "${AIHQ_URL}/mcp/x", Headers: map[string]string{"Authorization": "${AIHQ_TOKEN}"}},
		},
	}
	if err := cat.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !reflect.DeepEqual(cat.Servers, loaded.Servers) {
		t.Fatalf("expected round-trip to match, got %+v vs %+v", cat.Servers, loaded.Servers)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write fixture file: %v", err)
	}
}
