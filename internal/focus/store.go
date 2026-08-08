package focus

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/focus/v1"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type FocusError struct {
	Message string
}

func (e *FocusError) Error() string {
	return fmt.Sprintf("focus: %s", e.Message)
}

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
		return nil, fmt.Errorf("load focus: %w", err)
	}

	focus := &focusv1.Focus{}
	if err := protojson.Unmarshal(data, focus); err != nil {
		return nil, fmt.Errorf("load focus: %w", err)
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
		return nil, &FocusError{Message: errMsg}
	}

	problem := focus.GetProblem()
	if problem == nil {
		return nil, errors.New("no problem")
	}
	return problem, nil
}

func validateFocus(focus *focusv1.Focus) error {
	if focus.GetError() != "" {
		return nil
	}

	if focus.GetProblem() == nil {
		return errors.New("validate: no problem")
	}
	return nil
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
	if err := validateFocus(focus); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	data, err := protojson.Marshal(focus)
	if err != nil {
		return fmt.Errorf("save: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := atomicWrite(s.path, data); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	return nil
}

func NewStore(path string) *Store {
	return &Store{path: path}
}
