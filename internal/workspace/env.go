package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/spf13/afero"
	starlarkjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"google.golang.org/protobuf/encoding/protojson"
)

const problemPath = "problem.json"

type Env struct {
	fs      afero.Fs
	problem *problemv1.Problem
}

func clearFs(fs afero.Fs) error {
	entries, err := afero.ReadDir(fs, ".")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := fs.RemoveAll(entry.Name()); err != nil {
			return err
		}
	}

	return nil
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

func execTemplate(
	thread *starlark.Thread,
	templatesFs afero.Fs,
	templateName string,
	problemValue starlark.Value,
) (starlark.Value, error) {
	template, err := afero.ReadFile(templatesFs, templateName)
	if err != nil {
		return nil, err
	}

	globals, err := starlark.ExecFileOptions(
		&syntax.FileOptions{
			While:           true,
			TopLevelControl: true,
		},
		thread,
		templateName,
		template,
		starlark.StringDict{
			"problem": problemValue,
		},
	)
	if err != nil {
		return nil, err
	}

	value, ok := globals["files"]
	if !ok {
		return nil, errors.New("missing global \"files\"")
	}
	return value, nil
}

func decodeFiles(filesValue starlark.Value) (map[string]string, error) {
	dict, ok := filesValue.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("expected dict, got %s", filesValue.Type())
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

func writeFiles(fs afero.Fs, files map[string]string) error {
	for path, content := range files {
		path = filepath.FromSlash(path)
		path = filepath.Clean(path)

		if !filepath.IsLocal(path) || path == "." {
			return fmt.Errorf("invalid path %q", path)
		}

		if err := fs.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(path), err)
		}

		if err := afero.WriteFile(fs, path, []byte(content), 0644); err != nil {
			return fmt.Errorf("write %q: %w", path, err)
		}
	}
	return nil
}

func (e *Env) Init(templatesFs afero.Fs, templateName string) error {
	thread := &starlark.Thread{}

	if err := clearFs(e.fs); err != nil {
		return err
	}

	problemValue, err := encodeProblem(thread, e.problem)
	if err != nil {
		return fmt.Errorf("problem: %w", err)
	}

	filesValue, err := execTemplate(
		thread,
		templatesFs,
		templateName,
		problemValue,
	)
	if err != nil {
		return fmt.Errorf("template %q: %w", templateName, err)
	}

	files, err := decodeFiles(filesValue)
	if err != nil {
		return fmt.Errorf("template %q: %w", templateName, err)
	}

	if err = writeFiles(e.fs, files); err != nil {
		return fmt.Errorf("template %q: %w", templateName, err)
	}

	return nil
}

func (e *Env) Exists() (bool, error) {
	if _, err := e.fs.Stat(problemPath); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func (e *Env) Submit(path string) error {
	return nil
}

func NewEnv(fs afero.Fs, problem *problemv1.Problem) *Env {
	return &Env{fs: fs, problem: problem}
}

func LoadEnv(fs afero.Fs) (*Env, error) {
	data, err := afero.ReadFile(fs, problemPath)
	if err != nil {
		return nil, err
	}

	var problem problemv1.Problem
	if err := json.Unmarshal(data, &problem); err != nil {
		return nil, err
	}

	env := &Env{
		fs:      fs,
		problem: &problem,
	}
	return env, nil
}
