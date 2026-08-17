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
		return nil, err
	}

	v, ok := globals["files"]
	if !ok {
		return nil, errors.New("files is unbound")
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
		fileName = filepath.Clean(fileName)
		if !filepath.IsLocal(fileName) || fileName == "." {
			errs = errors.Join(errs, fmt.Errorf("filename is not local: %q", fileName))
			continue
		}

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
	problemValue, err := encodeProblem(thread, problem)
	if err != nil {
		return fmt.Errorf("render template %q: %w", t.path, err)
	}

	filesValue, err := execTemplate(thread, filepath.Base(t.path), t.src, problemValue)
	if err != nil {
		return fmt.Errorf("render template %q: %w", t.path, err)
	}

	files, err := decodeFiles(filesValue)
	if err != nil {
		return fmt.Errorf("render template %q: %w", t.path, err)
	}

	if err := writeFiles(dir, files); err != nil {
		return fmt.Errorf("render template %q: %w", t.path, err)
	}
	return nil
}

func (c *CLI) templatesDir() string {
	return filepath.Join(c.Cfg.HomeDir, "templates")
}

func resolveTemplate(name string, cwd string, templatesDir string, defaultTemplate string) (*template, error) {
	if filepath.IsAbs(name) {
		src, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}
		return &template{path: name, src: src}, nil
	}

	if name != "" {
		var errs error

		path := filepath.Join(cwd, name)
		src, err := os.ReadFile(path)
		if err == nil {
			return &template{path: path, src: src}, nil
		}
		errs = errors.Join(errs, err)

		path = filepath.Join(templatesDir, name)
		src, err = os.ReadFile(path)
		if err == nil {
			return &template{path: path, src: src}, nil
		}
		errs = errors.Join(errs, err)

		return nil, errs
	}

	if defaultTemplate != "" {
		src, err := os.ReadFile(defaultTemplate)
		if err == nil {
			return &template{path: defaultTemplate, src: src}, nil
		}
	}

	matches, err := filepath.Glob(filepath.Join(templatesDir, "*.star"))
	if err != nil {
		return nil, err
	}
	if len(matches) >= 1 {
		path := matches[0]
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return &template{path: path, src: src}, nil
	}

	return nil, nil
}
