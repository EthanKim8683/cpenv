package cli

import (
	"errors"
	"fmt"
	"os"

	submissionv1 "github.com/EthanKim8683/cpenv/internal/gen/status/v1"
)

func (c *CLI) Status(limit int) ([]*submissionv1.Submission, error) {
	var subs []*submissionv1.Submission
	if w, err := openWorkspace(c.CWD); err == nil {
		defer w.close()
		subs, err = c.Submissions.TailProblem(w.problem.GetId(), limit)
		if err != nil {
			return nil, fmt.Errorf("status: %w", err)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		subs, err = c.Submissions.Tail(limit)
		if err != nil {
			return nil, fmt.Errorf("status: %w", err)
		}
	} else {
		return nil, fmt.Errorf("status: %w", err)
	}
	return subs, nil
}
