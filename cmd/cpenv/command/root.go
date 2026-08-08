package command

import (
	"github.com/EthanKim8683/cpenv/internal/cli"
	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/spf13/cobra"
)

var c *cli.CLI
var w *cli.Workspace

var rootCmd = &cobra.Command{
	Use:   "cpenv",
	Short: "Competitive programming environment CLI.",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		c = &cli.CLI{
			Cfg:             cfg,
			FocusStore:      nil,
			SubmissionStore: nil,
			SubmitClient:    nil,
		}

		w = &cli.Workspace{}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
