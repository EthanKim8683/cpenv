package cli

import (
	"errors"
	"fmt"
	"os"
)

func removeAll(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var errs error
	for _, entry := range entries {
		errs = errors.Join(errs, os.RemoveAll(entry.Name()))
	}
	return errs
}

func (c *CLI) Reset(templateName string) error {
	w, err := openWorkspace(c.Cwd)
	if err != nil {
		return fmt.Errorf("reset %q: %w", c.Cwd, err)
	}

	if err := removeAll(c.Cwd); err != nil {
		_ = w.close()
		return fmt.Errorf("reset %q: %w", c.Cwd, err)
	}

	tPath, err := c.resolveTemplatePath(templateName)
	if err != nil {
		_ = w.close()
		return fmt.Errorf("reset %q: %w", c.Cwd, err)
	}
	t := &template{path: tPath}

	if err := t.render(w.dir, w.problem); err != nil {
		_ = w.close()
		return fmt.Errorf("reset %q: %w", c.Cwd, err)
	}

	if err := w.close(); err != nil {
		return fmt.Errorf("reset %q: %w", c.Cwd, err)
	}
	return nil
}
