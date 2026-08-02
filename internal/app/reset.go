package app

import (
	"fmt"

	"github.com/EthanKim8683/cpenv/internal/workspace"
	"github.com/spf13/afero"
)

func (a *App) Reset(tmpl string) error {
	fs := afero.NewBasePathFs(afero.NewOsFs(), a.WorkingDir)

	ws, err := workspace.Open(fs)
	if err != nil {
		return fmt.Errorf("reset: %w", err)
	}

	if err := ws.Clear(); err != nil {
		return fmt.Errorf("reset: %w", err)
	}

	tmpl, err = a.resolveTemplate(tmpl)
	if err != nil {
		return fmt.Errorf("reset: %w", err)
	}

	if err := a.renderTemplate(fs, tmpl, ws.Problem()); err != nil {
		return fmt.Errorf("reset: %w", err)
	}

	return nil
}
