package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use: "cpenv",
}

func Execute() error {
	return rootCmd.Execute()
}
