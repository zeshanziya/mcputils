package core

import (
	"testing"

	"github.com/zeshanziya/mcputils/internal/catalog"
	"github.com/zeshanziya/mcputils/internal/scope"
	"github.com/zeshanziya/mcputils/internal/state"
)

func TestEnableServer_UnknownNameErrors(t *testing.T) {
	cat := &catalog.Catalog{}
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	if err := EnableServer(cat, fa, hist, testScope(), "nope"); err == nil {
		t.Fatal("expected error for unknown server name")
	}
}

func TestDisableServer_CallsAdapter(t *testing.T) {
	cat := &catalog.Catalog{Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}}}
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	if err := DisableServer(cat, fa, hist, testScope(), "fs"); err != nil {
		t.Fatalf("DisableServer failed: %v", err)
	}
	if !fa.wasCalledWith("fs", false) {
		t.Fatal("expected SetEnabled(fs, false) called")
	}
}

func TestDisableServer_UnknownNameErrors(t *testing.T) {
	cat := &catalog.Catalog{}
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	if err := DisableServer(cat, fa, hist, testScope(), "nope"); err == nil {
		t.Fatal("expected error for unknown server name")
	}
}

func TestEnableServer_ProjectScope(t *testing.T) {
	cat := &catalog.Catalog{Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}}}
	fa := newFakeAdapter()
	hist := newHistLogger(t)
	sc := scope.Scope{IsGlobal: false, ProjectRoot: "/proj"}

	// A fresh project-scope server is already enabled by default (nothing in
	// disabledMcpServers), so disable it first to produce a genuine transition
	// before re-enabling it.
	if err := DisableServer(cat, fa, hist, sc, "fs"); err != nil {
		t.Fatalf("DisableServer failed: %v", err)
	}
	if err := EnableServer(cat, fa, hist, sc, "fs"); err != nil {
		t.Fatalf("EnableServer failed: %v", err)
	}

	entries, err := hist.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 2 || entries[0].Scope != "/proj" || entries[1].Scope != "/proj" {
		t.Fatalf("expected 2 history entries scoped to /proj, got %+v", entries)
	}
}

func TestEnableServer_NoOpWhenAlreadyEnabled(t *testing.T) {
	cat := &catalog.Catalog{Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}}}
	fa := newFakeAdapter()
	hist := newHistLogger(t)
	sc := scope.Scope{IsGlobal: true}

	if err := EnableServer(cat, fa, hist, sc, "fs"); err != nil {
		t.Fatalf("first EnableServer failed: %v", err)
	}
	if err := EnableServer(cat, fa, hist, sc, "fs"); err != nil {
		t.Fatalf("second EnableServer failed: %v", err)
	}

	if !fa.wasCalledWith("fs", true) {
		t.Fatal("expected SetEnabled(fs, true) to still be called on the repeat call")
	}
	entries, err := hist.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 history entry for one real transition, got %d: %+v", len(entries), entries)
	}
}

func TestDisableServer_NoOpWhenAlreadyDisabled(t *testing.T) {
	cat := &catalog.Catalog{Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}}}
	fa := newFakeAdapter()
	hist := newHistLogger(t)
	sc := scope.Scope{IsGlobal: true}

	if err := DisableServer(cat, fa, hist, sc, "fs"); err != nil {
		t.Fatalf("first DisableServer failed: %v", err)
	}
	if err := DisableServer(cat, fa, hist, sc, "fs"); err != nil {
		t.Fatalf("second DisableServer failed: %v", err)
	}

	entries, err := hist.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no history entries (already disabled by default), got %d: %+v", len(entries), entries)
	}
}

func TestProfileReEnable_OverridesIndividualDisable(t *testing.T) {
	cat := &catalog.Catalog{
		Servers:  []catalog.Server{{Name: "fs", Command: "fs-server"}},
		Profiles: []catalog.Profile{{Name: "work", Servers: []string{"fs"}}},
	}
	st := &state.State{}
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	if err := DisableServer(cat, fa, hist, testScope(), "fs"); err != nil {
		t.Fatalf("DisableServer failed: %v", err)
	}
	if err := EnableProfile(cat, st, fa, hist, testScope(), "work"); err != nil {
		t.Fatalf("EnableProfile failed: %v", err)
	}

	if !fa.wasCalledWith("fs", true) {
		t.Fatal("expected SetEnabled(fs, true) called by EnableProfile")
	}
	enabled, _ := fa.IsEnabled(testScope(), "fs")
	if !enabled {
		t.Fatal("expected final state enabled")
	}
}

func TestEnableServer_CallsAdapter(t *testing.T) {
	cat := &catalog.Catalog{Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}}}
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	if err := EnableServer(cat, fa, hist, testScope(), "fs"); err != nil {
		t.Fatalf("EnableServer failed: %v", err)
	}
	if !fa.wasCalledWith("fs", true) {
		t.Fatal("expected SetEnabled(fs, true) called")
	}

	entries, err := hist.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(entries))
	}
	if entries[0].Trigger != "server" || entries[0].State != "enabled" {
		t.Fatalf("unexpected history entry: %+v", entries[0])
	}
}
