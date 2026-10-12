package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeshanziya/mcputils/internal/catalog"
)

func writeRawJSON(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("setup mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}
}

func TestCLI_Import_AddsNewServers(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeRawJSON(t, filepath.Join(home, ".claude.json"), `{"mcpServers":{"fs":{"command":"fs-server"}}}`)

	out, err := exec(t, "--global", "import")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if !strings.Contains(out, "fs") {
		t.Fatalf("expected output to mention fs as imported, got %q", out)
	}

	cat, err := catalog.Load(filepath.Join(home, ".config", "mcputils", "servers.yaml"))
	if err != nil {
		t.Fatalf("failed to load catalog: %v", err)
	}
	s, ok := cat.FindServer("fs")
	if !ok || s.Command != "fs-server" {
		t.Fatalf("expected fs added to catalog with command fs-server, got %v", cat.Servers)
	}
}

func TestCLI_Import_SkipsExisting(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Servers: []catalog.Server{{Name: "fs", Command: "existing-command"}},
	})
	writeRawJSON(t, filepath.Join(home, ".claude.json"), `{"mcpServers":{"fs":{"command":"from-source"}}}`)

	out, err := exec(t, "--global", "import")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if !strings.Contains(out, "already in catalog") {
		t.Fatalf("expected output to mention skip reason, got %q", out)
	}

	cat, err := catalog.Load(catPath)
	if err != nil {
		t.Fatalf("failed to load catalog: %v", err)
	}
	s, _ := cat.FindServer("fs")
	if s.Command != "existing-command" {
		t.Fatalf("expected existing catalog entry untouched, got %v", s)
	}
}

func TestCLI_Import_HTTPServer_SavesCorrectCatalogYAML(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeRawJSON(t, filepath.Join(home, ".claude.json"), `{"mcpServers":{"aihq":{"type":"http","url":"https://aihq.example.com/mcp","headers":{"Authorization":"Bearer xyz"}}}}`)

	out, err := exec(t, "--global", "import")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if !strings.Contains(out, "aihq") {
		t.Fatalf("expected output to mention aihq as imported, got %q", out)
	}

	cat, err := catalog.Load(filepath.Join(home, ".config", "mcputils", "servers.yaml"))
	if err != nil {
		t.Fatalf("failed to load catalog: %v", err)
	}
	s, ok := cat.FindServer("aihq")
	if !ok || !s.IsHTTP() || s.URL != "https://aihq.example.com/mcp" {
		t.Fatalf("expected aihq added as http server, got %v", cat.Servers)
	}
	if s.Command != "" {
		t.Fatalf("expected no command set for http server, got %q", s.Command)
	}
}

func TestCLI_Import_NothingToImport(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	out, err := exec(t, "--global", "import")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if !strings.Contains(out, "nothing to import") {
		t.Fatalf("expected output to indicate nothing to import, got %q", out)
	}
}
