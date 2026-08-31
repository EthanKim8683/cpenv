package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/focus/v1"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
)

func (c *CLI) focusedProblem(ctx context.Context) (*problemv1.Problem, error) {
	res, err := c.FocusClient.Load(ctx, &focusv1.LoadRequest{})
	if err != nil {
		return nil, fmt.Errorf("focus: %w", err)
	}

	focus := res.GetFocus()
	if focus == nil {
		return nil, nil
	}

	if focus.Error != nil {
		return nil, fmt.Errorf("focus: extension error: %s", focus.GetError())
	}
	return focus.GetProblem(), nil
}

func (c *CLI) Focus(ctx context.Context, templateName string) (string, error) {
	problem, err := c.focusedProblem(ctx)
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}
	if problem == nil {
		return "", fmt.Errorf("focus: no focused problem")
	}

	dir := c.workspaceDir(problem.GetId())
	_, err = openWorkspace(dir)
	if err == nil {
		return dir, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("focus: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}
	err = os.Mkdir(dir, 0755)
	if errors.Is(err, os.ErrExist) {
		if _, err = initWorkspace(dir, problem); err != nil {
			return "", fmt.Errorf("focus: %w", err)
		}
		return dir, nil
	}
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	if _, err = initWorkspace(dir, problem); err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	defaultTemplate, err := c.Preferences.DefaultTemplate()
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	t, err := resolveTemplate(templateName, c.CWD, c.templatesDir(), defaultTemplate)
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}
	if t == nil {
		return dir, nil
	}
	if err := t.render(dir, problem); err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	if err := c.Preferences.SetDefaultTemplate(t.path); err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}
	return dir, nil
}
