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
	"google.golang.org/protobuf/encoding/protojson"
)

const problemFile = ".problem.json"

type Workspace struct {
	dir             string
	problem         *problemv1.Problem
	submissionStore submission.Store
	submitClient    submitv1connect.SubmitServiceClient
	templateGetter  templateGetter
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
	if err := os.MkdirAll(w.dir, 0755); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	data, err := protojson.Marshal(w.problem)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}

	if err := os.WriteFile(filepath.Join(w.dir, problemFile), data, 0644); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	t, err := w.templateGetter.template(tmpl)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}

	if err := t.render(w.dir, w.problem); err != nil {
		return fmt.Errorf("init: %w", err)
	}
	return nil
}

func (w *Workspace) Reset(tmpl string) error {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if name == problemFile {
			continue
		}
		if err := os.RemoveAll(filepath.Join(w.dir, name)); err != nil {
			return fmt.Errorf("reset: %w", err)
		}
	}

	if err := w.init(tmpl); err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	return nil
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

func NewWorkspace(
	dir string,
	submissionStore submission.Store,
	submitClient submitv1connect.SubmitServiceClient,
	templateGetter templateGetter,
) (*Workspace, error) {
	data, err := os.ReadFile(filepath.Join(dir, problemFile))
	if err != nil {
		return nil, err
	}

	problem := &problemv1.Problem{}
	if err := protojson.Unmarshal(data, problem); err != nil {
		return nil, err
	}

	return &Workspace{
		dir:             dir,
		problem:         problem,
		submissionStore: submissionStore,
		submitClient:    submitClient,
		templateGetter:  templateGetter,
	}, nil
}
