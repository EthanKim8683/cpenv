package app

import (
	"errors"
	"fmt"
	"os"

	"github.com/EthanKim8683/cpenv/internal/state"
	"github.com/EthanKim8683/cpenv/internal/template"
	"github.com/EthanKim8683/cpenv/internal/workspace"
	"github.com/spf13/afero"
)

func (a *App) Focus(tmpl string) (string, error) {
	store := state.NewFileStore(a.Cfg.StatePath())

	st, err := store.Load()
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	problem, err := st.FocusedProblem()
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}
	if problem == nil {
		return "", fmt.Errorf("focus: no focused problem")
	}

	dir := a.Cfg.WorkspaceDir(problem.Id)

	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("focus: %w", err)
	}

	fs := afero.NewBasePathFs(afero.NewOsFs(), dir)

	if _, err := workspace.Create(fs, problem); err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	if tmpl == "" {
		tmpl = st.Template
		if tmpl == "" {
			return "", fmt.Errorf("focus: no template")
		}
	}

	src, err := os.ReadFile(tmpl)
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	if err := template.Render(fs, tmpl, src, problem); err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	return dir, nil
}
