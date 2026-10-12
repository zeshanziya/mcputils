package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeshanziya/mcputils/internal/catalog"
	"github.com/zeshanziya/mcputils/internal/state"
)

func TestCLI_Status_CombinedView(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Servers: []catalog.Server{
			{Name: "a", Command: "a-server"},
			{Name: "b", Command: "b-server"},
		},
		Profiles: []catalog.Profile{
			{Name: "work", Servers: []string{"a"}},
			{Name: "idle", Servers: []string{"b"}},
		},
	})
	claudePath := filepath.Join(home, ".claude.json")
	writeFileAt(t, claudePath, `{"mcpServers":{"a":{"command":"a-server"}}}`)

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
	rootCmd.SetArgs([]string{"--global", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out := buf.String()
	for _, want := range []string{"work", "idle", "a", "b", "active", "inactive", "enabled", "disabled"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got %q", want, out)
		}
	}
}
