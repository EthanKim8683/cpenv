package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	submitv1 "github.com/EthanKim8683/cpenv/internal/gen/submit/v1"
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

	// TODO: turn into streamlined client in extension package; will read file + return extension error
	res, err := c.Submitter.Submit(ctx, &submitv1.SubmitRequest{
		ProblemId: w.problem.GetId(),
		FileName:  filepath.Base(path),
		Content:   content,
	})
	if err != nil {
		return fmt.Errorf("submit %q: %w", path, err)
	}

	// TODO: change to extension error
	if errMsg := res.GetError(); errMsg != "" {
		return fmt.Errorf("submit %q: %s", path, errMsg)
	}
	return nil
}
