package focus

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/proto/focus/v1"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/proto/problem/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type Store struct {
	mu   sync.RWMutex
	path string
}

func (s *Store) Load() (*focusv1.Focus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return &focusv1.Focus{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("load focus state: %w", err)
	}

	focus := &focusv1.Focus{}
	if err := protojson.Unmarshal(data, focus); err != nil {
		return nil, fmt.Errorf("load focus state: %w", err)
	}
	return focus, nil
}

func (s *Store) Problem() (*problemv1.Problem, error) {
	focus, err := s.Load()
	if err != nil {
		return nil, err
	}
	if focus == nil {
		return nil, errors.New("no focus")
	}

	if errMsg := focus.GetError(); errMsg != "" {
		// TODO: maybe error struct?
		return nil, errors.New(errMsg)
	}

	problem := focus.GetProblem()
	if problem == nil {
		return nil, errors.New("no problem")
	}
	return problem, nil
}

func atomicWrite(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp")
	if err != nil {
		return fmt.Errorf("atomic write %q: %w", path, err)
	}
	tmpPath := tmp.Name()

	_, err = tmp.Write(data)
	if err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("atomic write %q: %w", path, err)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("atomic write %q: %w", path, err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("atomic write %q: %w", path, err)
	}
	return nil
}

func (s *Store) save(focus *focusv1.Focus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := protojson.Marshal(focus)
	if err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	if err := atomicWrite(s.path, data); err != nil {
		return fmt.Errorf("save state: %w", err)
	}
	return nil
}

func NewStore(path string) *Store {
	return &Store{path: path}
}
