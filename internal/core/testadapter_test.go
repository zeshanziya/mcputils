package core

import (
	"github.com/zeshanziya/mcputils/internal/adapter"
	"github.com/zeshanziya/mcputils/internal/catalog"
	"github.com/zeshanziya/mcputils/internal/scope"
)

type fakeAdapter struct {
	enabled map[string]bool
	calls   []fakeCall
	sweep   map[string]int
}

type fakeCall struct {
	Server  string
	Enabled bool
}

func newFakeAdapter() *fakeAdapter {
	return &fakeAdapter{enabled: map[string]bool{}}
}

func (f *fakeAdapter) IsEnabled(sc scope.Scope, serverName string) (bool, error) {
	v, ok := f.enabled[serverName]
	if !ok {
		// Mirrors the real Claude Code adapter's missing-entry defaults
		// (internal/adapter/claudecode): project scope defaults to enabled
		// (nothing in disabledMcpServers), global scope defaults to
		// not-enabled (no mcpServers entry yet).
		return !sc.IsGlobal, nil
	}
	return v, nil
}

func (f *fakeAdapter) SetEnabled(sc scope.Scope, serverName string, enabled bool) (int, error) {
	f.enabled[serverName] = enabled
	f.calls = append(f.calls, fakeCall{Server: serverName, Enabled: enabled})
	swept := 0
	if f.sweep != nil {
		swept = f.sweep[serverName]
	}
	return swept, nil
}

func (f *fakeAdapter) wasCalledWith(serverName string, enabled bool) bool {
	for _, c := range f.calls {
		if c.Server == serverName && c.Enabled == enabled {
			return true
		}
	}
	return false
}

func (f *fakeAdapter) ImportCandidates(sc scope.Scope) ([]catalog.Server, []adapter.ImportConflict, error) {
	return nil, nil, nil
}

func (f *fakeAdapter) wasCalledFor(serverName string) bool {
	for _, c := range f.calls {
		if c.Server == serverName {
			return true
		}
	}
	return false
}
