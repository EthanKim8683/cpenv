package workspace

import (
	"errors"
	"fmt"
	"path/filepath"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/spf13/afero"
	starlarkjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"google.golang.org/protobuf/encoding/protojson"
)

func encodeProblem(thread *starlark.Thread, problem *problemv1.Problem) (starlark.Value, error) {
	data, err := protojson.Marshal(problem)
	if err != nil {
		return nil, fmt.Errorf("encode problem: marshal: %w", err)
	}

	value, err := starlark.Call(
		thread,
		starlarkjson.Module.Members["decode"],
		starlark.Tuple{starlark.String(data)},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("encode problem: starlark decode: %w", err)
	}

	return value, nil
}

func renderFiles(
	thread *starlark.Thread,
	scaffoldsFs afero.Fs,
	scaffoldFile string,
	problemValue starlark.Value,
) (starlark.Value, error) {
	scaffold, err := afero.ReadFile(scaffoldsFs, scaffoldFile)
	if err != nil {
		return nil, fmt.Errorf("render files: read file: %w", err)
	}

	globals, err := starlark.ExecFileOptions(
		&syntax.FileOptions{
			While:           true,
			TopLevelControl: true,
		},
		thread,
		scaffoldFile,
		scaffold,
		starlark.StringDict{
			"problem": problemValue,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("render files: starlark exec: %w", err)
	}

	value, ok := globals["files"]
	if !ok {
		return nil, fmt.Errorf("render files: missing global files")
	}
	return value, nil
}

func decodeFiles(filesValue starlark.Value) (map[string]string, error) {
	dict, ok := filesValue.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("decode files: expected dict, got %s", filesValue.Type())
	}

	files := make(map[string]string)
	var errs error
	for key, value := range dict.Entries() {
		skip := false

		path, ok := starlark.AsString(key)
		if !ok {
			errs = errors.Join(errs, fmt.Errorf("decode file %s: expected string path, got %s", key, key.Type()))
			skip = true
		}

		content, ok := starlark.AsString(value)
		if !ok {
			errs = errors.Join(errs, fmt.Errorf("decode file %s: expected string content, got %s", key, value.Type()))
			skip = true
		}

		if skip {
			continue
		}
		files[path] = content
	}
	if errs != nil {
		return nil, fmt.Errorf("decode files: %w", errs)
	}
	return files, nil
}

func materializeFiles(workspaceFs afero.Fs, files map[string]string) error {
	for path, content := range files {
		path = filepath.FromSlash(path)

		if err := workspaceFs.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("materialize files: mkdir %q: %w", filepath.Dir(path), err)
		}

		if err := afero.WriteFile(workspaceFs, path, []byte(content), 0644); err != nil {
			return fmt.Errorf("materialize files: write %q: %w", path, err)
		}
	}
	return nil
}

func scaffold(
	workspaceFs afero.Fs,
	scaffoldsFs afero.Fs,
	scaffoldFile string,
	problem *problemv1.Problem,
) error {
	thread := &starlark.Thread{}

	problemValue, err := encodeProblem(thread, problem)
	if err != nil {
		return fmt.Errorf("scaffold %q: %w", scaffoldFile, err)
	}

	filesValue, err := renderFiles(
		thread,
		scaffoldsFs,
		scaffoldFile,
		problemValue,
	)
	if err != nil {
		return fmt.Errorf("scaffold %q: %w", scaffoldFile, err)
	}

	files, err := decodeFiles(filesValue)
	if err != nil {
		return fmt.Errorf("scaffold %q: %w", scaffoldFile, err)
	}

	if err = materializeFiles(workspaceFs, files); err != nil {
		return fmt.Errorf("scaffold %q: %w", scaffoldFile, err)
	}

	return nil
}
