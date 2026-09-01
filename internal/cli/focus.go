package cli

import (
	"context"
	"errors"
	"fmt"

	activeproblemv1 "github.com/EthanKim8683/cpenv/internal/gen/active_problem/v1"
	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/focus/v1"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
)

func (c *CLI) activeProblem(ctx context.Context) (*problemv1.Problem, error) {
	res, err := c.ActiveProblemClient.Load(ctx, &activeproblemv1.LoadRequest{})
	if err != nil {
		return nil, err
	}

	activeProblem := res.GetActiveProblem()
	if activeProblem == nil {
		return nil, errors.New("no active problem")
	}

	if activeProblem.Error != nil {
		return nil, fmt.Errorf("extension error: %s", activeProblem.GetError())
	}
	return activeProblem.GetProblem(), nil
}

func (c *CLI) focus(ctx context.Context, dir string) error {
	if c.Cfg.EnvId == nil {
		return nil
	}

	res, err := c.FocusClient.Focus(ctx, &focusv1.FocusRequest{
		EnvId: *c.Cfg.EnvId,
		Dir:   dir,
	})
	if err != nil {
		return err
	}

	if res.Error != nil {
		return fmt.Errorf("environment error: %s", res.GetError())
	}
	return nil
}

func (c *CLI) Focus(ctx context.Context, templateName string) (string, error) {
	problem, err := c.activeProblem(ctx)
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	dir := c.workspaceDir(problem.GetId())
	_, created, err := ensureWorkspace(dir, problem)
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	if created {
		t, err := c.resolveTemplate(templateName)
		if err != nil {
			return "", fmt.Errorf("focus: %w", err)
		}

		if err := t.render(dir, problem); err != nil {
			return "", fmt.Errorf("focus: %w", err)
		}

		if err := c.Preferences.SetDefaultTemplate(t.path); err != nil {
			return "", fmt.Errorf("focus: %w", err)
		}
	}

	if err := c.focus(ctx, dir); err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}
	return dir, nil
}
