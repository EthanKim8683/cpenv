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
	Update(fn func(*State) error) error
}

type fileStore struct {
	lock *flock.Flock
	path string
}

func (s *fileStore) load() (*State, error) {
	var st State

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return &st, nil
	} else if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &st); err != nil {
		return nil, err
	}

	return &st, nil
}

func (s *fileStore) save(st *State) error {
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return err
	}

	return nil
}

func (s *fileStore) Load() (*State, error) {
	s.lock.RLock()
	defer s.lock.Unlock()

	st, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}

	return st, nil
}

func (s *fileStore) Update(fn func(*State) error) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	st, err := s.load()
	if err != nil {
		return fmt.Errorf("update state: load: %w", err)
	}

	if err := fn(st); err != nil {
		return fmt.Errorf("update state: %w", err)
	}

	if err := s.save(st); err != nil {
		return fmt.Errorf("update state: save: %w", err)
	}

	return nil
}

func NewFileStore(path string) Store {
	return &fileStore{
		path: path,
		lock: flock.New(path + ".lock"),
	}
}
