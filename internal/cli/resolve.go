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

type resolveFunc[T any] func() (T, error)

func resolveFirst[T any](resolveFuncs []resolveFunc[T]) (T, error) {
	var errs error
	for _, resolve := range resolveFuncs {
		v, err := resolve()
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

func resolveTemplateAbsPath(name string) resolveFunc[*template] {
	return func() (*template, error) {
		if !filepath.IsAbs(name) {
			return nil, skipResolver
		}
		return newTemplate(name)
	}
}

func resolveTemplateRelPath(name string, dir string) resolveFunc[*template] {
	return func() (*template, error) {
		if filepath.IsAbs(name) || name == "" {
			return nil, skipResolver
		}
		return newTemplate(filepath.Join(dir, name))
	}
}

func resolveTemplateDefault(name string, p Preferences) resolveFunc[*template] {
	return func() (*template, error) {
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

func resolveTemplateGlob(name string, dir string) resolveFunc[*template] {
	return func() (*template, error) {
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

func (c *CLI) resolveTemplate(templateName string) (*template, error) {
	return resolveFirst([]resolveFunc[*template]{
		resolveTemplateAbsPath(templateName),
		resolveTemplateRelPath(templateName, c.CWD),
		resolveTemplateRelPath(templateName, c.templatesDir()),
		resolveTemplateDefault(templateName, c.Preferences),
		resolveTemplateGlob(templateName, c.templatesDir()),
	})
}

func resolveSolutionAbsPath(name string) resolveFunc[*solution] {
	return func() (*solution, error) {
		if !filepath.IsAbs(name) {
			return nil, skipResolver
		}
		return newSolution(name)
	}
}

func resolveSolutionRelPath(name string, dir string) resolveFunc[*solution] {
	return func() (*solution, error) {
		if filepath.IsAbs(name) || name == "" {
			return nil, skipResolver
		}
		return newSolution(filepath.Join(dir, name))
	}
}

func resolveSolutionGlob(name string, dir string) resolveFunc[*solution] {
	return func() (*solution, error) {
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

func (c *CLI) resolveSolution(solutionName string) (*solution, error) {
	return resolveFirst([]resolveFunc[*solution]{
		resolveSolutionAbsPath(solutionName),
		resolveSolutionRelPath(solutionName, c.CWD),
		resolveSolutionGlob(solutionName, c.CWD),
	})
}
