package workspace

import (
	"fmt"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/spf13/afero"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const problemFile = ".problem.json"

type Workspace struct {
	fs      afero.Fs
	problem *problemv1.Problem
}

func (ws *Workspace) Problem() *problemv1.Problem {
	return ws.problem
}

func (ws *Workspace) Clear() error {
	entries, err := afero.ReadDir(ws.fs, ".")
	if err != nil {
		return fmt.Errorf("clear workspace: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == problemFile {
			continue
		}

		if err := ws.fs.RemoveAll(name); err != nil {
			return fmt.Errorf("clear workspace: %w", err)
		}
	}

	return nil
}

func Open(fs afero.Fs) (*Workspace, error) {
	data, err := afero.ReadFile(fs, problemFile)
	if err != nil {
		return nil, fmt.Errorf("open workspace: %w", err)
	}

	var problem problemv1.Problem
	if err := protojson.Unmarshal(data, &problem); err != nil {
		return nil, fmt.Errorf("open workspace: %w", err)
	}

	return &Workspace{fs: fs, problem: &problem}, nil
}

func Create(fs afero.Fs, problem *problemv1.Problem) (*Workspace, error) {
	data, err := protojson.Marshal(problem)
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	if err := fs.MkdirAll(".", 0755); err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	if err := afero.WriteFile(fs, problemFile, data, 0644); err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	return &Workspace{fs: fs, problem: proto.CloneOf(problem)}, nil
}
