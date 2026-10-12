package adapter

import (
	"github.com/zeshanziya/mcputils/internal/catalog"
	"github.com/zeshanziya/mcputils/internal/scope"
)

type ClientAdapter interface {
	IsEnabled(sc scope.Scope, serverName string) (bool, error)
	SetEnabled(sc scope.Scope, serverName string, enabled bool) (sweptProjects int, err error)

	// ImportCandidates returns the servers currently defined at this client's
	// real config for sc, already deduplicated across whatever raw sources
	// that scope has. conflicts lists any same-name collision resolved while
	// deduplicating those sources.
	ImportCandidates(sc scope.Scope) (candidates []catalog.Server, conflicts []ImportConflict, err error)
}

type ImportConflict struct {
	Name   string
	Reason string
}
