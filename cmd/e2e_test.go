package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeshanziya/mcputils/internal/catalog"
)

func exec(t *testing.T, args ...string) (string, error) {
	t.Helper()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	globalFlag = false
	projectFlag = ""
	err := rootCmd.Execute()
	return buf.String(), err
}

func readJSONFile(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal %s: %v", path, err)
	}
	return m
}

func TestE2E_ProfileLifecycle(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Servers: []catalog.Server{{Name: "a", Command: "a-server"}, {Name: "b", Command: "b-server"}},
	})

	if _, err := exec(t, "--global", "profile", "create", "work"); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := exec(t, "--global", "profile", "add-server", "work", "a"); err != nil {
		t.Fatalf("add-server a failed: %v", err)
	}
	if _, err := exec(t, "--global", "profile", "add-server", "work", "b"); err != nil {
		t.Fatalf("add-server b failed: %v", err)
	}
	if _, err := exec(t, "--global", "profile", "enable", "work"); err != nil {
		t.Fatalf("enable failed: %v", err)
	}

	claudePath := filepath.Join(home, ".claude.json")
	m := readJSONFile(t, claudePath)
	servers := m["mcpServers"].(map[string]interface{})
	if _, ok := servers["a"]; !ok {
		t.Fatalf("expected a enabled, got %v", servers)
	}
	if _, ok := servers["b"]; !ok {
		t.Fatalf("expected b enabled, got %v", servers)
	}

	if _, err := exec(t, "--global", "profile", "disable", "work"); err != nil {
		t.Fatalf("disable failed: %v", err)
	}
	m = readJSONFile(t, claudePath)
	servers = m["mcpServers"].(map[string]interface{})
	if _, ok := servers["a"]; ok {
		t.Fatalf("expected a disabled, got %v", servers)
	}
	if _, ok := servers["b"]; ok {
		t.Fatalf("expected b disabled, got %v", servers)
	}
}

func TestE2E_SharedServerAcrossProfiles(t *testing.T) {
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

	claudePath := filepath.Join(home, ".claude.json")

	exec(t, "--global", "profile", "enable", "a")
	exec(t, "--global", "profile", "enable", "b")
	exec(t, "--global", "profile", "disable", "a")

	m := readJSONFile(t, claudePath)
	servers := m["mcpServers"].(map[string]interface{})
	if _, ok := servers["shared"]; !ok {
		t.Fatal("expected shared server still enabled after disabling only one profile")
	}

	exec(t, "--global", "profile", "disable", "b")
	m = readJSONFile(t, claudePath)
	servers = m["mcpServers"].(map[string]interface{})
	if _, ok := servers["shared"]; ok {
		t.Fatal("expected shared server disabled after disabling both profiles")
	}
}

func TestE2E_IndividualOverrideThenProfileReEnable(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Servers:  []catalog.Server{{Name: "fs", Command: "fs-server"}},
		Profiles: []catalog.Profile{{Name: "work", Servers: []string{"fs"}}},
	})
	claudePath := filepath.Join(home, ".claude.json")

	exec(t, "--global", "profile", "enable", "work")
	exec(t, "--global", "server", "disable", "fs")

	m := readJSONFile(t, claudePath)
	servers := m["mcpServers"].(map[string]interface{})
	if _, ok := servers["fs"]; ok {
		t.Fatal("expected fs disabled after individual disable")
	}

	exec(t, "--global", "profile", "enable", "work")
	m = readJSONFile(t, claudePath)
	servers = m["mcpServers"].(map[string]interface{})
	if _, ok := servers["fs"]; !ok {
		t.Fatal("expected fs re-enabled by profile enable")
	}
}

func TestE2E_GlobalVsProjectIsolation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectA := t.TempDir()
	projectB := t.TempDir()

	writeCatalogFixture(t, filepath.Join(home, ".config", "mcputils", "servers.yaml"), &catalog.Catalog{
		Servers: []catalog.Server{{Name: "gs", Command: "global-server"}},
	})
	writeCatalogFixture(t, filepath.Join(projectA, ".mcputils", "servers.yaml"), &catalog.Catalog{
		Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}},
	})
	writeCatalogFixture(t, filepath.Join(projectB, ".mcputils", "servers.yaml"), &catalog.Catalog{
		Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}},
	})

	exec(t, "--project", projectA, "server", "disable", "fs")
	exec(t, "--global", "server", "disable", "gs")

	claudePath := filepath.Join(home, ".claude.json")
	m := readJSONFile(t, claudePath)
	projects := m["projects"].(map[string]interface{})

	projA := projects[projectA].(map[string]interface{})
	arrA := projA["disabledMcpServers"].([]interface{})
	if len(arrA) != 1 || arrA[0] != "fs" {
		t.Fatalf("expected project A disabled array to contain fs, got %v", arrA)
	}

	if _, ok := projects[projectB]; ok {
		t.Fatalf("expected project B untouched, got %v", projects[projectB])
	}

	servers := m["mcpServers"].(map[string]interface{})
	if _, ok := servers["gs"]; ok {
		t.Fatalf("expected gs entry removed from top-level mcpServers, got %v", servers)
	}
	if _, ok := servers["fs"]; ok {
		t.Fatalf("expected top-level mcpServers never touched by project-scope disable, got %v", servers)
	}
}

