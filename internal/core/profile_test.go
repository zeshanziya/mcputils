package core

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeshanziya/mcputils/internal/catalog"
	"github.com/zeshanziya/mcputils/internal/history"
	"github.com/zeshanziya/mcputils/internal/scope"
	"github.com/zeshanziya/mcputils/internal/state"
)

func newHistLogger(t *testing.T) *history.Logger {
	t.Helper()
	dir := t.TempDir()
	return history.New(filepath.Join(dir, "history.jsonl"))
}

func testScope() scope.Scope {
	return scope.Scope{IsGlobal: true}
}

func TestCreateProfile_DuplicateErrors(t *testing.T) {
	cat := &catalog.Catalog{}
	if err := CreateProfile(cat, "work"); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	if err := CreateProfile(cat, "work"); err == nil {
		t.Fatal("expected error on duplicate profile create")
	}
}

func TestAddServerToProfile_UnknownProfileErrors(t *testing.T) {
	cat := &catalog.Catalog{Servers: []catalog.Server{{Name: "fs", Command: "fs-server"}}}
	err := AddServerToProfile(cat, "nope", "fs")
	if err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestRemoveServerFromProfile_Success(t *testing.T) {
	cat := &catalog.Catalog{
		Servers:  []catalog.Server{{Name: "fs", Command: "fs-server"}},
		Profiles: []catalog.Profile{{Name: "work", Servers: []string{"fs"}}},
	}
	if err := RemoveServerFromProfile(cat, "work", "fs"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	p, _ := cat.FindProfile("work")
	if len(p.Servers) != 0 {
		t.Fatalf("expected fs removed, got %v", p.Servers)
	}
}

func TestRemoveServerFromProfile_UnknownProfileErrors(t *testing.T) {
	cat := &catalog.Catalog{}
	if err := RemoveServerFromProfile(cat, "nope", "fs"); err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestAddServerToProfile_UnknownServerErrors(t *testing.T) {
	cat := &catalog.Catalog{Profiles: []catalog.Profile{{Name: "work"}}}
	err := AddServerToProfile(cat, "work", "nope")
	if err == nil {
		t.Fatal("expected error for unknown server")
	}
	p, _ := cat.FindProfile("work")
	if len(p.Servers) != 0 {
		t.Fatalf("expected profile unchanged, got %v", p.Servers)
	}
}

func TestAddServerToProfile_AlreadyPresentIsNoOp(t *testing.T) {
	cat := &catalog.Catalog{
		Servers:  []catalog.Server{{Name: "fs", Command: "fs-server"}},
		Profiles: []catalog.Profile{{Name: "work", Servers: []string{"fs"}}},
	}
	if err := AddServerToProfile(cat, "work", "fs"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	p, _ := cat.FindProfile("work")
	count := 0
	for _, s := range p.Servers {
		if s == "fs" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected fs exactly once, got %v", p.Servers)
	}
}

func buildProfileFixture() (*catalog.Catalog, *state.State) {
	cat := &catalog.Catalog{
		Servers: []catalog.Server{
			{Name: "a", Command: "a-server"},
			{Name: "b", Command: "b-server"},
			{Name: "c", Command: "c-server"},
		},
		Profiles: []catalog.Profile{
			{Name: "work", Servers: []string{"a", "b", "c"}},
		},
	}
	st := &state.State{}
	return cat, st
}

func TestEnableProfile_EnablesAllMembers(t *testing.T) {
	cat, st := buildProfileFixture()
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	if err := EnableProfile(cat, st, fa, hist, testScope(), "work"); err != nil {
		t.Fatalf("EnableProfile failed: %v", err)
	}

	for _, name := range []string{"a", "b", "c"} {
		if !fa.wasCalledWith(name, true) {
			t.Fatalf("expected SetEnabled(%s, true) to be called", name)
		}
	}
	if !st.IsActive("work") {
		t.Fatal("expected profile marked active")
	}
}

func TestEnableProfile_DoesNotTouchOthers(t *testing.T) {
	cat, st := buildProfileFixture()
	cat.Servers = append(cat.Servers, catalog.Server{Name: "other", Command: "other-server"})
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	if err := EnableProfile(cat, st, fa, hist, testScope(), "work"); err != nil {
		t.Fatalf("EnableProfile failed: %v", err)
	}

	if fa.wasCalledFor("other") {
		t.Fatal("expected adapter never called for 'other'")
	}
}

func TestDisableProfile_SkipsSharedServer(t *testing.T) {
	cat := &catalog.Catalog{
		Servers: []catalog.Server{
			{Name: "x", Command: "x-server"},
			{Name: "y", Command: "y-server"},
		},
		Profiles: []catalog.Profile{
			{Name: "a", Servers: []string{"x", "y"}},
			{Name: "b", Servers: []string{"x"}},
		},
	}
	st := &state.State{}
	st.SetActive("a", true)
	st.SetActive("b", true)
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	if err := DisableProfile(cat, st, fa, hist, testScope(), "a"); err != nil {
		t.Fatalf("DisableProfile failed: %v", err)
	}

	if fa.wasCalledFor("x") {
		t.Fatal("expected adapter NOT called for shared server x")
	}
	if !fa.wasCalledWith("y", false) {
		t.Fatal("expected adapter called false for y")
	}
}

func TestDisableProfile_DisablesUniqueServer(t *testing.T) {
	cat := &catalog.Catalog{
		Servers:  []catalog.Server{{Name: "y", Command: "y-server"}},
		Profiles: []catalog.Profile{{Name: "a", Servers: []string{"y"}}},
	}
	st := &state.State{}
	st.SetActive("a", true)
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	if err := DisableProfile(cat, st, fa, hist, testScope(), "a"); err != nil {
		t.Fatalf("DisableProfile failed: %v", err)
	}

	if !fa.wasCalledWith("y", false) {
		t.Fatal("expected SetEnabled(y, false) called")
	}
}

func TestEnableProfile_MissingCatalogServerErrors(t *testing.T) {
	cat := &catalog.Catalog{
		Servers:  []catalog.Server{{Name: "a", Command: "a-server"}},
		Profiles: []catalog.Profile{{Name: "work", Servers: []string{"a", "ghost"}}},
	}
	st := &state.State{}
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	err := EnableProfile(cat, st, fa, hist, testScope(), "work")
	if err == nil {
		t.Fatal("expected error for missing catalog server")
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("expected error to name 'ghost', got %v", err)
	}
	if len(fa.calls) != 0 {
		t.Fatalf("expected adapter never called, got %v", fa.calls)
	}
}

func TestEnableProfile_UnknownProfileErrors(t *testing.T) {
	cat := &catalog.Catalog{}
	st := &state.State{}
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	err := EnableProfile(cat, st, fa, hist, testScope(), "nope")
	if err == nil {
		t.Fatal("expected error for unknown profile")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Fatalf("expected error to name 'nope', got %v", err)
	}
	if len(fa.calls) != 0 {
		t.Fatalf("expected adapter never called, got %v", fa.calls)
	}
}

func TestDisableProfile_UnknownProfileErrors(t *testing.T) {
	cat := &catalog.Catalog{}
	st := &state.State{}
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	err := DisableProfile(cat, st, fa, hist, testScope(), "nope")
	if err == nil {
		t.Fatal("expected error for unknown profile")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Fatalf("expected error to name 'nope', got %v", err)
	}
	if len(fa.calls) != 0 {
		t.Fatalf("expected adapter never called, got %v", fa.calls)
	}
}

func TestDisableProfile_InactiveProfileNoOp(t *testing.T) {
	cat := &catalog.Catalog{
		Servers:  []catalog.Server{{Name: "y", Command: "y-server"}},
		Profiles: []catalog.Profile{{Name: "work", Servers: []string{"y"}}},
	}
	st := &state.State{}
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	if err := DisableProfile(cat, st, fa, hist, testScope(), "work"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !fa.wasCalledWith("y", false) {
		t.Fatal("expected SetEnabled(y, false) still called")
	}
	if st.IsActive("work") {
		t.Fatal("expected profile absent from ActiveProfiles")
	}
}

func TestDisableProfile_IgnoresPriorIndividualEnable(t *testing.T) {
	cat := &catalog.Catalog{
		Servers:  []catalog.Server{{Name: "x", Command: "x-server"}},
		Profiles: []catalog.Profile{{Name: "p", Servers: []string{"x"}}},
	}
	st := &state.State{}
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	if err := EnableServer(cat, fa, hist, testScope(), "x"); err != nil {
		t.Fatalf("EnableServer failed: %v", err)
	}
	if err := EnableProfile(cat, st, fa, hist, testScope(), "p"); err != nil {
		t.Fatalf("EnableProfile failed: %v", err)
	}
	if err := DisableProfile(cat, st, fa, hist, testScope(), "p"); err != nil {
		t.Fatalf("DisableProfile failed: %v", err)
	}

	if !fa.wasCalledWith("x", false) {
		t.Fatal("expected SetEnabled(x, false) to be called despite prior individual enable")
	}
	enabled, _ := fa.IsEnabled(testScope(), "x")
	if enabled {
		t.Fatal("expected final IsEnabled(x) to be false")
	}
}

func TestEnableProfile_NoOpWhenAlreadyEnabledDoesNotReLog(t *testing.T) {
	cat, st := buildProfileFixture()
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	if err := EnableProfile(cat, st, fa, hist, testScope(), "work"); err != nil {
		t.Fatalf("first EnableProfile failed: %v", err)
	}
	if err := EnableProfile(cat, st, fa, hist, testScope(), "work"); err != nil {
		t.Fatalf("second EnableProfile failed: %v", err)
	}

	for _, name := range []string{"a", "b", "c"} {
		if !fa.wasCalledWith(name, true) {
			t.Fatalf("expected SetEnabled(%s, true) to still be called on the repeat call", name)
		}
	}
	entries, err := hist.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected exactly 3 history entries (one per server's real transition), got %d: %+v", len(entries), entries)
	}
}

func TestDisableProfile_NoOpWhenAlreadyDisabledDoesNotReLog(t *testing.T) {
	cat := &catalog.Catalog{
		Servers:  []catalog.Server{{Name: "y", Command: "y-server"}},
		Profiles: []catalog.Profile{{Name: "a", Servers: []string{"y"}}},
	}
	st := &state.State{}
	st.SetActive("a", true)
	fa := newFakeAdapter()
	hist := newHistLogger(t)

	if err := DisableProfile(cat, st, fa, hist, testScope(), "a"); err != nil {
		t.Fatalf("first DisableProfile failed: %v", err)
	}
	st.SetActive("a", true)
	if err := DisableProfile(cat, st, fa, hist, testScope(), "a"); err != nil {
		t.Fatalf("second DisableProfile failed: %v", err)
	}

	entries, err := hist.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no history entries (already disabled by default), got %d: %+v", len(entries), entries)
	}
}
