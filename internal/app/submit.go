package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	submitv1 "github.com/EthanKim8683/cpenv/gen/submit/v1"
	"github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
	"github.com/EthanKim8683/cpenv/internal/workspace"
	"github.com/spf13/afero"
)

func (a *App) defaultSolFile() (string, error) {
	matches, err := filepath.Glob(filepath.Join(a.WorkingDir, "sol.*"))
	if err != nil {
		return "", err
	}

	if len(matches) > 1 {
		return "", fmt.Errorf("multiple sol.* files found")
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no sol.* files found")
	}

	return matches[0], nil
}

func (a *App) resolveSolFile(solFile string) (string, error) {
	if solFile == "" {
		solFile, err := a.defaultSolFile()
		if err != nil {
			return "", fmt.Errorf("submit: %w", err)
		}

		return solFile, nil
	}

	if filepath.IsAbs(solFile) {
		return solFile, nil
	}

	return filepath.Join(a.WorkingDir, solFile), nil
}

func (a *App) Submit(ctx context.Context, solFile string) error {
	client := submitv1connect.NewSubmitServiceClient(
		a.HTTPClient,
		"http://localhost:"+a.Cfg.Port,
	)

	fs := afero.NewBasePathFs(afero.NewOsFs(), a.WorkingDir)

	ws, err := workspace.Open(fs)
	if err != nil {
		return fmt.Errorf("submit: %w", err)
	}

	solFile, err = a.resolveSolFile(solFile)
	if err != nil {
		return fmt.Errorf("submit: %w", err)
	}

	content, err := os.ReadFile(solFile)
	if err != nil {
		return fmt.Errorf("submit: %w", err)
	}

	if _, err := client.Submit(ctx, &submitv1.SubmitRequest{
		ProblemId: ws.Problem().Id,
		FileName:  filepath.Base(solFile),
		Content:   content,
	}); err != nil {
		return fmt.Errorf("submit: %w", err)
	}

	return nil
}
