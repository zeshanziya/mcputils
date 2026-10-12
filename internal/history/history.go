package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	Timestamp     time.Time `json:"timestamp"`
	Scope         string    `json:"scope"`
	Trigger       string    `json:"trigger"`
	Server        string    `json:"server"`
	State         string    `json:"state"`
	SweptProjects int       `json:"swept_projects,omitempty"`
}

type Logger struct {
	path string
}

func New(path string) *Logger {
	return &Logger{path: path}
}

func (l *Logger) Append(scope, trigger, server, state string, sweptProjects int) error {
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return err
	}

	entry := Entry{
		Timestamp:     time.Now().UTC(),
		Scope:         scope,
		Trigger:       trigger,
		Server:        server,
		State:         state,
		SweptProjects: sweptProjects,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

func (l *Logger) List() ([]Entry, error) {
	f, err := os.Open(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(lines))
	for i, line := range lines {
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			if i == len(lines)-1 {
				break
			}
			return nil, fmt.Errorf("malformed history entry at line %d: %w", i+1, err)
		}
		entries = append(entries, e)
	}

	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries, nil
}
