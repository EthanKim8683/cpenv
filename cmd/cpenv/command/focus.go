package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/EthanKim8683/cpenv/internal/state"
	"github.com/EthanKim8683/cpenv/internal/template"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

// alias focus="cd \"$(cpx focus)\""

var focusCmd = &cobra.Command{
	Use:  "focus",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
			os.Exit(1)
		}

		store := state.NewFileStore(cfg.StatePath)
		state, err := store.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
			os.Exit(1)
		}

		problem := state.Focus.Problem
		if problem == nil {
			fmt.Fprintf(os.Stderr, "cpenv: no focused problem\n")
			os.Exit(1)
		}

		path := filepath.Join(cfg.EnvsDir, problem.Id)

		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			templatesFs := afero.NewBasePathFs(afero.NewOsFs(), cfg.TemplatesDir)

			templateName := state.LastUsedTemplateName
			if templateName == "" {
				templateNames, err := afero.Glob(templatesFs, "*.star")
				if err != nil {
					fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
					os.Exit(1)
				}

				if len(templateNames) == 0 {
					fmt.Fprintf(os.Stderr, "cpenv: no templates found\n")
					os.Exit(1)
				}

				templateName = templateNames[0]
			}

			if err := os.MkdirAll(path, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
				os.Exit(1)
			}

			fs := afero.NewBasePathFs(afero.NewOsFs(), path)

			data, err := protojson.Marshal(problem)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
				os.Exit(1)
			}

			if err := afero.WriteFile(fs, "problem.json", data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
				os.Exit(1)
			}

			if err := template.Render(templatesFs, templateName, fs, problem); err != nil {
				fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
				os.Exit(1)
			}

			state.LastUsedTemplateName = templateName
			if err := store.Save(state); err != nil {
				fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
				os.Exit(1)
			}
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(path)
	},
}

func init() {
	rootCmd.AddCommand(focusCmd)
}
