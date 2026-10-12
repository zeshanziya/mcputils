package scope

import (
	"errors"
	"os"
	"path/filepath"
)

type Scope struct {
	CatalogPath string
	StatePath   string
	IsGlobal    bool
	ProjectRoot string
}

func Resolve(globalFlag bool, projectFlag string, cwd string) (Scope, error) {
	if globalFlag && projectFlag != "" {
		return Scope{}, errors.New("--global and --project are mutually exclusive")
	}

	if globalFlag {
		home, err := os.UserHomeDir()
		if err != nil {
			return Scope{}, err
		}
		dir := filepath.Join(home, ".config", "mcputils")
		return Scope{
			CatalogPath: filepath.Join(dir, "servers.yaml"),
			StatePath:   filepath.Join(dir, "state.yaml"),
			IsGlobal:    true,
		}, nil
	}

	var root string
	if projectFlag != "" {
		root = projectFlag
	} else {
		found, err := findProjectRoot(cwd)
		if err != nil {
			return Scope{}, err
		}
		root = found
	}

	dir := filepath.Join(root, ".mcputils")
	return Scope{
		CatalogPath: filepath.Join(dir, "servers.yaml"),
		StatePath:   filepath.Join(dir, "state.yaml"),
		IsGlobal:    false,
		ProjectRoot: root,
	}, nil
}

func findProjectRoot(cwd string) (string, error) {
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".mcp.json")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no .mcp.json found in current directory or any parent; pass --global or --project")
		}
		dir = parent
	}
}
