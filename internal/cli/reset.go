package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func clearDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var errs error
	for _, entry := range entries {
		errs = errors.Join(errs, os.RemoveAll(filepath.Join(dir, entry.Name())))
	}
	return errs
}

func (c *CLI) Reset(templateName string) error {
	w, err := openWorkspace(c.CWD)
	if err != nil {
		return fmt.Errorf("reset: %w", err)
	}

	if err := clearDir(c.CWD); err != nil {
		_ = w.close()
		return fmt.Errorf("reset: %w", err)
	}

	if err := c.renderTemplate(templateName, w.dir, w.problem); err != nil {
		_ = w.close()
		return fmt.Errorf("reset: %w", err)
	}

	if err := w.close(); err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	return nil
}
