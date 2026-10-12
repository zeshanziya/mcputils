package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Name    string            `yaml:"name"`
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
}

type Profile struct {
	Name    string   `yaml:"name"`
	Servers []string `yaml:"servers,omitempty"`
}

type Catalog struct {
	Servers  []Server  `yaml:"servers,omitempty"`
	Profiles []Profile `yaml:"profiles,omitempty"`
}

func Load(path string) (*Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Catalog{}, nil
		}
		return nil, err
	}

	cat := &Catalog{}
	if err := yaml.Unmarshal(data, cat); err != nil {
		return nil, err
	}

	seenServers := map[string]bool{}
	for _, s := range cat.Servers {
		if seenServers[s.Name] {
			return nil, fmt.Errorf("duplicate server name: %s", s.Name)
		}
		seenServers[s.Name] = true
	}

	seenProfiles := map[string]bool{}
	for _, p := range cat.Profiles {
		if seenProfiles[p.Name] {
			return nil, fmt.Errorf("duplicate profile name: %s", p.Name)
		}
		seenProfiles[p.Name] = true
	}

	return cat, nil
}

func (c *Catalog) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (c *Catalog) FindServer(name string) (*Server, bool) {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			return &c.Servers[i], true
		}
	}
	return nil, false
}

func (c *Catalog) FindProfile(name string) (*Profile, bool) {
	for i := range c.Profiles {
		if c.Profiles[i].Name == name {
			return &c.Profiles[i], true
		}
	}
	return nil, false
}

func (c *Catalog) RemoveServer(name string) error {
	var referencing []string
	for _, p := range c.Profiles {
		for _, s := range p.Servers {
			if s == name {
				referencing = append(referencing, p.Name)
				break
			}
		}
	}
	if len(referencing) > 0 {
		return fmt.Errorf("server %s is referenced by profile(s): %s", name, strings.Join(referencing, ", "))
	}

	out := make([]Server, 0, len(c.Servers))
	for _, s := range c.Servers {
		if s.Name != name {
			out = append(out, s)
		}
	}
	c.Servers = out
	return nil
}
