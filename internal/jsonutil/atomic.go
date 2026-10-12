package jsonutil

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func WriteAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}

func readMap(path string) (map[string]interface{}, error) {
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

func toStringSlice(v interface{}) []string {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func fromStringSlice(s []string) []interface{} {
	out := make([]interface{}, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}

func writeMap(path string, m map[string]interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return WriteAtomic(path, data)
}

func PatchStringArray(path string, key string, mutate func(current []string) []string) error {
	m, err := readMap(path)
	if err != nil {
		return err
	}
	current := toStringSlice(m[key])
	m[key] = fromStringSlice(mutate(current))
	return writeMap(path, m)
}

func PatchNestedStringArray(path, outerKey, outerSubKey, innerKey string, mutate func(current []string) []string) error {
	m, err := readMap(path)
	if err != nil {
		return err
	}
	outer, ok := m[outerKey].(map[string]interface{})
	if !ok {
		outer = map[string]interface{}{}
	}
	sub, ok := outer[outerSubKey].(map[string]interface{})
	if !ok {
		sub = map[string]interface{}{}
	}
	current := toStringSlice(sub[innerKey])
	sub[innerKey] = fromStringSlice(mutate(current))
	outer[outerSubKey] = sub
	m[outerKey] = outer
	return writeMap(path, m)
}

func PatchMapEntry(path, mapKey, entryKey string, value interface{}) error {
	m, err := readMap(path)
	if err != nil {
		return err
	}
	inner, ok := m[mapKey].(map[string]interface{})
	if !ok {
		inner = map[string]interface{}{}
	}
	if value == nil {
		delete(inner, entryKey)
	} else {
		inner[entryKey] = value
	}
	m[mapKey] = inner
	return writeMap(path, m)
}

func SweepNestedStringArrays(path, outerKey, innerKey, name string) (int, error) {
	m, err := readMap(path)
	if err != nil {
		return 0, err
	}
	outer, ok := m[outerKey].(map[string]interface{})
	if !ok {
		return 0, nil
	}
	sweptCount := 0
	for subKey, rawSub := range outer {
		sub, ok := rawSub.(map[string]interface{})
		if !ok {
			continue
		}
		current := toStringSlice(sub[innerKey])
		removed := false
		out := make([]string, 0, len(current))
		for _, v := range current {
			if v == name {
				removed = true
				continue
			}
			out = append(out, v)
		}
		if removed {
			sub[innerKey] = fromStringSlice(out)
			outer[subKey] = sub
			sweptCount++
		}
	}
	if sweptCount == 0 {
		return 0, nil
	}
	m[outerKey] = outer
	if err := writeMap(path, m); err != nil {
		return 0, err
	}
	return sweptCount, nil
}
