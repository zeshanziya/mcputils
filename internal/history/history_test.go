package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppend_CreatesFileAndParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "history.jsonl")
	l := New(path)

	if err := l.Append("global", "server", "fs", "enabled", 0); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	var e Entry
	if err := json.Unmarshal([]byte(lines[0]), &e); err != nil {
		t.Fatalf("failed to unmarshal entry: %v", err)
	}
	if e.Scope != "global" || e.Trigger != "server" || e.Server != "fs" || e.State != "enabled" {
		t.Fatalf("unexpected entry: %+v", e)
	}
}

func TestAppend_AppendsWithoutTruncating(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	l := New(path)

	if err := l.Append("global", "server", "a", "enabled", 0); err != nil {
		t.Fatalf("Append failed: %v", err)
	}
	if err := l.Append("global", "server", "b", "enabled", 0); err != nil {
		t.Fatalf("Append failed: %v", err)
	}
	before, _ := os.ReadFile(path)
	beforeLines := strings.Split(strings.TrimSpace(string(before)), "\n")

	if err := l.Append("global", "server", "c", "enabled", 0); err != nil {
		t.Fatalf("Append failed: %v", err)
	}
	after, _ := os.ReadFile(path)
	afterLines := strings.Split(strings.TrimSpace(string(after)), "\n")

	if len(afterLines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(afterLines))
	}
	for i := range beforeLines {
		if beforeLines[i] != afterLines[i] {
			t.Fatalf("expected line %d unchanged, before=%q after=%q", i, beforeLines[i], afterLines[i])
		}
	}
}

func TestAppend_OmitsSweptProjectsWhenZero(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	l := New(path)

	if err := l.Append("global", "server", "fs", "enabled", 0); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "swept_projects") {
		t.Fatalf("expected no swept_projects key, got %s", data)
	}
}

func TestAppend_IncludesSweptProjectsWhenNonZero(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	l := New(path)

	if err := l.Append("global", "server", "fs", "enabled", 2); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `"swept_projects":2`) {
		t.Fatalf(`expected "swept_projects":2 in output, got %s`, data)
	}
}

func TestList_ReturnsNewestFirst(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	l := New(path)

	l.Append("global", "server", "a", "enabled", 0)
	l.Append("global", "server", "b", "enabled", 0)
	l.Append("global", "server", "c", "enabled", 0)

	entries, err := l.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Server != "c" || entries[1].Server != "b" || entries[2].Server != "a" {
		t.Fatalf("expected [c,b,a], got %v", []string{entries[0].Server, entries[1].Server, entries[2].Server})
	}
}

func TestList_EmptyFileOrMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.jsonl")
	l := New(path)

	entries, err := l.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty slice, got %v", entries)
	}
}

func TestList_SkipsMalformedTrailingLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	l := New(path)

	l.Append("global", "server", "a", "enabled", 0)
	l.Append("global", "server", "b", "enabled", 0)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("failed to open log for corrupt append: %v", err)
	}
	if _, err := f.WriteString(`{"timestamp":"2026-01-01T00:00:00Z","scope":"glob`); err != nil {
		t.Fatalf("failed to write corrupt trailing line: %v", err)
	}
	f.Close()

	entries, err := l.List()
	if err != nil {
		t.Fatalf("expected malformed trailing line to be tolerated, got error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 valid entries, got %d", len(entries))
	}
	if entries[0].Server != "b" || entries[1].Server != "a" {
		t.Fatalf("expected [b,a], got %v", []string{entries[0].Server, entries[1].Server})
	}

	// A malformed line NOT at the end is a real error.
	path2 := filepath.Join(dir, "history2.jsonl")
	content := `{"timestamp":"2026-01-01T00:00:00Z","scope":"global","trigger":"server","server":"a","state":"enabled"}
not valid json at all
{"timestamp":"2026-01-01T00:00:01Z","scope":"global","trigger":"server","server":"b","state":"enabled"}
`
	if err := os.WriteFile(path2, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	l2 := New(path2)
	if _, err := l2.List(); err == nil {
		t.Fatal("expected error for malformed line not at the end")
	}
}
