package claudecode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zeshanziya/mcputils/internal/catalog"
	"github.com/zeshanziya/mcputils/internal/scope"
)

func newAdapter(t *testing.T, cat *catalog.Catalog) (*Adapter, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "claude.json")
	return &Adapter{cat: cat, path: path}, path
}

func writeConfig(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("setup mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}
}

func readConfigForTest(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}
	return m
}

func projectScope(root string) scope.Scope {
	return scope.Scope{IsGlobal: false, ProjectRoot: root}
}

func globalScope() scope.Scope {
	return scope.Scope{IsGlobal: true}
}

func TestSetEnabled_Project_AddsToDisabledArray(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{})
	writeConfig(t, path, `{"mcpServers":{"fs":{"command":"fs-server"}}}`)

	if _, err := a.SetEnabled(projectScope("/proj"), "fs", false); err != nil {
		t.Fatalf("SetEnabled failed: %v", err)
	}

	m := readConfigForTest(t, path)
	projects := m["projects"].(map[string]interface{})
	proj := projects["/proj"].(map[string]interface{})
	arr := proj["disabledMcpServers"].([]interface{})
	if len(arr) != 1 || arr[0] != "fs" {
		t.Fatalf("expected disabledMcpServers to contain fs, got %v", arr)
	}
	servers := m["mcpServers"].(map[string]interface{})
	if _, ok := servers["fs"]; !ok {
		t.Fatalf("expected top-level mcpServers.fs untouched, got %v", servers)
	}
}

func TestSetEnabled_Project_RemovesFromDisabledArray(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{})
	writeConfig(t, path, `{"projects":{"/proj":{"disabledMcpServers":["fs"]}}}`)

	if _, err := a.SetEnabled(projectScope("/proj"), "fs", true); err != nil {
		t.Fatalf("SetEnabled failed: %v", err)
	}

	m := readConfigForTest(t, path)
	projects := m["projects"].(map[string]interface{})
	proj := projects["/proj"].(map[string]interface{})
	arr := proj["disabledMcpServers"].([]interface{})
	for _, v := range arr {
		if v == "fs" {
			t.Fatalf("expected fs removed, got %v", arr)
		}
	}
}

func TestSetEnabled_Project_OtherProjectsUntouched(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{})
	writeConfig(t, path, `{"projects":{"/other":{"disabledMcpServers":["keep"]}},"mcpServers":{"other-server":{"command":"x"}}}`)

	if _, err := a.SetEnabled(projectScope("/proj"), "fs", false); err != nil {
		t.Fatalf("SetEnabled failed: %v", err)
	}

	m := readConfigForTest(t, path)
	projects := m["projects"].(map[string]interface{})
	other := projects["/other"].(map[string]interface{})
	arr := other["disabledMcpServers"].([]interface{})
	if len(arr) != 1 || arr[0] != "keep" {
		t.Fatalf("expected /other untouched, got %v", arr)
	}
	servers := m["mcpServers"].(map[string]interface{})
	if _, ok := servers["other-server"]; !ok {
		t.Fatalf("expected other-server untouched, got %v", servers)
	}
}

func TestIsEnabled_Project_ReflectsArrayMembership(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{})
	writeConfig(t, path, `{"projects":{"/proj":{"disabledMcpServers":["fs"]}}}`)

	enabled, err := a.IsEnabled(projectScope("/proj"), "fs")
	if err != nil {
		t.Fatalf("IsEnabled failed: %v", err)
	}
	if enabled {
		t.Fatal("expected fs to be disabled")
	}

	enabled, err = a.IsEnabled(projectScope("/proj"), "other")
	if err != nil {
		t.Fatalf("IsEnabled failed: %v", err)
	}
	if !enabled {
		t.Fatal("expected other to be enabled")
	}
}

func TestSetEnabled_Global_Disable_RemovesEntry(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{})
	writeConfig(t, path, `{"mcpServers":{"fs":{"command":"fs-server"}},"projects":{"/proj":{"disabledMcpServers":[]}}}`)

	if _, err := a.SetEnabled(globalScope(), "fs", false); err != nil {
		t.Fatalf("SetEnabled failed: %v", err)
	}

	m := readConfigForTest(t, path)
	servers := m["mcpServers"].(map[string]interface{})
	if _, ok := servers["fs"]; ok {
		t.Fatalf("expected fs entry removed, got %v", servers)
	}
}

