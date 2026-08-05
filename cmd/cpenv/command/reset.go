package command

// import (
// 	"github.com/spf13/cobra"
// )

// var resetTmpl string

// var resetCmd = &cobra.Command{
// 	Use:   "reset [-t template]",
// 	Short: "Reset the current workspace.",
// 	Long:  "Re-initialize the current workspace.",
// 	Args:  cobra.NoArgs,
// 	RunE: func(_ *cobra.Command, _ []string) error {
// 		if err := a.Reset(resetTmpl); err != nil {
// 			return err
// 		}

// 		return nil
// 	},
// }

// func init() {
// 	rootCmd.AddCommand(resetCmd)
// 	resetCmd.Flags().StringVarP(&resetTmpl, "template", "t", "", "template to use")
// }
