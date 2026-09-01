package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

const problemFile = ".problem.json"

type workspace struct {
	dir     string
	problem *problemv1.Problem
}

func (w *workspace) clear() error {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return fmt.Errorf("clear workspace %q: %w", w.dir, err)
	}

	var errs error
	for _, entry := range entries {
		if entry.Name() == problemFile {
			continue
		}
		errs = errors.Join(errs, os.RemoveAll(filepath.Join(w.dir, entry.Name())))
	}
	if errs != nil {
		return fmt.Errorf("clear workspace %q: %w", w.dir, errs)
	}
	return nil
}

func openWorkspace(dir string) (*workspace, error) {
	data, err := os.ReadFile(filepath.Join(dir, problemFile))
	if err != nil {
		return nil, fmt.Errorf("open workspace %q: %w", dir, err)
	}

	problem := &problemv1.Problem{}
	if err := protojson.Unmarshal(data, problem); err != nil {
		return nil, fmt.Errorf("open workspace %q: %w", dir, err)
	}
	return &workspace{dir: dir, problem: problem}, nil
}

func initWorkspace(dir string, problem *problemv1.Problem) (*workspace, error) {
	data, err := protojson.Marshal(problem)
	if err != nil {
		return nil, fmt.Errorf("create workspace %q: %w", dir, err)
	}

	if err := os.WriteFile(filepath.Join(dir, problemFile), data, 0644); err != nil {
		return nil, fmt.Errorf("create workspace %q: %w", dir, err)
	}
	return &workspace{dir: dir, problem: problem}, nil
}

func ensureWorkspace(dir string, problem *problemv1.Problem) (*workspace, bool, error) {
	w, err := openWorkspace(dir)
	if err == nil {
		return w, false, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, false, err
	}

	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return nil, false, err
	}
	err = os.Mkdir(dir, 0755)
	if err == nil {
		if w, err := initWorkspace(dir, problem); err == nil {
			return w, true, nil
		}
		return nil, false, err
	}
	if errors.Is(err, os.ErrExist) {
		if w, err := initWorkspace(dir, problem); err == nil {
			return w, false, nil
		}
		return nil, false, err
	}
	return nil, false, err
}
