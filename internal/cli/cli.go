package cli

import (
	"fmt"

	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/EthanKim8683/cpenv/internal/focus"
	submissionv1 "github.com/EthanKim8683/cpenv/internal/gen/submission/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/submit/v1/submitv1connect"
	"github.com/EthanKim8683/cpenv/internal/submission"
)

type CLI struct {
	Cfg             *config.Config
	FocusStore      focus.Store
	SubmissionStore submission.Store
	SubmitClient    submitv1connect.SubmitServiceClient
}

func (c *CLI) NewHome() *Home {
	return &Home{
		dir:             c.Cfg.HomeDir,
		submissionStore: c.SubmissionStore,
		submitClient:    c.SubmitClient,
	}
}

func (c *CLI) Focus(tmpl string) (string, error) {
	problem, err := c.FocusStore.Problem()
	if err != nil {
		return "", fmt.Errorf("focus: %w", err)
	}

	h := c.NewHome()
	if ok, err := h.exists(); err != nil {
		return "", fmt.Errorf("focus: %w", err)
	} else if !ok {
		if err := h.init(); err != nil {
			return "", fmt.Errorf("focus: %w", err)
		}
	}

	w := h.newWorkspace(problem)
	if ok, err := w.exists(); err != nil {
		return "", fmt.Errorf("focus: %w", err)
	} else if !ok {
		if err := w.init(tmpl); err != nil {
			return "", fmt.Errorf("focus: %w", err)
		}
	}

	return h.workspaceDir(problem.GetId()), nil
}

func (c *CLI) Status(limit int) ([]*submissionv1.Submission, error) {
	subs, err := c.SubmissionStore.Tail(limit)
	if err != nil {
		return nil, fmt.Errorf("status: %w", err)
	}
	return subs, nil
}
