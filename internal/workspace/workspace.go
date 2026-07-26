package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/spf13/afero"
)

const statePath = "state.json"
const metadataPath = "metadata.json"
const focusDir = "focus"
const envsDir = "envs"

type state struct {
	TemplateName string
}

type metadata struct {
	problem *problemv1.Problem
}

type Workspace struct {
	mu              sync.Mutex
	templatesFs     afero.Fs
	workspaceFs     afero.Fs
	workspaceLinker afero.Linker
}

func (w *Workspace) state() (*state, error) {
	content, err := afero.ReadFile(w.workspaceFs, statePath)
	if err != nil {
		return nil, err
	}

	var state state
	if err := json.Unmarshal(content, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func (w *Workspace) Focus(problem *problemv1.Problem) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	path := filepath.Join(envsDir, problem.Id)
	if _, err := w.workspaceFs.Stat(path); os.IsNotExist(err) {
		if err := w.workspaceFs.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("mkdir: %w", problem.Id, err)
		}

		// get state
		state := &state{}

		if err := initEnv(
			afero.NewBasePathFs(w.workspaceFs, path),
			w.templatesFs,
			state.TemplateName,
			problem,
		); err != nil {
			return fmt.Errorf("focus %s: %w", problem.Id, err)
		}

		// write metadata
		metadata := &metadata{
			problem: problem,
		}
	} else if err != nil {
		return fmt.Errorf("focus %s: stat: %w", problem.Id, err)
	}

	if err := w.workspaceFs.Remove(path); err != nil {
		return fmt.Errorf("focus %s: remove: %w", problem.Id, err)
	}

	if err := w.workspaceLinker.SymlinkIfPossible(path, path); err != nil {
		return fmt.Errorf("focus %s: symlink: %w", problem.Id, err)
	}

	return nil
}

func (w *Workspace) Init(templatePath string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := w.workspaceFs.Stat(focusDir); err != nil {
		return fmt.Errorf("init: stat %q: %w", focusDir, err)
	}

	// read metadata
	metadata := &metadata{}

	if err := initEnv(
		afero.NewBasePathFs(w.workspaceFs, focusDir),
		w.templatesFs,
		templatePath,
		metadata.problem,
	); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	return nil
}
