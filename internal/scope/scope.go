package scope

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"

	"github.com/zeshanziya/mcputils/internal/jsonutil"
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

	home, err := os.UserHomeDir()
	if err != nil {
		return Scope{}, err
	}
	configDir := filepath.Join(home, ".config", "mcputils")
	catalogPath := filepath.Join(configDir, "servers.yaml")

	if globalFlag {
		return Scope{
			CatalogPath: catalogPath,
			StatePath:   filepath.Join(configDir, "state", "global.yaml"),
			IsGlobal:    true,
		}, nil
	}

	var root string
	if projectFlag != "" {
		root = projectFlag
	} else {
		found, err := detectProjectRoot(cwd)
		if err != nil {
			return Scope{}, err
		}
		root = found
	}

	return Scope{
		CatalogPath: catalogPath,
		StatePath:   filepath.Join(configDir, "state", stateFileName(root)),
		IsGlobal:    false,
		ProjectRoot: root,
	}, nil
}

func stateFileName(root string) string {
	sum := sha256.Sum256([]byte(root))
	return hex.EncodeToString(sum[:])[:16] + ".yaml"
}

func detectProjectRoot(cwd string) (string, error) {
	if _, err := os.Stat(filepath.Join(cwd, ".mcp.json")); err == nil {
		return cwd, nil
	}

	if home, err := os.UserHomeDir(); err == nil {
		claudePath := filepath.Join(home, ".claude.json")
		if projects, err := jsonutil.ReadMapEntries(claudePath, "projects"); err == nil {
			if _, ok := projects[cwd]; ok {
				return cwd, nil
			}
		}
	}

	return "", errors.New("no .mcp.json and no ~/.claude.json project entry found for the current directory; pass --global or --project")
}
