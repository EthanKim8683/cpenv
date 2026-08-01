package command

import (
	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/EthanKim8683/cpenv/internal/state"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type app struct {
	cfg   *config.Config
	store state.Store
}

func (a *app) templatesFs() afero.Fs {
	return afero.NewBasePathFs(afero.NewOsFs(), a.cfg.TemplatesDir)
}

var a *app

var rootCmd = &cobra.Command{
	Use: "cpenv",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		a = &app{
			cfg:   cfg,
			store: state.NewFileStore(cfg.StatePath),
		}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
