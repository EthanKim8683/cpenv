package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
	"go.starlark.net/syntax"
	"google.golang.org/protobuf/encoding/protojson"
)

type template struct {
	path       string
	src        []byte
	stateStore stateStore
}

func encodeProblem(thread *starlark.Thread, problem *problemv1.Problem) (starlark.Value, error) {
	data, err := protojson.Marshal(problem)
	if err != nil {
		return nil, err
	}

	value, err := starlark.Call(
		thread,
		starlarkjson.Module.Members["decode"],
		starlark.Tuple{starlark.String(data)},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func execTemplate(thread *starlark.Thread, tmpl string, src []byte, problem starlark.Value) (starlark.Value, error) {
	globals, err := starlark.ExecFileOptions(
		&syntax.FileOptions{
			While:           true,
			TopLevelControl: true,
		},
		thread,
		tmpl,
		src,
		starlark.StringDict{
			"problem": problem,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("starlark: %w", err)
	}

	value, ok := globals["files"]
	if !ok {
		return nil, errors.New("missing global \"files\"")
	}
	return value, nil
}

func decodeFiles(value starlark.Value) (map[string]string, error) {
	dict, ok := value.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("expected dict, got %s", value.Type())
	}

	files := make(map[string]string)
	var errs error
	for key, value := range dict.Entries() {
		skip := false

		path, ok := starlark.AsString(key)
		if !ok {
			errs = errors.Join(errs, fmt.Errorf("file %s: expected string path, got %s", key, key.Type()))
			skip = true
		}

		content, ok := starlark.AsString(value)
		if !ok {
			errs = errors.Join(errs, fmt.Errorf("file %s: expected string content, got %s", key, value.Type()))
			skip = true
		}

		if skip {
			continue
		}
		files[path] = content
	}
	if errs != nil {
		return nil, errs
	}
	return files, nil
}

func writeFiles(dir string, files map[string]string) error {
	for path, content := range files {
		path = filepath.FromSlash(path)
		path = filepath.Join(dir, path)
		path = filepath.Clean(path)

		if !filepath.IsLocal(path) || path == "." {
			return fmt.Errorf("path %q is not local", path)
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(path), err)
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("write %q: %w", path, err)
		}
	}
	return nil
}

func (t *template) render(dir string, problem *problemv1.Problem) error {
	thread := &starlark.Thread{}

	problemValue, err := encodeProblem(thread, problem)
	if err != nil {
		return fmt.Errorf("render template %q: encode problem: %w", t.path, err)
	}

	filesValue, err := execTemplate(thread, t.path, t.src, problemValue)
	if err != nil {
		return fmt.Errorf("render template %q: exec: %w", t.path, err)
	}

	files, err := decodeFiles(filesValue)
	if err != nil {
		return fmt.Errorf("render template %q: decode files: %w", t.path, err)
	}

	if err := writeFiles(dir, files); err != nil {
		return fmt.Errorf("render template %q: write files: %w", t.path, err)
	}
	return nil
}

func newTemplate(path string, stateStore stateStore) (*template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("new template %q: %w", path, err)
	}

	return &template{
		path:       path,
		src:        data,
		stateStore: stateStore,
	}, nil
}

type templateGetter interface {
	template(name string) (*template, error)
}
