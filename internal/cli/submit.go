package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

func (c *CLI) resolveSolution(name string) (string, error) {
	if path := name; filepath.IsAbs(path) {
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("resolve solution %q: %w", name, err)
		}
		return path, nil
	}

	if name != "" {
		path := filepath.Join(c.CWD, name)
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("resolve solution %q: %w", name, err)
		}
		return path, nil
	}

	matches, err := doublestar.FilepathGlob(filepath.Join(c.CWD, "**", "sol.*"))
	if err != nil {
		return "", fmt.Errorf("resolve solution: %w", err)
	}
	if len(matches) == 0 {
		return "", errors.New("resolve solution: no sol.* files")
	}
	if len(matches) > 1 {
		return "", errors.New("resolve solution: multiple sol.* files")
	}
	return matches[0], nil
}

func (c *CLI) Submit(ctx context.Context, name string) error {
	path, err := c.resolveSolution(name)
	if err != nil {
		return fmt.Errorf("submit %q: %w", name, err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("submit %q: %w", path, err)
	}

	w, err := openWorkspace(c.CWD)
	if err != nil {
		return fmt.Errorf("submit %q: %w", path, err)
	}
	defer w.close()

	if err := c.SubmitRequester.Request(
		ctx,
		w.problem.GetId(),
		filepath.Base(path),
		content,
	); err != nil {
		return fmt.Errorf("submit %q: %w", path, err)
	}
	return nil
}
