package command

import (
	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:   "submit [file]",
	Short: "Submit a solution to the current problem.",
	Long:  "Submit a solution (default: sol.*) to the problem corresponding to the current workspace via browser extension.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var path string
		if len(args) > 0 {
			path = args[0]
		}

		if err := c.Submit(cmd.Context(), path); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
}
