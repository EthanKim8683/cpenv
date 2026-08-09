package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/submit/v1/submitv1connect"
	"github.com/EthanKim8683/cpenv/internal/submission"
	"github.com/adrg/xdg"
)

type Home struct {
	dir             string
	submissionStore submission.Store
	submitClient    submitv1connect.SubmitServiceClient
}

func (h *Home) statePath() string {
	return filepath.Join(xdg.StateHome, "cpenv", "state.json")
}

func (h *Home) workspacesDir() string {
	return filepath.Join(h.dir, "workspaces")
}

func (h *Home) workspaceDir(problemID string) string {
	return filepath.Join(h.workspacesDir(), problemID)
}

func (h *Home) templatesDir() string {
	return filepath.Join(h.dir, "templates")
}

func (h *Home) exists() (bool, error) {
	_, err := os.Stat(h.dir)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (h *Home) init() error {
	if err := os.MkdirAll(h.dir, 0755); err != nil {
		return fmt.Errorf("init home: %w", err)
	}

	if err := os.MkdirAll(h.workspacesDir(), 0755); err != nil {
		return fmt.Errorf("init home: %w", err)
	}
	if err := os.MkdirAll(h.templatesDir(), 0755); err != nil {
		return fmt.Errorf("init home: %w", err)
	}

	tmplPath := filepath.Join(h.templatesDir(), "default.star")
	if err := os.WriteFile(tmplPath, []byte("files = {}"), 0644); err != nil {
		return fmt.Errorf("init home: %w", err)
	}
	ss := h.stateStore()
	if err := ss.save(&state{
		LastTemplatePath: tmplPath,
	}); err != nil {
		return fmt.Errorf("init home: %w", err)
	}
	return nil
}

func (h *Home) stateStore() stateStore {
	return &fileStateStore{path: h.statePath()}
}

func (h *Home) newWorkspace(problem *problemv1.Problem) *Workspace {
	return &Workspace{
		dir:             h.workspaceDir(problem.GetId()),
		problem:         problem,
		submissionStore: h.submissionStore,
		submitClient:    h.submitClient,
		templateGetter:  h,
	}
}

func (h *Home) template(tmpl string) (*template, error) {
	var path string
	if filepath.IsAbs(tmpl) {
		path = tmpl
	} else if tmpl == "" {
		ss := h.stateStore()
		s, err := ss.load()
		if err != nil {
			return nil, err
		}

		path = s.LastTemplatePath
	} else {
		path = filepath.Join(h.templatesDir(), tmpl)
	}
	return &template{path: path}, nil
}
