package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zeshanziya/mcputils/internal/catalog"
)

func TestCLI_ProfileEnable(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Servers: []catalog.Server{
			{Name: "a", Command: "a-server"},
			{Name: "b", Command: "b-server"},
		},
		Profiles: []catalog.Profile{{Name: "work", Servers: []string{"a", "b"}}},
	})

	globalFlag = false
	projectFlag = ""
	rootCmd.SetArgs([]string{"--global", "profile", "enable", "work"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	claudePath := filepath.Join(home, ".claude.json")
	data, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read claude config: %v", err)
	}
	m := map[string]interface{}{}
	json.Unmarshal(data, &m)
	servers := m["mcpServers"].(map[string]interface{})
	if _, ok := servers["a"]; !ok {
		t.Fatalf("expected a enabled, got %v", servers)
	}
	if _, ok := servers["b"]; !ok {
		t.Fatalf("expected b enabled, got %v", servers)
	}
}

func TestCLI_ProfileDisable_SharedServerKept(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Servers: []catalog.Server{{Name: "shared", Command: "shared-server"}},
		Profiles: []catalog.Profile{
			{Name: "a", Servers: []string{"shared"}},
			{Name: "b", Servers: []string{"shared"}},
		},
	})

	globalFlag = false
	projectFlag = ""
	rootCmd.SetArgs([]string{"--global", "profile", "enable", "a"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("enable a failed: %v", err)
	}
	globalFlag = false
	projectFlag = ""
	rootCmd.SetArgs([]string{"--global", "profile", "enable", "b"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("enable b failed: %v", err)
	}
	globalFlag = false
	projectFlag = ""
	rootCmd.SetArgs([]string{"--global", "profile", "disable", "a"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("disable a failed: %v", err)
	}

	claudePath := filepath.Join(home, ".claude.json")
	data, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read claude config: %v", err)
	}
	m := map[string]interface{}{}
	json.Unmarshal(data, &m)
	servers := m["mcpServers"].(map[string]interface{})
	if _, ok := servers["shared"]; !ok {
		t.Fatalf("expected shared server still enabled, got %v", servers)
	}
}