func TestSetEnabled_Global_Enable_RecreatesEntryFromCatalog(t *testing.T) {
	cat := &catalog.Catalog{Servers: []catalog.Server{
		{Name: "fs", Command: "fs-server", Args: []string{"--root"}, Env: map[string]string{"X": "1"}},
	}}
	a, path := newAdapter(t, cat)
	writeConfig(t, path, `{}`)

	if _, err := a.SetEnabled(globalScope(), "fs", true); err != nil {
		t.Fatalf("SetEnabled failed: %v", err)
	}

	m := readConfigForTest(t, path)
	servers := m["mcpServers"].(map[string]interface{})
	entry := servers["fs"].(map[string]interface{})
	if entry["command"] != "fs-server" {
		t.Fatalf("expected command fs-server, got %v", entry)
	}
}

func TestSetEnabled_Global_Enable_SweepsAllProjects(t *testing.T) {
	cat := &catalog.Catalog{Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}}}
	a, path := newAdapter(t, cat)
	writeConfig(t, path, `{"projects":{
		"/a":{"disabledMcpServers":["fs"]},
		"/b":{"disabledMcpServers":["fs"]},
		"/c":{"disabledMcpServers":[]}
	}}`)

	swept, err := a.SetEnabled(globalScope(), "fs", true)
	if err != nil {
		t.Fatalf("SetEnabled failed: %v", err)
	}
	if swept != 2 {
		t.Fatalf("expected sweptProjects 2, got %d", swept)
	}

	m := readConfigForTest(t, path)
	projects := m["projects"].(map[string]interface{})
	for key, raw := range projects {
		proj := raw.(map[string]interface{})
		arr := proj["disabledMcpServers"].([]interface{})
		for _, v := range arr {
			if v == "fs" {
				t.Fatalf("expected fs swept from %s, got %v", key, arr)
			}
		}
	}
}

func TestSetEnabled_Global_Enable_UnknownServerErrors(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{})
	writeConfig(t, path, `{}`)

	before, _ := os.ReadFile(path)

	_, err := a.SetEnabled(globalScope(), "nope", true)
	if err == nil {
		t.Fatal("expected error for unknown server")
	}

	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatalf("expected no write performed, before=%s after=%s", before, after)
	}
}

func TestIsEnabled_Global_ReflectsEntryPresence(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{})
	writeConfig(t, path, `{"mcpServers":{"fs":{"command":"fs-server"}}}`)

	enabled, err := a.IsEnabled(globalScope(), "fs")
	if err != nil {
		t.Fatalf("IsEnabled failed: %v", err)
	}
	if !enabled {
		t.Fatal("expected fs to be enabled (entry present)")
	}

	enabled, err = a.IsEnabled(globalScope(), "other")
	if err != nil {
		t.Fatalf("IsEnabled failed: %v", err)
	}
	if enabled {
		t.Fatal("expected other to be disabled (entry absent)")
	}
}

func TestSetEnabled_PreservesUnrelatedConfig(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}}})
	writeConfig(t, path, `{"someOtherSetting":true,"mcpServers":{"other":{"command":"y"}},"projects":{"/other":{"disabledMcpServers":["keep"]}}}`)

	if _, err := a.SetEnabled(projectScope("/proj"), "fs", false); err != nil {
		t.Fatalf("SetEnabled failed: %v", err)
	}

	m := readConfigForTest(t, path)
	if m["someOtherSetting"] != true {
		t.Fatalf("expected someOtherSetting preserved, got %v", m["someOtherSetting"])
	}
	servers := m["mcpServers"].(map[string]interface{})
	if _, ok := servers["other"]; !ok {
		t.Fatalf("expected other server preserved, got %v", servers)
	}
	projects := m["projects"].(map[string]interface{})
	other := projects["/other"].(map[string]interface{})
	arr := other["disabledMcpServers"].([]interface{})
	if len(arr) != 1 || arr[0] != "keep" {
		t.Fatalf("expected /other preserved, got %v", arr)
	}
}

func TestSetEnabled_Idempotent(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{})
	writeConfig(t, path, `{"projects":{"/proj":{"disabledMcpServers":["fs"]}}}`)

	if _, err := a.SetEnabled(projectScope("/proj"), "fs", false); err != nil {
		t.Fatalf("SetEnabled failed: %v", err)
	}
	after1, _ := os.ReadFile(path)

	if _, err := a.SetEnabled(projectScope("/proj"), "fs", false); err != nil {
		t.Fatalf("SetEnabled failed: %v", err)
	}
	after2, _ := os.ReadFile(path)

	m1 := map[string]interface{}{}
	m2 := map[string]interface{}{}
	json.Unmarshal(after1, &m1)
	json.Unmarshal(after2, &m2)
	b1, _ := json.Marshal(m1)
	b2, _ := json.Marshal(m2)
	if string(b1) != string(b2) {
		t.Fatalf("expected idempotent result, got %s vs %s", b1, b2)
	}
}

