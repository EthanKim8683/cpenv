package cli

import (
	"fmt"
)

func (c *CLI) Reset(templateName string) error {
	w, err := openWorkspace(c.CWD)
	if err != nil {
		return fmt.Errorf("reset %q: %w", c.CWD, err)
	}

	if err := w.clear(); err != nil {
		return fmt.Errorf("reset %q: %w", c.CWD, err)
	}

	t, err := c.resolveTemplate(templateName)
	if err != nil {
		return fmt.Errorf("reset %q: %w", c.CWD, err)
	}

	if err := t.render(c.CWD, w.problem); err != nil {
		return fmt.Errorf("reset %q: %w", c.CWD, err)
	}

	if err := c.Preferences.SetDefaultTemplate(t.path); err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	return nil
}
