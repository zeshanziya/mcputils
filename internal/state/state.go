package state

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type State struct {
	ActiveProfiles []string `yaml:"active_profiles"`
}

func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, err
	}
	st := &State{}
	if err := yaml.Unmarshal(data, st); err != nil {
		return nil, err
	}
	return st, nil
}

func (s *State) Save(path string) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (s *State) IsActive(profileName string) bool {
	for _, p := range s.ActiveProfiles {
		if p == profileName {
			return true
		}
	}
	return false
}

func (s *State) SetActive(profileName string, active bool) {
	if active {
		if s.IsActive(profileName) {
			return
		}
		s.ActiveProfiles = append(s.ActiveProfiles, profileName)
		return
	}
	out := make([]string, 0, len(s.ActiveProfiles))
	for _, p := range s.ActiveProfiles {
		if p != profileName {
			out = append(out, p)
		}
	}
	s.ActiveProfiles = out
}
