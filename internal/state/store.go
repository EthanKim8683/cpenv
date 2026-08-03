package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

type Store interface {
	Load() (*State, error)
	Save(state *State) error
}

type fileStore struct {
	lock *flock.Flock
	path string
}

func (s *fileStore) Load() (*State, error) {
	s.lock.RLock()
	defer s.lock.Unlock()

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

func (s *fileStore) Save(state *State) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

func NewFileStore(path string) Store {
	return &fileStore{
		path: path,
		lock: flock.New(path + ".lock"),
	}
}
