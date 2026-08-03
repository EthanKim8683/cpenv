package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/EthanKim8683/cpenv/internal/template"
	"github.com/EthanKim8683/cpenv/internal/workspace"
	"github.com/spf13/afero"
)

func (a *App) focusedProblem() (*problemv1.Problem, error) {
	state, err := a.StateStore.Load()
	if err != nil {
		return nil, err
	}

	problem, err := state.FocusedProblem()
	if err != nil {
		return nil, err
	}

	return problem, nil
}

func (a *App) defaultTemplate() (string, error) {
	state, err := a.StateStore.Load()
	if err != nil {
		return "", err
	}

	if tmpl := state.Template; tmpl != "" {
		return tmpl, nil
	}

	matches, err := filepath.Glob(filepath.Join(a.Cfg.Home, "templates", "*.star"))
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no templates found")
	}

	return matches[0], nil
}

func (a *App) resolveTemplate(tmpl string) (string, error) {
	if tmpl == "" {
		tmpl, err := a.defaultTemplate()
		if err != nil {
			return "", err
		}

		return tmpl, nil
	}

	if filepath.IsAbs(tmpl) {
		return tmpl, nil
	}

	return filepath.Join(a.Cfg.Home, "templates", tmpl), nil
}

func (a *App) saveTemplate(tmpl string) error {
	state, err := a.StateStore.Load()
	if err != nil {
		return err
	}

	state.Template = tmpl

	if err := a.StateStore.Save(state); err != nil {
		return err
	}

	return nil
}

func (a *App) renderTemplate(fs afero.Fs, tmpl string, problem *problemv1.Problem) error {
	src, err := os.ReadFile(tmpl)
	if err != nil {
		return err
	}

	if err := template.Render(fs, tmpl, src, problem); err != nil {
		return err
	}

	if err = a.saveTemplate(tmpl); err != nil {
		return err
	}

	return nil
}

func (a *App) Focus(tmpl string) (string, error) {
	problem, err := a.focusedProblem()
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	dir := filepath.Join(a.Cfg.Home, "workspaces", problem.Id)

	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("focus: %w", err)
	}

	fs := afero.NewBasePathFs(afero.NewOsFs(), dir)

	if _, err := workspace.Create(fs, problem); err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	tmpl, err = a.resolveTemplate(tmpl)
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	if err := a.renderTemplate(fs, tmpl, problem); err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	return dir, nil
}
