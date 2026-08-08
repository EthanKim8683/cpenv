package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:   "submit [file]",
	Short: "Submit a solution to the current problem.",
	Long:  "Submit a solution (default: sol.*) to the problem corresponding to the current workspace via browser extension.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		var path string
		if len(args) > 0 {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("submit: %w", err)
			}
			path = filepath.Join(dir, args[0])
		}
		return w.Submit(context.Background(), path)
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
}
