package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeshanziya/mcputils/internal/history"
)

func TestCLI_Log_ShowsEntriesNewestFirst(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	histPath := filepath.Join(home, ".config", "mcputils", "history.jsonl")
	l := history.New(histPath)
	if err := l.Append("global", "server", "a", "enabled", 0); err != nil {
		t.Fatalf("setup append failed: %v", err)
	}
	if err := l.Append("global", "server", "b", "disabled", 0); err != nil {
		t.Fatalf("setup append failed: %v", err)
	}

	globalFlag = false
	projectFlag = ""
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"log"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out := buf.String()
	idxB := strings.Index(out, "b")
	idxA := strings.Index(out, "a")
	if idxB == -1 || idxA == -1 || idxB > idxA {
		t.Fatalf("expected 'b' (newest) to appear before 'a', got %q", out)
	}
}

func TestCLI_Log_EmptyWhenNoHistory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	globalFlag = false
	projectFlag = ""
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"log"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
