package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	starlarkjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"google.golang.org/protobuf/encoding/protojson"
)

type template struct {
	path string
	src  []byte
}

func encodeProblem(thread *starlark.Thread, problem *problemv1.Problem) (starlark.Value, error) {
	data, err := protojson.Marshal(problem)
	if err != nil {
		return nil, fmt.Errorf("encode problem: %w", err)
	}

	v, err := starlark.Call(
		thread,
		starlarkjson.Module.Members["decode"],
		starlark.Tuple{starlark.String(data)},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("encode problem: %w", err)
	}
	return v, nil
}

func execTemplate(thread *starlark.Thread, name string, src []byte, problem starlark.Value) (starlark.Value, error) {
	globals, err := starlark.ExecFileOptions(
		&syntax.FileOptions{
			While:           true,
			TopLevelControl: true,
			GlobalReassign:  true,
		},
		thread,
		name,
		src,
		starlark.StringDict{
			"problem": problem,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("exec template: %w", err)
	}

	v, ok := globals["files"]
	if !ok {
		return nil, errors.New("exec template: files is unbound")
	}
	return v, nil
}

func decodeFiles(value starlark.Value) (map[string]string, error) {
	dict, ok := value.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("decode files: expected dict, got %s", value.Type())
	}

	files := make(map[string]string, dict.Len())
	var errs error
	for k, v := range dict.Entries() {
		skip := false

		fileName, ok := starlark.AsString(k)
		if !ok {
			errs = errors.Join(errs, fmt.Errorf("%s: expected string file name, got %s", k, k.Type()))
			skip = true
		}

		content, ok := starlark.AsString(v)
		if !ok {
			errs = errors.Join(errs, fmt.Errorf("%s: expected string content, got %s", k, v.Type()))
			skip = true
		}

		if skip {
			continue
		}

		files[fileName] = content
	}
	if errs != nil {
		return nil, fmt.Errorf("decode files: %w", errs)
	}
	return files, nil
}

func writeFiles(dir string, files map[string]string) error {
	var errs error
	for fileName, content := range files {
		path := filepath.Join(dir, fileName)

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		errs = errors.Join(errs, os.WriteFile(path, []byte(content), 0644))
	}
	if errs != nil {
		return fmt.Errorf("write files to %q: %w", dir, errs)
	}
	return nil
}

func (t *template) render(dir string, problem *problemv1.Problem) error {
	thread := &starlark.Thread{}
	pValue, err := encodeProblem(thread, problem)
	if err != nil {
		return fmt.Errorf("render %q: %w", t.path, err)
	}

	fValue, err := execTemplate(thread, t.path, t.src, pValue)
	if err != nil {
		return fmt.Errorf("render %q: %w", t.path, err)
	}

	files, err := decodeFiles(fValue)
	if err != nil {
		return fmt.Errorf("render %q: %w", t.path, err)
	}

	if err := writeFiles(dir, files); err != nil {
		return fmt.Errorf("render %q: %w", t.path, err)
	}
	return nil
}

func newTemplate(path string) (*template, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &template{path: path, src: src}, nil
}
