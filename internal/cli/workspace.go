package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	submissionv1 "github.com/EthanKim8683/cpenv/internal/gen/submission/v1"
	submitv1 "github.com/EthanKim8683/cpenv/internal/gen/submit/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/submit/v1/submitv1connect"
	"github.com/EthanKim8683/cpenv/internal/submission"
)

type Workspace struct {
	dir             string
	problem         *problemv1.Problem
	submissionStore *submission.Store
	submitClient    submitv1connect.SubmitServiceClient
}

func (w *Workspace) exists() (bool, error) {
	_, err := os.Stat(w.dir)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (w *Workspace) init(tmpl string) error {
	// template.render
}

func (w *Workspace) Reset(tmpl string) error {
	// clear and init
}

func (w *Workspace) Status(limit int) ([]*submissionv1.Submission, error) {
	subs, err := w.submissionStore.TailProblem(w.problem.GetId(), limit)
	if err != nil {
		return nil, fmt.Errorf("status: %w", err)
	}
	return subs, nil
}

func (w *Workspace) Submit(ctx context.Context, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("submit: %w", err)
	}

	res, err := w.submitClient.Submit(ctx, &submitv1.SubmitRequest{
		ProblemId: w.problem.GetId(),
		FileName:  filepath.Base(path),
		Content:   content,
	})
	if err != nil {
		return fmt.Errorf("submit: %w", err)
	}

	if errMsg := res.GetError(); errMsg != "" {
		return fmt.Errorf("submit: extension: %s", errMsg)
	}
	return nil
}
