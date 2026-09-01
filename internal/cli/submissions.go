package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	submissionsv1 "github.com/EthanKim8683/cpenv/internal/gen/submissions/v1"
)

func (c *CLI) Submissions(ctx context.Context, limit int) ([]*submissionsv1.Submission, error) {
	var problemID *string
	w, err := openWorkspace(c.CWD)
	if err == nil {
		problemID = new(w.problem.GetId())
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("submissions: %w", err)
	}

	res, err := c.SubmissionsClient.Tail(ctx, &submissionsv1.TailRequest{
		Limit:     new(uint32(limit)),
		ProblemId: problemID,
	})
	if err != nil {
		return nil, fmt.Errorf("submissions: %w", err)
	}
	return res.GetSubmissions(), nil
}
