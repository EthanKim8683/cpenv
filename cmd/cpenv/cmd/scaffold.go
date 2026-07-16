package cmd

import "github.com/spf13/cobra"

var scaffoldCmd = &cobra.Command{
	Use: "scaffold",
	Run: func(cmd *cobra.Command, args []string) {
		// rescaffold the current environment. which scaffold to use is passed as an
		// argument.
	},
}

func init() {
	rootCmd.AddCommand(scaffoldCmd)
}
