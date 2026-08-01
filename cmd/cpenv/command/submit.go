package command

import (
	"context"

	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:  "submit",
	Args: cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var subFile string
		if len(args) > 0 {
			subFile = args[0]
		}

		if err := a.Submit(context.Background(), subFile); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
}
