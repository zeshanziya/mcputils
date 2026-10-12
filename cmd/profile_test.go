package cmd

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/zeshanziya/mcputils/internal/catalog"
)

func TestCLI_ProfileCreate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	globalFlag = false
	projectFlag = ""
	rootCmd.SetArgs([]string{"--global", "profile", "create", "work"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	cat, err := catalog.Load(catPath)
	if err != nil {
		t.Fatalf("failed to load catalog: %v", err)
	}
	if _, ok := cat.FindProfile("work"); !ok {
		t.Fatalf("expected profile 'work' to exist, got %+v", cat.Profiles)
	}
}

func TestCLI_ProfileAddServer(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Servers:  []catalog.Server{{Name: "fs", Command: "fs-server"}},
		Profiles: []catalog.Profile{{Name: "work"}},
	})

	globalFlag = false
	projectFlag = ""
	rootCmd.SetArgs([]string{"--global", "profile", "add-server", "work", "fs"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cat, err := catalog.Load(catPath)
	if err != nil {
		t.Fatalf("failed to load catalog: %v", err)
	}
	p, ok := cat.FindProfile("work")
	if !ok || len(p.Servers) != 1 || p.Servers[0] != "fs" {
		t.Fatalf("expected profile 'work' to list 'fs', got %+v", p)
	}
}

func TestCLI_ProfileCreate_MissingArgErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	globalFlag = false
	projectFlag = ""
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--global", "profile", "create"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing name argument")
	}
}

func TestCLI_ProfileRemoveServer(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	catPath := filepath.Join(home, ".config", "mcputils", "servers.yaml")
	writeCatalogFixture(t, catPath, &catalog.Catalog{
		Servers:  []catalog.Server{{Name: "fs", Command: "fs-server"}},
		Profiles: []catalog.Profile{{Name: "work", Servers: []string{"fs"}}},
	})

	globalFlag = false
	projectFlag = ""
	rootCmd.SetArgs([]string{"--global", "profile", "remove-server", "work", "fs"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cat, err := catalog.Load(catPath)
	if err != nil {
		t.Fatalf("failed to load catalog: %v", err)
	}
	p, ok := cat.FindProfile("work")
	if !ok || len(p.Servers) != 0 {
		t.Fatalf("expected profile 'work' to no longer list 'fs', got %+v", p)
	}
}
