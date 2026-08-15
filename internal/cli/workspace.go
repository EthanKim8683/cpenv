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

func createWorkspace(dir string, problem *problemv1.Problem) (*workspace, error) {
	data, err := protojson.Marshal(problem)
	if err != nil {
		return nil, fmt.Errorf("create workspace %q: %w", dir, err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create workspace %q: %w", dir, err)
	}

	if err := os.WriteFile(filepath.Join(dir, problemFile), data, 0644); err != nil {
		return nil, fmt.Errorf("create workspace %q: %w", dir, err)
	}
	return &workspace{dir: dir, problem: problem}, nil
}

func (c *CLI) workspacesDir() string {
	return filepath.Join(c.Cfg.HomeDir, "workspaces")
}

func (c *CLI) workspaceDir(problemID string) string {
	return filepath.Join(c.workspacesDir(), problemID)
}
