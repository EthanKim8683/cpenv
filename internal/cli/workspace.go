package cli

import (
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
	closed  bool
}

func (w *workspace) close() error {
	if w.closed {
		return nil
	}

	data, err := protojson.Marshal(w.problem)
	if err != nil {
		return fmt.Errorf("close workspace %q: %w", w.dir, err)
	}

	if err := os.WriteFile(filepath.Join(w.dir, problemFile), data, 0644); err != nil {
		return fmt.Errorf("close workspace %q: %w", w.dir, err)
	}

	w.closed = true
	return nil
}

func openWorkspace(dir string) (*workspace, error) {
	data, err := os.ReadFile(filepath.Join(dir, problemFile))
	if err != nil {
		return nil, fmt.Errorf("open workspace %q: %w", dir, err)
	}

	p := &problemv1.Problem{}
	if err := protojson.Unmarshal(data, p); err != nil {
		return nil, fmt.Errorf("open workspace %q: %w", dir, err)
	}
	return &workspace{dir: dir, problem: p}, nil
}

func createWorkspace(dir string, problem *problemv1.Problem) (*workspace, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
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
