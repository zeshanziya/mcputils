package jsonutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readJSON(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal %s: %v", path, err)
	}
	return m
}

func TestWriteAtomic_TempRename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	if err := WriteAtomic(path, []byte(`{"a":1}`)); err != nil {
		t.Fatalf("WriteAtomic failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read target file: %v", err)
	}
	if string(data) != `{"a":1}` {
		t.Fatalf("expected content %q, got %q", `{"a":1}`, string(data))
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("expected no .tmp file left behind, found %s", e.Name())
		}
	}
}

func TestPatchStringArray_AddsName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	err := PatchStringArray(path, "key", func(current []string) []string {
		return append(current, "name")
	})
	if err != nil {
		t.Fatalf("PatchStringArray failed: %v", err)
	}

	m := readJSON(t, path)
	arr, ok := m["key"].([]interface{})
	if !ok || len(arr) != 1 || arr[0] != "name" {
		t.Fatalf(`expected key: ["name"], got %v`, m["key"])
	}
}

func TestPatchStringArray_RemovesName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	if err := WriteAtomic(path, []byte(`{"key":["a","name","b"]}`)); err != nil {
		t.Fatalf("setup WriteAtomic failed: %v", err)
	}

	err := PatchStringArray(path, "key", func(current []string) []string {
		out := []string{}
		for _, v := range current {
			if v != "name" {
				out = append(out, v)
			}
		}
		return out
	})
	if err != nil {
		t.Fatalf("PatchStringArray failed: %v", err)
	}

	m := readJSON(t, path)
	arr := m["key"].([]interface{})
	for _, v := range arr {
		if v == "name" {
			t.Fatalf("expected name to be removed, got %v", arr)
		}
	}
}

func TestPatchStringArray_PreservesUnknownFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	if err := WriteAtomic(path, []byte(`{"key":[],"other":"untouched","nested":{"x":1}}`)); err != nil {
		t.Fatalf("setup WriteAtomic failed: %v", err)
	}

	err := PatchStringArray(path, "key", func(current []string) []string {
		return append(current, "name")
	})
	if err != nil {
		t.Fatalf("PatchStringArray failed: %v", err)
	}

	m := readJSON(t, path)
	if m["other"] != "untouched" {
		t.Fatalf("expected 'other' to be preserved, got %v", m["other"])
	}
	nested, ok := m["nested"].(map[string]interface{})
	if !ok || nested["x"] != float64(1) {
		t.Fatalf("expected 'nested' to be preserved, got %v", m["nested"])
	}
}

func TestPatchStringArray_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.json")

	err := PatchStringArray(path, "key", func(current []string) []string {
		return append(current, "name")
	})
	if err != nil {
		t.Fatalf("PatchStringArray on missing file failed: %v", err)
	}

	m := readJSON(t, path)
	arr, ok := m["key"].([]interface{})
	if !ok || len(arr) != 1 || arr[0] != "name" {
		t.Fatalf(`expected key: ["name"], got %v`, m["key"])
	}
}

func TestPatchNestedStringArray_CreatesIntermediateObject(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	err := PatchNestedStringArray(path, "projects", "/x", "disabledMcpServers", func(current []string) []string {
		return append(current, "fs")
	})
	if err != nil {
		t.Fatalf("PatchNestedStringArray failed: %v", err)
	}

	m := readJSON(t, path)
	projects := m["projects"].(map[string]interface{})
	proj := projects["/x"].(map[string]interface{})
	arr := proj["disabledMcpServers"].([]interface{})
	if len(arr) != 1 || arr[0] != "fs" {
		t.Fatalf("expected disabledMcpServers to contain fs, got %v", arr)
	}
}

