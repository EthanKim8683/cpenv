package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type state struct {
	LastTemplatePath string
}

type stateStore interface {
	load() (*state, error)
	save(*state) error
}

type fileStateStore struct {
	path string
}

func (ss *fileStateStore) load() (*state, error) {
	data, err := os.ReadFile(ss.path)
	if errors.Is(err, os.ErrNotExist) {
		return &state{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}

	s := &state{}
	if err = json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}
	return s, nil
}

func (ss *fileStateStore) save(s *state) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	if err = os.MkdirAll(filepath.Dir(ss.path), 0755); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	if err = os.WriteFile(ss.path, data, 0644); err != nil {
		return fmt.Errorf("save state: %w", err)
	}
	return nil
}

var _ stateStore = (*fileStateStore)(nil)
