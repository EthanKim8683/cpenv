package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	submitv1 "github.com/EthanKim8683/cpenv/internal/gen/submit/v1"
)

// TODO: Use resolver to resolve path
func (c *CLI) Submit(ctx context.Context, name string) error {
	path, err := c.resolveSolutionPath(name)
	if err != nil {
		return fmt.Errorf("submit %q: %w", name, err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("submit %q: %w", path, err)
	}

	w, err := openWorkspace(c.Cwd)
	if err != nil {
		return fmt.Errorf("submit %q: %w", path, err)
	}
	defer w.close()

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
