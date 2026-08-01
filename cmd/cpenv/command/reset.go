package command

import (
	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/EthanKim8683/cpenv/internal/state"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func reset(
	cfg *config.Config,
	state *state.State,
	fs afero.Fs,
	templateName string,
) error {

}

var resetCmd = &cobra.Command{
	Use: "reset",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
	resetCmd.Flags().StringP("template", "t", "", "")
}
