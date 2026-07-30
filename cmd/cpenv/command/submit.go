package command

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	submitv1 "github.com/EthanKim8683/cpenv/gen/submit/v1"
	"github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/spf13/cobra"
)

// alias submit="cpx submit"

var submitCmd = &cobra.Command{
	Use:  "submit",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
			os.Exit(1)
		}

		client := submitv1connect.NewSubmitServiceClient(http.DefaultClient, "http://localhost:"+cfg.Port)

		var path string
		if len(args) == 1 {
			if path, err = filepath.Abs(args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
				os.Exit(1)
			}
		} else {
			files, err := filepath.Glob("sol.*")
			if err != nil {
				fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
				os.Exit(1)
			}

			if len(files) > 1 {
				fmt.Fprintf(os.Stderr, "cpenv: multiple solution files found\n")
				os.Exit(1)
			}

			if len(files) == 0 {
				fmt.Fprintf(os.Stderr, "cpenv: missing solution file\n")
				os.Exit(1)
			}

			path = files[0]
		}

		if _, err := client.Submit(context.Background(), &submitv1.SubmitRequest{
			Path: path,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
}