func TestPatchNestedStringArray_PreservesOtherOuterEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	if err := WriteAtomic(path, []byte(`{"projects":{"/x":{"disabledMcpServers":[]},"/y":{"disabledMcpServers":["keep"]}}}`)); err != nil {
		t.Fatalf("setup WriteAtomic failed: %v", err)
	}

	err := PatchNestedStringArray(path, "projects", "/x", "disabledMcpServers", func(current []string) []string {
		return append(current, "fs")
	})
	if err != nil {
		t.Fatalf("PatchNestedStringArray failed: %v", err)
	}

	m := readJSON(t, path)
	projects := m["projects"].(map[string]interface{})
	y := projects["/y"].(map[string]interface{})
	arr := y["disabledMcpServers"].([]interface{})
	if len(arr) != 1 || arr[0] != "keep" {
		t.Fatalf("expected /y entry untouched, got %v", arr)
	}
}

func TestPatchMapEntry_SetsEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	value := map[string]interface{}{"command": "fs-server", "args": []string{"--x"}}
	if err := PatchMapEntry(path, "mcpServers", "fs", value); err != nil {
		t.Fatalf("PatchMapEntry failed: %v", err)
	}

	m := readJSON(t, path)
	servers := m["mcpServers"].(map[string]interface{})
	entry := servers["fs"].(map[string]interface{})
	if entry["command"] != "fs-server" {
		t.Fatalf("expected command fs-server, got %v", entry["command"])
	}
}

func TestPatchMapEntry_DeletesEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	if err := WriteAtomic(path, []byte(`{"mcpServers":{"fs":{"command":"fs-server"},"other":{"command":"x"}}}`)); err != nil {
		t.Fatalf("setup WriteAtomic failed: %v", err)
	}

	if err := PatchMapEntry(path, "mcpServers", "fs", nil); err != nil {
		t.Fatalf("PatchMapEntry failed: %v", err)
	}

	m := readJSON(t, path)
	servers := m["mcpServers"].(map[string]interface{})
	if _, ok := servers["fs"]; ok {
		t.Fatalf("expected fs entry to be deleted, got %v", servers)
	}
	if _, ok := servers["other"]; !ok {
		t.Fatalf("expected other entry to be preserved, got %v", servers)
	}
}

func TestPatchMapEntry_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.json")

	value := map[string]interface{}{"command": "fs-server"}
	if err := PatchMapEntry(path, "mcpServers", "fs", value); err != nil {
		t.Fatalf("PatchMapEntry on missing file failed: %v", err)
	}

	m := readJSON(t, path)
	servers := m["mcpServers"].(map[string]interface{})
	entry := servers["fs"].(map[string]interface{})
	if entry["command"] != "fs-server" {
		t.Fatalf("expected command fs-server, got %v", entry["command"])
	}
}

func TestSweepNestedStringArrays_RemovesFromAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	content := `{"projects":{
		"/a":{"disabledMcpServers":["fs","other"]},
		"/b":{"disabledMcpServers":["fs"]},
		"/c":{"disabledMcpServers":[]},
		"/d":{"disabledMcpServers":["fs","keep"]},
		"/e":{"disabledMcpServers":["unrelated"]}
	}}`
	if err := WriteAtomic(path, []byte(content)); err != nil {
		t.Fatalf("setup WriteAtomic failed: %v", err)
	}

	count, err := SweepNestedStringArrays(path, "projects", "disabledMcpServers", "fs")
	if err != nil {
		t.Fatalf("SweepNestedStringArrays failed: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected sweptCount 3, got %d", count)
	}

	m := readJSON(t, path)
	projects := m["projects"].(map[string]interface{})
	for key, raw := range projects {
		proj := raw.(map[string]interface{})
		arr := proj["disabledMcpServers"].([]interface{})
		for _, v := range arr {
			if v == "fs" {
				t.Fatalf("expected fs removed from %s, got %v", key, arr)
			}
		}
	}
	aArr := projects["/a"].(map[string]interface{})["disabledMcpServers"].([]interface{})
	if len(aArr) != 1 || aArr[0] != "other" {
		t.Fatalf("expected /a to retain 'other', got %v", aArr)
	}
	dArr := projects["/d"].(map[string]interface{})["disabledMcpServers"].([]interface{})
	if len(dArr) != 1 || dArr[0] != "keep" {
		t.Fatalf("expected /d to retain 'keep', got %v", dArr)
	}
	eArr := projects["/e"].(map[string]interface{})["disabledMcpServers"].([]interface{})
	if len(eArr) != 1 || eArr[0] != "unrelated" {
		t.Fatalf("expected /e untouched, got %v", eArr)
	}
}

