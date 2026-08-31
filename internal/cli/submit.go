package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	submitv1 "github.com/EthanKim8683/cpenv/internal/gen/submit/v1"
)

type solution struct {
	path    string
	content []byte
}

func resolveSolution(cwd string, name string) (*solution, error) {
	if filepath.IsAbs(name) {
		content, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}
		return &solution{
			path:    name,
			content: content,
		}, nil
	}

	if name != "" {
		path := filepath.Join(cwd, name)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return &solution{
			path:    path,
			content: content,
		}, nil
	}

	matches, err := filepath.Glob(filepath.Join(cwd, "sol.*"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, errors.New("no sol.* files")
	}
	if len(matches) > 1 {
		return nil, errors.New("multiple sol.* files")
	}
	path := matches[0]
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &solution{
		path:    path,
		content: content,
	}, nil
}

func (c *CLI) Submit(ctx context.Context, name string) error {
	s, err := resolveSolution(c.CWD, name)
	if err != nil {
		return fmt.Errorf("submit %q: %w", name, err)
	}

	w, err := openWorkspace(c.CWD)
	if err != nil {
		return fmt.Errorf("submit %q: %w", s.path, err)
	}

	res, err := c.SubmitClient.Submit(ctx, &submitv1.SubmitRequest{
		ProblemId: w.problem.GetId(),
		FileName:  filepath.Base(s.path),
		Content:   s.content,
	})
	if err != nil {
		return fmt.Errorf("submit %q: %w", s.path, err)
	}
	if res.Error != nil {
		return fmt.Errorf("submit %q: extension error: %s", s.path, res.GetError())
	}
	return nil
}
