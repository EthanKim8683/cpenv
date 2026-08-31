package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var focusTmpl string

var focusCmd = &cobra.Command{
	Use:     "focus [-t template]",
	Short:   "Focus on the workspace for the current problem.",
	Long:    "Output the path to the workspace for the last opened problem, creating and initializing it if necessary.",
	Example: "cd \"$(cpenv focus -t template.star)\"",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		path, err := c.Focus(cmd.Context(), focusTmpl)
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(focusCmd)
	focusCmd.Flags().StringVarP(&focusTmpl, "template", "t", "", "template to use")
}
