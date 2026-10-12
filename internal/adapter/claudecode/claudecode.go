package claudecode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zeshanziya/mcputils/internal/catalog"
	"github.com/zeshanziya/mcputils/internal/jsonutil"
	"github.com/zeshanziya/mcputils/internal/scope"
)

type Adapter struct {
	cat  *catalog.Catalog
	path string
}

func New(cat *catalog.Catalog) *Adapter {
	home, _ := os.UserHomeDir()
	return &Adapter{cat: cat, path: filepath.Join(home, ".claude.json")}
}

func readConfig(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, nil
		}
		return nil, err
	}
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (a *Adapter) IsEnabled(sc scope.Scope, serverName string) (bool, error) {
	m, err := readConfig(a.path)
	if err != nil {
		return false, err
	}

	if sc.IsGlobal {
		servers, _ := m["mcpServers"].(map[string]interface{})
		_, ok := servers[serverName]
		return ok, nil
	}

	projects, _ := m["projects"].(map[string]interface{})
	proj, _ := projects[sc.ProjectRoot].(map[string]interface{})
	arr, _ := proj["disabledMcpServers"].([]interface{})
	for _, v := range arr {
		if v == serverName {
			return false, nil
		}
	}
	return true, nil
}

func (a *Adapter) SetEnabled(sc scope.Scope, serverName string, enabled bool) (int, error) {
	if sc.IsGlobal {
		return a.setEnabledGlobal(serverName, enabled)
	}
	return a.setEnabledProject(sc.ProjectRoot, serverName, enabled)
}

func (a *Adapter) setEnabledProject(projectRoot, serverName string, enabled bool) (int, error) {
	err := jsonutil.PatchNestedStringArray(a.path, "projects", projectRoot, "disabledMcpServers", func(current []string) []string {
		if enabled {
			out := make([]string, 0, len(current))
			for _, v := range current {
				if v != serverName {
					out = append(out, v)
				}
			}
			return out
		}
		for _, v := range current {
			if v == serverName {
				return current
			}
		}
		return append(current, serverName)
	})
	return 0, err
}

func (a *Adapter) setEnabledGlobal(serverName string, enabled bool) (int, error) {
	if !enabled {
		err := jsonutil.PatchMapEntry(a.path, "mcpServers", serverName, nil)
		return 0, err
	}

	server, ok := a.cat.FindServer(serverName)
	if !ok {
		return 0, fmt.Errorf("unknown server: %s", serverName)
	}

	entry := map[string]interface{}{
		"command": server.Command,
		"args":    server.Args,
		"env":     server.Env,
	}
	if err := jsonutil.PatchMapEntry(a.path, "mcpServers", serverName, entry); err != nil {
		return 0, err
	}

	swept, err := jsonutil.SweepNestedStringArrays(a.path, "projects", "disabledMcpServers", serverName)
	if err != nil {
		return 0, err
	}
	return swept, nil
}
