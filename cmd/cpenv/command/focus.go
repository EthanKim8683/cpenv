package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var focusTmpl string

var focusCmd = &cobra.Command{
	Use:  "focus",
	Args: cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		path, err := a.Focus(focusTmpl)
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
