package cmd

import "github.com/spf13/cobra"

var submitCmd = &cobra.Command{
	Use: "submit",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. submit current environment
		//
		// underlying stuff:
		// - submitting
		// - seeing submission status
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
}
