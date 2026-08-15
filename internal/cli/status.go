package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	statusv1 "github.com/EthanKim8683/cpenv/internal/gen/status/v1"
)

func (c *CLI) Status(ctx context.Context, limit int) ([]*statusv1.Submission, error) {
	var problemID *string
	w, err := openWorkspace(c.CWD)
	if err == nil {
		problemID = new(w.problem.GetId())
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("status: %w", err)
	}

	res, err := c.StatusClient.Tail(ctx, &statusv1.TailRequest{
		Limit:     new(uint32(limit)),
		ProblemId: problemID,
	})
	if err != nil {
		return nil, fmt.Errorf("status: %w", err)
	}
	return res.GetSubmissions(), nil
}