func TestSweepNestedStringArrays_NoOpWhenAbsentEverywhere(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	content := `{"projects":{"/a":{"disabledMcpServers":["other"]}}}`
	if err := WriteAtomic(path, []byte(content)); err != nil {
		t.Fatalf("setup WriteAtomic failed: %v", err)
	}

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read before: %v", err)
	}

	count, err := SweepNestedStringArrays(path, "projects", "disabledMcpServers", "fs")
	if err != nil {
		t.Fatalf("SweepNestedStringArrays failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected sweptCount 0, got %d", count)
	}

	afterMap := readJSON(t, path)
	beforeMap := map[string]interface{}{}
	if err := json.Unmarshal(before, &beforeMap); err != nil {
		t.Fatalf("failed to unmarshal before snapshot: %v", err)
	}
	beforeJSON, _ := json.Marshal(beforeMap)
	afterJSON, _ := json.Marshal(afterMap)
	if string(beforeJSON) != string(afterJSON) {
		t.Fatalf("expected content unchanged, before=%s after=%s", beforeJSON, afterJSON)
	}
}

func TestSweepNestedStringArrays_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.json")

	count, err := SweepNestedStringArrays(path, "projects", "disabledMcpServers", "fs")
	if err != nil {
		t.Fatalf("SweepNestedStringArrays on missing file failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected sweptCount 0, got %d", count)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected no file to be created")
	}
}

