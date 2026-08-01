package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Store interface {
	Load() (*State, error)
	Save(state *State) error
}

type fileStore struct {
	path string
}

func (s *fileStore) Load() (*State, error) {
	var state State

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return &state, nil
	} else if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if err = json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return &state, nil
}

func (s *fileStore) Save(state *State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return fmt.Errorf("store: %w", err)
	}

	f, err := os.CreateTemp(filepath.Dir(s.path), "*")
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}
	tmpPath := f.Name()

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("store: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("store: %w", err)
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("store: %w", err)
	}

	return nil
}

func NewFileStore(path string) *fileStore {
	return &fileStore{path: path}
}
