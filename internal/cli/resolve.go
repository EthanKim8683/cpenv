package cli

import (
	"errors"
	"fmt"
	"path/filepath"
)

var (
	skipResolver          = errors.New("skip resolver")
	ErrExhaustedResolvers = errors.New("exhausted resolvers")
)

func (c *CLI) workspacesDir() string {
	return filepath.Join(c.Cfg.HomeDir, "workspaces")
}

func (c *CLI) workspaceDir(problemID string) string {
	return filepath.Join(c.workspacesDir(), problemID)
}

func (c *CLI) templatesDir() string {
	return filepath.Join(c.Cfg.HomeDir, "templates")
}

type resolver[T any] func(name string) (T, error)

func resolveFirst[T any](name string, resolvers []resolver[T]) (T, error) {
	var errs error
	for _, resolve := range resolvers {
		v, err := resolve(name)
		if err == nil {
			return v, nil
		}
		if !errors.Is(err, skipResolver) {
			errs = errors.Join(errs, err)
		}
	}
	err := ErrExhaustedResolvers
	if errs != nil {
		err = fmt.Errorf("%w: %w", err, errs)
	}
	var zero T
	return zero, err
}

func resolveTemplateAbsPath() resolver[*template] {
	return func(name string) (*template, error) {
		if !filepath.IsAbs(name) {
			return nil, skipResolver
		}
		return newTemplate(name)
	}
}

func resolveTemplateRelPath(dir string) resolver[*template] {
	return func(name string) (*template, error) {
		if filepath.IsAbs(name) || name == "" {
			return nil, skipResolver
		}
		return newTemplate(filepath.Join(dir, name))
	}
}

func resolveTemplateDefault(p Preferences) resolver[*template] {
	return func(name string) (*template, error) {
		if name != "" {
			return nil, skipResolver
		}
		path, err := p.DefaultTemplate()
		if err != nil {
			return nil, err
		}
		return newTemplate(path)
	}
}

func resolveTemplateGlob(dir string) resolver[*template] {
	return func(name string) (*template, error) {
		if name != "" {
			return nil, skipResolver
		}
		matches, err := filepath.Glob(filepath.Join(dir, "*.star"))
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			return nil, errors.New("no *.star files")
		}
		return newTemplate(matches[0])
	}
}

func (c *CLI) resolveTemplate(name string) (*template, error) {
	return resolveFirst(name, []resolver[*template]{
		resolveTemplateAbsPath(),
		resolveTemplateRelPath(c.CWD),
		resolveTemplateRelPath(c.templatesDir()),
		resolveTemplateDefault(c.Preferences),
		resolveTemplateGlob(c.templatesDir()),
	})
}

func resolveSolutionAbsPath() resolver[*solution] {
	return func(name string) (*solution, error) {
		if !filepath.IsAbs(name) {
			return nil, skipResolver
		}
		return newSolution(name)
	}
}

func resolveSolutionRelPath(dir string) resolver[*solution] {
	return func(name string) (*solution, error) {
		if filepath.IsAbs(name) || name == "" {
			return nil, skipResolver
		}
		return newSolution(filepath.Join(dir, name))
	}
}

func resolveSolutionGlob(dir string) resolver[*solution] {
	return func(name string) (*solution, error) {
		if name != "" {
			return nil, skipResolver
		}
		matches, err := filepath.Glob(filepath.Join(dir, "sol.*"))
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			return nil, errors.New("no sol.* files")
		}
		if len(matches) > 1 {
			return nil, errors.New("multiple sol.* files")
		}
		return newSolution(matches[0])
	}
}

func (c *CLI) resolveSolution(name string) (*solution, error) {
	return resolveFirst(name, []resolver[*solution]{
		resolveSolutionAbsPath(),
		resolveSolutionRelPath(c.CWD),
		resolveSolutionGlob(c.CWD),
	})
}