func TestReadMapEntries_ReturnsEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"mcpServers":{"fs":{"command":"fs-server"}}}`), 0o600); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	entries, err := ReadMapEntries(path, "mcpServers")
	if err != nil {
		t.Fatalf("ReadMapEntries failed: %v", err)
	}
	if _, ok := entries["fs"]; !ok {
		t.Fatalf("expected entries to contain 'fs', got %v", entries)
	}
}

func TestReadMapEntries_MissingFileOrKeyReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	missingPath := filepath.Join(dir, "missing.json")
	entries, err := ReadMapEntries(missingPath, "mcpServers")
	if err != nil {
		t.Fatalf("ReadMapEntries on missing file failed: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty map for missing file, got %v", entries)
	}

	noKeyPath := filepath.Join(dir, "nokey.json")
	if err := os.WriteFile(noKeyPath, []byte(`{"other":1}`), 0o600); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}
	entries, err = ReadMapEntries(noKeyPath, "mcpServers")
	if err != nil {
		t.Fatalf("ReadMapEntries on missing key failed: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty map for missing key, got %v", entries)
	}
}

func TestReadNestedMapEntries_ReturnsEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{"projects":{"/x":{"mcpServers":{"a":{"command":"a-server"}}}}}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	entries, err := ReadNestedMapEntries(path, "projects", "/x", "mcpServers")
	if err != nil {
		t.Fatalf("ReadNestedMapEntries failed: %v", err)
	}
	if _, ok := entries["a"]; !ok {
		t.Fatalf("expected entries to contain 'a', got %v", entries)
	}
}

func TestReadNestedMapEntries_MissingAnyLevelReturnsEmpty(t *testing.T) {
	dir := t.TempDir()

	missingFile := filepath.Join(dir, "missing.json")
	entries, err := ReadNestedMapEntries(missingFile, "projects", "/x", "mcpServers")
	if err != nil || len(entries) != 0 {
		t.Fatalf("expected empty map, no error for missing file; got %v, %v", entries, err)
	}

	noOuterKey := filepath.Join(dir, "no_outer.json")
	os.WriteFile(noOuterKey, []byte(`{}`), 0o600)
	entries, err = ReadNestedMapEntries(noOuterKey, "projects", "/x", "mcpServers")
	if err != nil || len(entries) != 0 {
		t.Fatalf("expected empty map, no error for missing outer key; got %v, %v", entries, err)
	}

	noSubKey := filepath.Join(dir, "no_sub.json")
	os.WriteFile(noSubKey, []byte(`{"projects":{}}`), 0o600)
	entries, err = ReadNestedMapEntries(noSubKey, "projects", "/x", "mcpServers")
	if err != nil || len(entries) != 0 {
		t.Fatalf("expected empty map, no error for missing outerSubKey; got %v, %v", entries, err)
	}

	noInnerKey := filepath.Join(dir, "no_inner.json")
	os.WriteFile(noInnerKey, []byte(`{"projects":{"/x":{}}}`), 0o600)
	entries, err = ReadNestedMapEntries(noInnerKey, "projects", "/x", "mcpServers")
	if err != nil || len(entries) != 0 {
		t.Fatalf("expected empty map, no error for missing innerKey; got %v, %v", entries, err)
	}
}

func TestPatchNestedMapEntry_CreatesIntermediateObjects(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	err := PatchNestedMapEntry(path, "projects", "/x", "mcpServers", "fs", map[string]interface{}{"command": "fs-server"})
	if err != nil {
		t.Fatalf("PatchNestedMapEntry failed: %v", err)
	}

	m := readJSON(t, path)
	projects := m["projects"].(map[string]interface{})
	proj := projects["/x"].(map[string]interface{})
	servers := proj["mcpServers"].(map[string]interface{})
	entry := servers["fs"].(map[string]interface{})
	if entry["command"] != "fs-server" {
		t.Fatalf("expected command fs-server, got %v", entry)
	}
}

func TestPatchNestedMapEntry_PreservesOtherEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"projects":{"/x":{"mcpServers":{"y":{"command":"y-server"}}}}}`), 0o600); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	err := PatchNestedMapEntry(path, "projects", "/x", "mcpServers", "z", map[string]interface{}{"command": "z-server"})
	if err != nil {
		t.Fatalf("PatchNestedMapEntry failed: %v", err)
	}

	m := readJSON(t, path)
	servers := m["projects"].(map[string]interface{})["/x"].(map[string]interface{})["mcpServers"].(map[string]interface{})
	y := servers["y"].(map[string]interface{})
	if y["command"] != "y-server" {
		t.Fatalf("expected y entry preserved, got %v", y)
	}
	z := servers["z"].(map[string]interface{})
	if z["command"] != "z-server" {
		t.Fatalf("expected z entry added, got %v", z)
	}
}

func TestPatchNestedMapEntry_DeletesEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"projects":{"/x":{"mcpServers":{"y":{"command":"y-server"},"z":{"command":"z-server"}}}}}`), 0o600); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	err := PatchNestedMapEntry(path, "projects", "/x", "mcpServers", "z", nil)
	if err != nil {
		t.Fatalf("PatchNestedMapEntry failed: %v", err)
	}

	m := readJSON(t, path)
	servers := m["projects"].(map[string]interface{})["/x"].(map[string]interface{})["mcpServers"].(map[string]interface{})
	if _, ok := servers["z"]; ok {
		t.Fatalf("expected z deleted, got %v", servers)
	}
	if _, ok := servers["y"]; !ok {
		t.Fatalf("expected y preserved, got %v", servers)
	}
}

func TestPatchNestedMapEntry_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.json")

	err := PatchNestedMapEntry(path, "projects", "/x", "mcpServers", "fs", map[string]interface{}{"command": "fs-server"})
	if err != nil {
		t.Fatalf("PatchNestedMapEntry on missing file failed: %v", err)
	}

	m := readJSON(t, path)
	servers := m["projects"].(map[string]interface{})["/x"].(map[string]interface{})["mcpServers"].(map[string]interface{})
	entry := servers["fs"].(map[string]interface{})
	if entry["command"] != "fs-server" {
		t.Fatalf("expected command fs-server, got %v", entry)
	}
}
