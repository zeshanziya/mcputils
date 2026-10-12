package core

import (
	"github.com/zeshanziya/mcputils/internal/adapter"
	"github.com/zeshanziya/mcputils/internal/catalog"
	"github.com/zeshanziya/mcputils/internal/scope"
)

type ImportResult struct {
	Imported []string
	Skipped  []adapter.ImportConflict
}

func ImportServers(cat *catalog.Catalog, ad adapter.ClientAdapter, sc scope.Scope) (ImportResult, error) {
	candidates, conflicts, err := ad.ImportCandidates(sc)
	if err != nil {
		return ImportResult{}, err
	}

	result := ImportResult{}
	result.Skipped = append(result.Skipped, conflicts...)

	for _, candidate := range candidates {
		if _, ok := cat.FindServer(candidate.Name); ok {
			result.Skipped = append(result.Skipped, adapter.ImportConflict{
				Name:   candidate.Name,
				Reason: "already in catalog",
			})
			continue
		}
		cat.Servers = append(cat.Servers, candidate)
		result.Imported = append(result.Imported, candidate.Name)
	}

	return result, nil
}
