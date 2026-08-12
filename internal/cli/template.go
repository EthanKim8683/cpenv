package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	"github.com/bmatcuk/doublestar/v4"
	bolt "go.etcd.io/bbolt"
	starlarkjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	templateBucketKey  = []byte("template")
	defaultTemplateKey = []byte("default-template")
)

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

func renderTemplate(path string, dir string, problem *problemv1.Problem) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("render %q: %w", path, err)
	}

	thread := &starlark.Thread{}
	pValue, err := encodeProblem(thread, problem)
	if err != nil {
		return fmt.Errorf("render %q: %w", path, err)
	}

	fValue, err := execTemplate(thread, path, src, pValue)
	if err != nil {
		return fmt.Errorf("render %q: %w", path, err)
	}

	files, err := decodeFiles(fValue)
	if err != nil {
		return fmt.Errorf("render %q: %w", path, err)
	}

	if err := writeFiles(dir, files); err != nil {
		return fmt.Errorf("render %q: %w", path, err)
	}
	return nil
}

func (c *CLI) templatesDir() string {
	return filepath.Join(c.Cfg.HomeDir, "templates")
}

func (c *CLI) getDefaultTemplate() (string, error) {
	var path string
	if err := c.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(templateBucketKey)
		if b == nil {
			return nil
		}

		v := b.Get(defaultTemplateKey)
		if v == nil {
			return nil
		}

		path = string(v)
		return nil
	}); err != nil {
		return "", fmt.Errorf("get default template: %w", err)
	}
	return path, nil
}

func (c *CLI) resolveTemplate(name string) (string, error) {
	if path := name; filepath.IsAbs(path) {
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("resolve template %q: %w", name, err)
		}
		return path, nil
	}

	if name != "" {
		var errs error

		path := filepath.Join(c.CWD, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		} else {
			errs = errors.Join(errs, err)
		}

		path = filepath.Join(c.templatesDir(), name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		} else {
			errs = errors.Join(errs, err)
		}

		return "", fmt.Errorf("resolve template %q: %w", name, errs)
	}

	path, err := c.getDefaultTemplate()
	if err != nil {
		return "", fmt.Errorf("resolve template: %w", err)
	}
	if path != "" {
		return path, nil
	}

	matches, err := doublestar.FilepathGlob(filepath.Join(c.templatesDir(), "**", "*.star"))
	if err != nil {
		return "", fmt.Errorf("resolve template: %w", err)
	}
	for _, path := range matches {
		return path, nil
	}

	return "", nil
}

func (c *CLI) setDefaultTemplate(path string) error {
	if err := c.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(templateBucketKey)
		if err != nil {
			return err
		}
		return b.Put(defaultTemplateKey, []byte(path))
	}); err != nil {
		return fmt.Errorf("set default template to %q: %w", path, err)
	}
	return nil
}

func (c *CLI) renderTemplate(name string, dir string, problem *problemv1.Problem) error {
	path, err := c.resolveTemplate(name)
	if err != nil {
		return err
	}
	if path == "" {
		return nil
	}

	if err := renderTemplate(path, dir, problem); err != nil {
		return err
	}

	return c.setDefaultTemplate(path)
}
