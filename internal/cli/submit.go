package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	submitv1 "github.com/EthanKim8683/cpenv/internal/gen/submit/v1"
)

type solution struct {
	path    string
	content []byte
}

func newSolution(path string) (*solution, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read solution %q: %w", path, err)
	}
	return &solution{path: path, content: content}, nil
}

func (c *CLI) submit(ctx context.Context, problem *problemv1.Problem, sol *solution) error {
	res, err := c.SubmitClient.Submit(ctx, &submitv1.SubmitRequest{
		ProblemId: problem.GetId(),
		FileName:  filepath.Base(sol.path),
		Content:   sol.content,
	})
	if err != nil {
		return err
	}

	if res.Error != nil {
		return fmt.Errorf("extension error: %s", res.GetError())
	}
	return nil
}

func (c *CLI) Submit(ctx context.Context, name string) error {
	sol, err := c.resolveSolution(name)
	if err != nil {
		return fmt.Errorf("submit %q: %w", name, err)
	}

	w, err := openWorkspace(c.CWD)
	if err != nil {
		return fmt.Errorf("submit %q: %w", sol.path, err)
	}

	if err := c.submit(ctx, w.problem, sol); err != nil {
		return fmt.Errorf("submit %q: %w", sol.path, err)
	}
	return nil
}
