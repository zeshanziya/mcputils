package core

import (
	"fmt"

	"github.com/zeshanziya/mcputils/internal/adapter"
	"github.com/zeshanziya/mcputils/internal/catalog"
	"github.com/zeshanziya/mcputils/internal/history"
	"github.com/zeshanziya/mcputils/internal/scope"
)

func EnableServer(cat *catalog.Catalog, ad adapter.ClientAdapter, hist *history.Logger, sc scope.Scope, serverName string) error {
	if _, ok := cat.FindServer(serverName); !ok {
		return fmt.Errorf("unknown server: %s", serverName)
	}
	wasEnabled, err := ad.IsEnabled(sc, serverName)
	if err != nil {
		return err
	}
	sweptProjects, err := ad.SetEnabled(sc, serverName, true)
	if err != nil {
		return err
	}
	if wasEnabled {
		return nil
	}
	return hist.Append(scopeLabel(sc), "server", serverName, "enabled", sweptProjects)
}

func DisableServer(cat *catalog.Catalog, ad adapter.ClientAdapter, hist *history.Logger, sc scope.Scope, serverName string) error {
	if _, ok := cat.FindServer(serverName); !ok {
		return fmt.Errorf("unknown server: %s", serverName)
	}
	wasEnabled, err := ad.IsEnabled(sc, serverName)
	if err != nil {
		return err
	}
	sweptProjects, err := ad.SetEnabled(sc, serverName, false)
	if err != nil {
		return err
	}
	if !wasEnabled {
		return nil
	}
	return hist.Append(scopeLabel(sc), "server", serverName, "disabled", sweptProjects)
}
