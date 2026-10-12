package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zeshanziya/mcputils/internal/adapter"
	"github.com/zeshanziya/mcputils/internal/adapter/claudecode"
	"github.com/zeshanziya/mcputils/internal/catalog"
	"github.com/zeshanziya/mcputils/internal/history"
	"github.com/zeshanziya/mcputils/internal/scope"
	"github.com/zeshanziya/mcputils/internal/state"
)

func resolveScope() (scope.Scope, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return scope.Scope{}, err
	}
	return scope.Resolve(globalFlag, projectFlag, cwd)
}

func loadCatalog(sc scope.Scope) (*catalog.Catalog, error) {
	return catalog.Load(sc.CatalogPath)
}

func loadState(sc scope.Scope) (*state.State, error) {
	return state.Load(sc.StatePath)
}

func newAdapter(cat *catalog.Catalog) adapter.ClientAdapter {
	return claudecode.New(cat)
}

func newHistoryLogger() *history.Logger {
	home, _ := os.UserHomeDir()
	return history.New(filepath.Join(home, ".config", "mcputils", "history.jsonl"))
}

func joinNames(names []string) string {
	return strings.Join(names, ", ")
}

func joinSkipped(skipped []adapter.ImportConflict) string {
	parts := make([]string, len(skipped))
	for i, s := range skipped {
		parts[i] = fmt.Sprintf("%s (%s)", s.Name, s.Reason)
	}
	return strings.Join(parts, ", ")
}
