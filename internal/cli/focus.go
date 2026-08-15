package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/focus/v1"
)

func (c *CLI) Focus(ctx context.Context, templateName string) (string, error) {
	res, err := c.FocusClient.Load(ctx, &focusv1.LoadRequest{})
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}
	focus := res.GetFocus()
	if focus.Error != nil {
		return "", fmt.Errorf("focus: extension error: %s", focus.GetError())
	}
	problem := focus.GetProblem()
	if problem == nil {
		return "", errors.New("focus: no problem")
	}

	dir := c.workspaceDir(problem.GetId())
	_, err = openWorkspace(dir)
	if err == nil {
		return dir, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("focus: %w", err)
	}
	if _, err = createWorkspace(dir, problem); err != nil {
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
