package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmdExists(t *testing.T) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "mcputils") {
		t.Fatalf("expected output to contain %q, got %q", "mcputils", buf.String())
	}
}

func TestPersistentFlagsRegistered(t *testing.T) {
	if rootCmd.PersistentFlags().Lookup("global") == nil {
		t.Fatal("expected --global flag to be registered")
	}
	if rootCmd.PersistentFlags().Lookup("project") == nil {
		t.Fatal("expected --project flag to be registered")
	}
}
