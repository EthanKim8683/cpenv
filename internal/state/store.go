package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	path string
}

func (s *Store) Load() (*State, error) {
	var state State

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return &state, nil
	} else if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	if err = json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	return &state, nil
}

func (s *Store) Save(state *State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("save: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

func NewStore(path string) *Store {
	return &Store{path: path}
}
