package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zeshanziya/mcputils/internal/catalog"
)

func TestCLI_ServerDisable(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}},
	})

	globalFlag = false
	projectFlag = ""
	rootCmd.SetArgs([]string{"--global", "server", "disable", "fs"})
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
	if _, ok := servers["fs"]; ok {
		t.Fatalf("expected fs entry removed (global disable), got %v", servers)
	}
}

func TestCLI_ServerEnable_UnknownName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	globalFlag = false
	projectFlag = ""
	rootCmd.SetArgs([]string{"--global", "server", "enable", "nope"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown server")
	}
}

func TestCLI_ServerEnable(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	project := t.TempDir()
	projCatPath := filepath.Join(project, ".mcputils", "servers.yaml")
	writeCatalogFixture(t, projCatPath, &catalog.Catalog{
		Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}},
	})

	globalFlag = false
	projectFlag = project
	rootCmd.SetArgs([]string{"--project", project, "server", "disable", "fs"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("setup disable failed: %v", err)
	}

	globalFlag = false
	projectFlag = project
	rootCmd.SetArgs([]string{"--project", project, "server", "enable", "fs"})
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
	projects := m["projects"].(map[string]interface{})
	proj := projects[project].(map[string]interface{})
	arr := proj["disabledMcpServers"].([]interface{})
	for _, v := range arr {
		if v == "fs" {
			t.Fatalf("expected fs removed from disabled array (project scope), got %v", arr)
		}
	}
}
