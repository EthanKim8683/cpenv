package cli

import (
	"errors"
	"fmt"
	"os"
)

func (c *CLI) Focus(templateName string) (string, error) {
	p, err := c.FocusedProblem.Load()
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	dir := c.workspaceDir(p.GetId())
	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("focus: %w", err)
	}

	w, err := createWorkspace(dir, p)
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	tPath, err := c.resolveTemplatePath(templateName)
	if err != nil {
		_ = w.close()
		return "", fmt.Errorf("focus: %w", err)
	}
	t := &template{path: tPath}

	if err := t.render(w.dir, p); err != nil {
		_ = w.close()
		return "", fmt.Errorf("focus: %w", err)
	}

	if err := w.close(); err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}
	return dir, nil
}
