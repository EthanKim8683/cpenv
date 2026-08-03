package app

import (
	"fmt"

	"github.com/EthanKim8683/cpenv/internal/state"
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

	if tmpl == "" {
		store := state.NewFileStore(a.Cfg.StatePath())

		st, err := store.Load()
		if err != nil {
			return fmt.Errorf("reset: %w", err)
		}

		tmpl = st.Template
		if tmpl == "" {
			return fmt.Errorf("reset: no template")
		}
	}

	return nil
}
