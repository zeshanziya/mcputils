package adapter

import "github.com/zeshanziya/mcputils/internal/scope"

type ClientAdapter interface {
	IsEnabled(sc scope.Scope, serverName string) (bool, error)
	SetEnabled(sc scope.Scope, serverName string, enabled bool) (sweptProjects int, err error)
}
