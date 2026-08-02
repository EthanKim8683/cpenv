package command

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/EthanKim8683/cpenv/internal/app"
	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/EthanKim8683/cpenv/internal/state"
	"github.com/spf13/cobra"
)

var a *app.App

var rootCmd = &cobra.Command{
	Use:   "cpenv",
	Short: "Competitive programming environment utility.",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		stateStore := state.NewFileStore(filepath.Join(cfg.StatePath))

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		a = &app.App{
			Cfg:        cfg,
			HTTPClient: http.DefaultClient,
			StateStore: stateStore,
			WorkingDir: dir,
		}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
