package app

import (
	"fmt"
	"os"

	"github.com/EthanKim8683/cpenv/internal/env"
	"github.com/spf13/afero"
)

func (a *App) Reset(tmpl string) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("reset: %w", err)
	}

	fs := afero.NewBasePathFs(afero.NewOsFs(), dir)

	e, err := env.Open(fs)
	if err != nil {
		return fmt.Errorf("reset: %w", err)
	}

	if err := e.Clear(); err != nil {
		return fmt.Errorf("reset: %w", err)
	}

	if err := a.renderTemplate(fs, tmpl, e.Problem()); err != nil {
		return fmt.Errorf("reset: %w", err)
	}

	return nil
}
