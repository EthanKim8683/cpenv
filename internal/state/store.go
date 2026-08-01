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
		return nil, fmt.Errorf("load state: %w", err)
	}

	if err = json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}

	return &state, nil
}

func atomicWrite(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return nil
}

func (s *fileStore) Save(state *State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	if err := atomicWrite(s.path, data); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

func NewFileStore(path string) Store {
	return &fileStore{path: path}
}