func TestIsEnabled_MissingConfigFile(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{})

	enabled, err := a.IsEnabled(projectScope("/proj"), "fs")
	if err != nil {
		t.Fatalf("IsEnabled failed: %v", err)
	}
	if !enabled {
		t.Fatal("expected project scope to return true on missing file")
	}

	enabled, err = a.IsEnabled(globalScope(), "fs")
	if err != nil {
		t.Fatalf("IsEnabled failed: %v", err)
	}
	if enabled {
		t.Fatal("expected global scope to return false on missing file")
	}

	if _, err := os.Stat(path); err == nil {
		t.Fatal("expected no file to be created by a read")
	}
}

func TestSetEnabled_MissingConfigFile_CreatesFile(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{})

	if _, err := a.SetEnabled(projectScope("/proj"), "fs", false); err != nil {
		t.Fatalf("SetEnabled failed: %v", err)
	}

	m := readConfigForTest(t, path)
	projects := m["projects"].(map[string]interface{})
	proj := projects["/proj"].(map[string]interface{})
	arr := proj["disabledMcpServers"].([]interface{})
	if len(arr) != 1 || arr[0] != "fs" {
		t.Fatalf("expected disabledMcpServers to contain fs, got %v", arr)
	}
}

func findServer(servers []catalog.Server, name string) (catalog.Server, bool) {
	for _, s := range servers {
		if s.Name == name {
			return s, true
		}
	}
	return catalog.Server{}, false
}

func TestImportCandidates_Global_ReturnsAllTopLevelServers(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{})
	writeConfig(t, path, `{"mcpServers":{"fs":{"command":"fs-server"},"git":{"command":"git-server","args":["--root"]}}}`)

	candidates, conflicts, err := a.ImportCandidates(globalScope())
	if err != nil {
		t.Fatalf("ImportCandidates failed: %v", err)
	}
	if len(conflicts) != 0 {
		t.Fatalf("expected no conflicts, got %v", conflicts)
	}
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %v", candidates)
	}
	fs, ok := findServer(candidates, "fs")
	if !ok || fs.Command != "fs-server" {
		t.Fatalf("expected fs candidate with command fs-server, got %v", candidates)
	}
}

func TestImportCandidates_Project_MergesBothSources(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{})
	projectRoot := t.TempDir()
	mcpJSONPath := filepath.Join(projectRoot, ".mcp.json")
	writeConfig(t, mcpJSONPath, `{"mcpServers":{"a":{"command":"a-server"}}}`)
	writeConfig(t, path, `{"projects":{"`+projectRoot+`":{"mcpServers":{"b":{"command":"b-server"}}}}}`)

	candidates, conflicts, err := a.ImportCandidates(projectScope(projectRoot))
	if err != nil {
		t.Fatalf("ImportCandidates failed: %v", err)
	}
	if len(conflicts) != 0 {
		t.Fatalf("expected no conflicts, got %v", conflicts)
	}
	if _, ok := findServer(candidates, "a"); !ok {
		t.Fatalf("expected candidate 'a', got %v", candidates)
	}
	if _, ok := findServer(candidates, "b"); !ok {
		t.Fatalf("expected candidate 'b', got %v", candidates)
	}
}

func TestImportCandidates_Project_CollisionPrefersMcpJSON(t *testing.T) {
	a, path := newAdapter(t, &catalog.Catalog{})
	projectRoot := t.TempDir()
	mcpJSONPath := filepath.Join(projectRoot, ".mcp.json")
	writeConfig(t, mcpJSONPath, `{"mcpServers":{"fs":{"command":"from-mcp-json"}}}`)
	writeConfig(t, path, `{"projects":{"`+projectRoot+`":{"mcpServers":{"fs":{"command":"from-local"}}}}}`)

	candidates, conflicts, err := a.ImportCandidates(projectScope(projectRoot))
	if err != nil {
		t.Fatalf("ImportCandidates failed: %v", err)
	}
	fs, ok := findServer(candidates, "fs")
	if !ok || fs.Command != "from-mcp-json" {
		t.Fatalf("expected fs candidate from .mcp.json, got %v", candidates)
	}
	if len(conflicts) != 1 || conflicts[0].Name != "fs" {
		t.Fatalf("expected one conflict for fs, got %v", conflicts)
	}
}

func TestImportCandidates_MissingSourcesReturnEmpty(t *testing.T) {
	a, _ := newAdapter(t, &catalog.Catalog{})

	candidates, conflicts, err := a.ImportCandidates(globalScope())
	if err != nil || len(candidates) != 0 || len(conflicts) != 0 {
		t.Fatalf("expected empty results for missing global config, got %v, %v, %v", candidates, conflicts, err)
	}

	candidates, conflicts, err = a.ImportCandidates(projectScope(t.TempDir()))
	if err != nil || len(candidates) != 0 || len(conflicts) != 0 {
		t.Fatalf("expected empty results for missing project sources, got %v, %v, %v", candidates, conflicts, err)
	}
}
