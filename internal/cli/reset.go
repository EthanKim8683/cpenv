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

	defaultTemplate, err := c.Preferences.DefaultTemplate()
	if err != nil {
		return fmt.Errorf("reset %q: %w", c.CWD, err)
	}

	t, err := resolveTemplate(templateName, c.CWD, c.templatesDir(), defaultTemplate)
	if err != nil {
		return fmt.Errorf("reset %q: %w", c.CWD, err)
	}
	if t == nil {
		return nil
	}
	if err := t.render(w.dir, w.problem); err != nil {
		return fmt.Errorf("reset %q: %w", c.CWD, err)
	}

	if err := c.Preferences.SetDefaultTemplate(t.path); err != nil {
		return fmt.Errorf("reset %q: %w", c.CWD, err)
	}
	return nil
}