func TestE2E_ConfigFileUnrelatedFieldsPreserved(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}},
	})
	claudePath := filepath.Join(home, ".claude.json")
	writeFileAt(t, claudePath, `{"someSetting":42,"mcpServers":{"other":{"command":"y"}},"projects":{"/other":{"disabledMcpServers":["keep"]}}}`)

	project := t.TempDir()
	writeCatalogFixture(t, filepath.Join(project, ".mcputils", "servers.yaml"), &catalog.Catalog{
		Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}},
	})

	exec(t, "--project", project, "server", "disable", "fs")
	exec(t, "--project", project, "server", "enable", "fs")
	exec(t, "--global", "server", "disable", "fs")
	exec(t, "--global", "server", "enable", "fs")

	m := readJSONFile(t, claudePath)
	if m["someSetting"] != float64(42) {
		t.Fatalf("expected someSetting preserved, got %v", m["someSetting"])
	}
	servers := m["mcpServers"].(map[string]interface{})
	if _, ok := servers["other"]; !ok {
		t.Fatalf("expected other server preserved, got %v", servers)
	}
	projects := m["projects"].(map[string]interface{})
	other := projects["/other"].(map[string]interface{})
	arr := other["disabledMcpServers"].([]interface{})
	if len(arr) != 1 || arr[0] != "keep" {
		t.Fatalf("expected /other untouched, got %v", arr)
	}
}

func TestE2E_CatalogDuplicateNameRejected(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeFileAt(t, catPath, `
servers:
  - name: fs
    command: fs-server
  - name: fs
    command: other-server
`)

	_, err := exec(t, "--global", "list", "servers")
	if err == nil {
		t.Fatal("expected error for duplicate catalog server name")
	}
}

func TestE2E_RemoveServerStillReferencedRejected(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Servers:  []catalog.Server{{Name: "fs", Command: "fs-server"}},
		Profiles: []catalog.Profile{{Name: "work", Servers: []string{"fs"}}},
	})

	_, err := exec(t, "--global", "profile", "remove-server", "work", "fs")
	if err != nil {
		t.Fatalf("remove-server from profile should succeed on its own: %v", err)
	}

	// Re-add then attempt catalog-level removal via RemoveServer semantics:
	// profile still references fs after re-adding.
	exec(t, "--global", "profile", "add-server", "work", "fs")

	cat, err := catalog.Load(catPath)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	before, _ := os.ReadFile(catPath)

	err = cat.RemoveServer("fs")
	if err == nil {
		t.Fatal("expected error removing server still referenced by profile")
	}
	if !strings.Contains(err.Error(), "work") {
		t.Fatalf("expected error to name 'work', got %v", err)
	}

	after, _ := os.ReadFile(catPath)
	if string(before) != string(after) {
		t.Fatal("expected catalog file on disk unchanged since RemoveServer was rejected in-memory")
	}
}

func TestE2E_GlobalEnableReEnablesEverywhere(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectA := t.TempDir()
	projectB := t.TempDir()

	writeCatalogFixture(t, filepath.Join(home, ".config", "mcputils", "servers.yaml"), &catalog.Catalog{
		Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}},
	})
	writeCatalogFixture(t, filepath.Join(projectA, ".mcputils", "servers.yaml"), &catalog.Catalog{
		Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}},
	})
	writeCatalogFixture(t, filepath.Join(projectB, ".mcputils", "servers.yaml"), &catalog.Catalog{
		Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}},
	})

	exec(t, "--project", projectA, "server", "disable", "fs")
	exec(t, "--project", projectB, "server", "disable", "fs")
	exec(t, "--global", "server", "disable", "fs")
	exec(t, "--global", "server", "enable", "fs")

	claudePath := filepath.Join(home, ".claude.json")
	m := readJSONFile(t, claudePath)
	servers := m["mcpServers"].(map[string]interface{})
	entry, ok := servers["fs"].(map[string]interface{})
	if !ok || entry["command"] != "fs-server" {
		t.Fatalf("expected fs re-created from catalog, got %v", servers)
	}

	projects := m["projects"].(map[string]interface{})
	for _, root := range []string{projectA, projectB} {
		proj := projects[root].(map[string]interface{})
		arr := proj["disabledMcpServers"].([]interface{})
		for _, v := range arr {
			if v == "fs" {
				t.Fatalf("expected fs swept from %s, got %v", root, arr)
			}
		}
	}
}

func TestE2E_HistoryLogRecordsProfileAndServerActions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Servers:  []catalog.Server{{Name: "a", Command: "a-server"}, {Name: "b", Command: "b-server"}},
		Profiles: []catalog.Profile{{Name: "work", Servers: []string{"a", "b"}}},
	})

	exec(t, "--global", "profile", "enable", "work")
	exec(t, "--global", "server", "disable", "a")

	out, err := exec(t, "log")
	if err != nil {
		t.Fatalf("log failed: %v", err)
	}

	for _, want := range []string{"a", "b", "work", "server", "profile:work"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected log output to contain %q, got %q", want, out)
		}
	}
	lineCount := strings.Count(strings.TrimSpace(out), "\n") + 1
	if lineCount != 3 {
		t.Fatalf("expected 3 log entries, got %d lines: %q", lineCount, out)
	}
}

func TestE2E_HistoryLogRecordsGlobalSweep(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectA := t.TempDir()
	projectB := t.TempDir()

	writeCatalogFixture(t, filepath.Join(home, ".config", "mcputils", "servers.yaml"), &catalog.Catalog{
		Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}},
	})
	writeCatalogFixture(t, filepath.Join(projectA, ".mcputils", "servers.yaml"), &catalog.Catalog{
		Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}},
	})
	writeCatalogFixture(t, filepath.Join(projectB, ".mcputils", "servers.yaml"), &catalog.Catalog{
		Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}},
	})

	exec(t, "--project", projectA, "server", "disable", "fs")
	exec(t, "--project", projectB, "server", "disable", "fs")
	exec(t, "--global", "server", "enable", "fs")

	out, err := exec(t, "log")
	if err != nil {
		t.Fatalf("log failed: %v", err)
	}
	if !strings.Contains(out, "swept_projects=2") {
		t.Fatalf("expected swept_projects=2 in log output, got %q", out)
	}
}
