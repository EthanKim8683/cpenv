package command

import (
	"context"

	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:  "submit",
	Args: cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var solFile string
		if len(args) > 0 {
			solFile = args[0]
		}

		if err := a.Submit(context.Background(), solFile); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
}
