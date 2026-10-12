package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/zeshanziya/mcputils/internal/catalog"
)

// runCLI executes rootCmd with args against an isolated $HOME (for global
// scope) and returns captured stdout/stderr combined, and the exit error.
func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	globalFlag = false
	projectFlag = ""

	err := rootCmd.Execute()
	return buf.String(), err
}

func writeCatalogFixture(t *testing.T, path string, cat *catalog.Catalog) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("setup mkdir failed: %v", err)
	}
	if err := cat.Save(path); err != nil {
		t.Fatalf("setup save catalog failed: %v", err)
	}
}
