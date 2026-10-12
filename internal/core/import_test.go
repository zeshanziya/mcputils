package core

import (
	"errors"
	"testing"

	"github.com/zeshanziya/mcputils/internal/adapter"
	"github.com/zeshanziya/mcputils/internal/catalog"
	"github.com/zeshanziya/mcputils/internal/scope"
)

type fakeImportAdapter struct {
	*fakeAdapter
	candidates []catalog.Server
	conflicts  []adapter.ImportConflict
	err        error
}

func (f *fakeImportAdapter) ImportCandidates(sc scope.Scope) ([]catalog.Server, []adapter.ImportConflict, error) {
	return f.candidates, f.conflicts, f.err
}

func TestImportServers_AddsNewCandidates(t *testing.T) {
	cat := &catalog.Catalog{}
	fa := &fakeImportAdapter{
		fakeAdapter: newFakeAdapter(),
		candidates: []catalog.Server{
			{Name: "a", Command: "a-server"},
			{Name: "b", Command: "b-server"},
		},
	}

	result, err := ImportServers(cat, fa, testScope())
	if err != nil {
		t.Fatalf("ImportServers failed: %v", err)
	}
	if len(result.Imported) != 2 {
		t.Fatalf("expected 2 imported, got %v", result.Imported)
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("expected no skipped, got %v", result.Skipped)
	}
	if _, ok := cat.FindServer("a"); !ok {
		t.Fatal("expected 'a' added to catalog")
	}
	if _, ok := cat.FindServer("b"); !ok {
		t.Fatal("expected 'b' added to catalog")
	}
}

func TestImportServers_SkipsAlreadyInCatalog(t *testing.T) {
	cat := &catalog.Catalog{Servers: []catalog.Server{{Name: "fs", Command: "existing"}}}
	fa := &fakeImportAdapter{
		fakeAdapter: newFakeAdapter(),
		candidates:  []catalog.Server{{Name: "fs", Command: "from-source"}},
	}

	result, err := ImportServers(cat, fa, testScope())
	if err != nil {
		t.Fatalf("ImportServers failed: %v", err)
	}
	if len(result.Imported) != 0 {
		t.Fatalf("expected nothing imported, got %v", result.Imported)
	}
	if len(result.Skipped) != 1 || result.Skipped[0].Name != "fs" || result.Skipped[0].Reason != "already in catalog" {
		t.Fatalf("expected fs skipped as already in catalog, got %v", result.Skipped)
	}
	s, _ := cat.FindServer("fs")
	if s.Command != "existing" {
		t.Fatalf("expected existing catalog entry untouched, got %v", s)
	}
}

func TestImportServers_PropagatesAdapterConflicts(t *testing.T) {
	cat := &catalog.Catalog{}
	fa := &fakeImportAdapter{
		fakeAdapter: newFakeAdapter(),
		candidates:  []catalog.Server{{Name: "x", Command: "x-server"}},
		conflicts:   []adapter.ImportConflict{{Name: "y", Reason: "duplicate across sources (kept .mcp.json)"}},
	}

	result, err := ImportServers(cat, fa, testScope())
	if err != nil {
		t.Fatalf("ImportServers failed: %v", err)
	}
	if len(result.Skipped) != 1 || result.Skipped[0].Name != "y" {
		t.Fatalf("expected adapter conflict propagated, got %v", result.Skipped)
	}
	if len(result.Imported) != 1 || result.Imported[0] != "x" {
		t.Fatalf("expected x still imported, got %v", result.Imported)
	}
}

func TestImportServers_NeverTouchesProfiles(t *testing.T) {
	cat := &catalog.Catalog{Profiles: []catalog.Profile{{Name: "work", Servers: []string{}}}}
	fa := &fakeImportAdapter{
		fakeAdapter: newFakeAdapter(),
		candidates:  []catalog.Server{{Name: "new", Command: "new-server"}},
	}

	if _, err := ImportServers(cat, fa, testScope()); err != nil {
		t.Fatalf("ImportServers failed: %v", err)
	}
	if len(cat.Profiles) != 1 || cat.Profiles[0].Name != "work" || len(cat.Profiles[0].Servers) != 0 {
		t.Fatalf("expected profiles untouched, got %v", cat.Profiles)
	}
}

func TestImportServers_AddsHTTPCandidateUnchanged(t *testing.T) {
	cat := &catalog.Catalog{}
	fa := &fakeImportAdapter{
		fakeAdapter: newFakeAdapter(),
		candidates: []catalog.Server{
			{Name: "aihq", Type: "http", URL: "https://aihq.example.com/mcp", Headers: map[string]string{"Authorization": "Bearer xyz"}},
		},
	}

	result, err := ImportServers(cat, fa, testScope())
	if err != nil {
		t.Fatalf("ImportServers failed: %v", err)
	}
	if len(result.Imported) != 1 || result.Imported[0] != "aihq" {
		t.Fatalf("expected aihq imported, got %v", result.Imported)
	}
	s, ok := cat.FindServer("aihq")
	if !ok || !s.IsHTTP() || s.URL != "https://aihq.example.com/mcp" {
		t.Fatalf("expected http server added unchanged, got %+v", s)
	}
}

func TestImportServers_AdapterErrorPropagates(t *testing.T) {
	cat := &catalog.Catalog{}
	wantErr := errors.New("boom")
	fa := &fakeImportAdapter{fakeAdapter: newFakeAdapter(), err: wantErr}

	_, err := ImportServers(cat, fa, testScope())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected adapter error propagated, got %v", err)
	}
	if len(cat.Servers) != 0 {
		t.Fatalf("expected catalog unchanged, got %v", cat.Servers)
	}
}
