package env

import (
	"errors"
	"fmt"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/spf13/afero"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const problemFile = "problem.json"

type Env struct {
	fs      afero.Fs
	problem *problemv1.Problem
}

func (e *Env) Problem() *problemv1.Problem {
	return e.problem
}

func (e *Env) Clear() error {
	entries, err := afero.ReadDir(e.fs, ".")
	if err != nil {
		return fmt.Errorf("clear env: %w", err)
	}

	var errs error
	for _, entry := range entries {
		if entry.Name() == problemFile {
			continue
		}

		errs = errors.Join(errs, e.fs.RemoveAll(entry.Name()))
	}
	if errs != nil {
		return fmt.Errorf("clear env: %w", errs)
	}

	return nil
}

func Open(fs afero.Fs) (*Env, error) {
	data, err := afero.ReadFile(fs, problemFile)
	if err != nil {
		return nil, fmt.Errorf("open env: %w", err)
	}

	var problem problemv1.Problem
	if err := protojson.Unmarshal(data, &problem); err != nil {
		return nil, fmt.Errorf("open env: %w", err)
	}

	return &Env{
		fs:      fs,
		problem: &problem,
	}, nil
}

func Create(fs afero.Fs, problem *problemv1.Problem) (*Env, error) {
	data, err := protojson.Marshal(problem)
	if err != nil {
		return nil, fmt.Errorf("create env: %w", err)
	}

	if err := fs.MkdirAll(".", 0755); err != nil {
		return nil, fmt.Errorf("create env: %w", err)
	}

	if err := afero.WriteFile(fs, problemFile, data, 0644); err != nil {
		return nil, fmt.Errorf("create env: %w", err)
	}

	return &Env{
		fs:      fs,
		problem: proto.CloneOf(problem),
	}, nil
}
