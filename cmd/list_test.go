package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeshanziya/mcputils/internal/catalog"
	"github.com/zeshanziya/mcputils/internal/state"
)

func TestCLI_ListServers_ShowsState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Servers: []catalog.Server{
			{Name: "a", Command: "a-server"},
			{Name: "b", Command: "b-server"},
		},
	})
	claudePath := filepath.Join(home, ".claude.json")
	writeFileAt(t, claudePath, `{"mcpServers":{"a":{"command":"a-server"}}}`)

	globalFlag = false
	projectFlag = ""
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--global", "list", "servers"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatalf("expected output to list both servers, got %q", out)
	}
	if !strings.Contains(out, "enabled") || !strings.Contains(out, "disabled") {
		t.Fatalf("expected output to show enabled/disabled markers, got %q", out)
	}
}

func TestCLI_ListProfiles_ShowsActive(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Profiles: []catalog.Profile{{Name: "work"}, {Name: "idle"}},
	})
	statePath := filepath.Join(home, ".config", "mcputils", "state.yaml")
	st := &state.State{}
	st.SetActive("work", true)
	if err := st.Save(statePath); err != nil {
		t.Fatalf("setup save state failed: %v", err)
	}

	globalFlag = false
	projectFlag = ""
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--global", "list", "profiles"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "work") || !strings.Contains(out, "idle") {
		t.Fatalf("expected both profiles listed, got %q", out)
	}
	if !strings.Contains(out, "active") || !strings.Contains(out, "inactive") {
		t.Fatalf("expected active/inactive markers, got %q", out)
	}
}

func writeFileAt(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("setup mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}
}
