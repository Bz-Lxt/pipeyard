package config

import (
	"path/filepath"
	"strings"
)

func (c Config) Join(name string) string {
	name = filepath.Base(name)
	return filepath.Join(c.Dir, name)
}

func (c Config) IsInside(path string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	root, err := filepath.Abs(c.Dir)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func (c Config) SnapshotName() string {
	return filepath.Join(c.Dir, "yard.snapshot.json")
}

func (c Config) EventsName() string {
	return filepath.Join(c.Dir, "events.jsonl")
}
