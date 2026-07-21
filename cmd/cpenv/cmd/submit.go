package cmd

import (
	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use: "submit",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
}
