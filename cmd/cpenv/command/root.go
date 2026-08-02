package command

import (
	"net/http"

	"github.com/EthanKim8683/cpenv/internal/app"
	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/spf13/cobra"
)

var a *app.App

var rootCmd = &cobra.Command{
	Use: "cpenv",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		a = app.NewApp(cfg, http.DefaultClient)

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
