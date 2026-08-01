package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/EthanKim8683/cpenv/internal/env"
	"github.com/EthanKim8683/cpenv/internal/template"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var focusCmd = &cobra.Command{
	Use:  "focus",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := store.Load()
		if err != nil {
			return err
		}

		focus := state.Focus
		if errMsg := focus.Error; errMsg != "" {
			return fmt.Errorf("extension: %s", errMsg)
		}

		problem := focus.Problem
		if problem == nil {
			return fmt.Errorf("no focused problem")
		}

		path := filepath.Join(cfg.EnvsDir, problem.Id)
		if _, err := os.Stat(path); err == nil {
			fmt.Println(path)
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		templatesFs := afero.NewBasePathFs(afero.NewOsFs(), cfg.TemplatesDir)

		templateName := state.LastUsedTemplateName
		if templateName == "" {
			if templateName, err = anyTemplate(templatesFs); err != nil {
				return err
			}
		}

		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}

		fs := afero.NewBasePathFs(afero.NewOsFs(), path)

		e, err := env.Create(fs, problem)
		if err != nil {
			return err
		}

		if err := e.Clear(); err != nil {
			return err
		}

		if err := template.Render(
			templatesFs,
			templateName,
			fs,
			problem,
		); err != nil {
			return err
		}

		fmt.Println(path)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(focusCmd)
}
