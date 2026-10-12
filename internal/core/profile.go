package core

import (
	"fmt"

	"github.com/zeshanziya/mcputils/internal/adapter"
	"github.com/zeshanziya/mcputils/internal/catalog"
	"github.com/zeshanziya/mcputils/internal/history"
	"github.com/zeshanziya/mcputils/internal/scope"
	"github.com/zeshanziya/mcputils/internal/state"
)

func CreateProfile(cat *catalog.Catalog, name string) error {
	if _, ok := cat.FindProfile(name); ok {
		return fmt.Errorf("profile already exists: %s", name)
	}
	cat.Profiles = append(cat.Profiles, catalog.Profile{Name: name})
	return nil
}

func AddServerToProfile(cat *catalog.Catalog, profileName, serverName string) error {
	if _, ok := cat.FindServer(serverName); !ok {
		return fmt.Errorf("unknown server: %s", serverName)
	}
	p, ok := cat.FindProfile(profileName)
	if !ok {
		return fmt.Errorf("unknown profile: %s", profileName)
	}
	for _, s := range p.Servers {
		if s == serverName {
			return nil
		}
	}
	p.Servers = append(p.Servers, serverName)
	return nil
}

func RemoveServerFromProfile(cat *catalog.Catalog, profileName, serverName string) error {
	p, ok := cat.FindProfile(profileName)
	if !ok {
		return fmt.Errorf("unknown profile: %s", profileName)
	}
	out := make([]string, 0, len(p.Servers))
	for _, s := range p.Servers {
		if s != serverName {
			out = append(out, s)
		}
	}
	p.Servers = out
	return nil
}

func EnableProfile(cat *catalog.Catalog, st *state.State, ad adapter.ClientAdapter, hist *history.Logger, sc scope.Scope, profileName string) error {
	p, ok := cat.FindProfile(profileName)
	if !ok {
		return fmt.Errorf("unknown profile: %s", profileName)
	}

	for _, serverName := range p.Servers {
		if _, ok := cat.FindServer(serverName); !ok {
			return fmt.Errorf("profile %s references unknown server: %s", profileName, serverName)
		}
	}

	scopeLabel := scopeLabel(sc)
	trigger := "profile:" + profileName
	for _, serverName := range p.Servers {
		wasEnabled, err := ad.IsEnabled(sc, serverName)
		if err != nil {
			return err
		}
		sweptProjects, err := ad.SetEnabled(sc, serverName, true)
		if err != nil {
			return err
		}
		if wasEnabled {
			continue
		}
		if err := hist.Append(scopeLabel, trigger, serverName, "enabled", sweptProjects); err != nil {
			return err
		}
	}

	st.SetActive(profileName, true)
	return nil
}

func DisableProfile(cat *catalog.Catalog, st *state.State, ad adapter.ClientAdapter, hist *history.Logger, sc scope.Scope, profileName string) error {
	p, ok := cat.FindProfile(profileName)
	if !ok {
		return fmt.Errorf("unknown profile: %s", profileName)
	}

	scopeLabel := scopeLabel(sc)
	trigger := "profile:" + profileName
	for _, serverName := range p.Servers {
		if sharedWithOtherActiveProfile(cat, st, profileName, serverName) {
			continue
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
			continue
		}
		if err := hist.Append(scopeLabel, trigger, serverName, "disabled", sweptProjects); err != nil {
			return err
		}
	}

	st.SetActive(profileName, false)
	return nil
}

func sharedWithOtherActiveProfile(cat *catalog.Catalog, st *state.State, excludeProfile, serverName string) bool {
	for _, activeName := range st.ActiveProfiles {
		if activeName == excludeProfile {
			continue
		}
		p, ok := cat.FindProfile(activeName)
		if !ok {
			continue
		}
		for _, s := range p.Servers {
			if s == serverName {
				return true
			}
		}
	}
	return false
}

func scopeLabel(sc scope.Scope) string {
	if sc.IsGlobal {
		return "global"
	}
	return sc.ProjectRoot
}
