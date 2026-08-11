package cli

import (
	"errors"
	"path/filepath"
)

func (c *CLI) workspacesDir() string {
	return filepath.Join(c.Cfg.HomeDir, "workspaces")
}

func (c *CLI) workspaceDir(problemID string) string {
	return filepath.Join(c.workspacesDir(), problemID)
}

func (c *CLI) templatesDir() string {
	return filepath.Join(c.Cfg.HomeDir, "templates")
}

func (c *CLI) resolveTemplatePath(name string) (string, error) {
	if filepath.IsAbs(name) {
		return name, nil
	}
	if name != "" {
		return filepath.Join(c.templatesDir(), name), nil
	}
	return "", errors.New("not implemented")
}

func (c *CLI) resolveSolutionPath(name string) (string, error) {
	if filepath.IsAbs(name) {
		return name, nil
	}
	if name != "" {
		return filepath.Join(c.Cwd, name), nil
	}
	return "", errors.New("not implemented")
}
