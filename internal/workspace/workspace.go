package workspace

import (
	"encoding/json"
	"sync"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/spf13/afero"
)

const statePath = "state.json"
const focusDir = "focus"
const envsDir = "envs"

type state struct {
	TemplateName string
}

type Workspace struct {
	mu              sync.Mutex
	templatesFs     afero.Fs
	workspaceFs     afero.Fs
	workspaceLinker afero.Linker
}

func (w *Workspace) Focus(problem *problemv1.Problem) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	env := NewEnv(afero.NewBasePathFs(w.workspaceFs, envsDir), problem)
	if exists, err := env.Exists(); err != nil {
		return err
	} else if !exists {
		data, err := afero.ReadFile(w.workspaceFs, statePath)
		if err != nil {
			return err
		}

		var state state
		if err := json.Unmarshal(data, &state); err != nil {
			return err
		}

		env.Init(w.templatesFs, state.TemplateName)
	}

	if err := w.workspaceFs.Remove(focusDir); err != nil {
		return err
	}

	if err := w.workspaceLinker.SymlinkIfPossible(envsDir, focusDir); err != nil {
		return err
	}

	return nil
}
