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

func (a *App) defaultSubmissionFile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	matches, err := filepath.Glob(filepath.Join(dir, "sol.*"))
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

func (a *App) Submit(ctx context.Context, subFile string) error {
	client := submitv1connect.NewSubmitServiceClient(
		a.HTTPClient,
		"http://localhost:"+a.Cfg.Port,
	)

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("submit: %w", err)
	}

	fs := afero.NewBasePathFs(afero.NewOsFs(), dir)

	ws, err := workspace.Open(fs)
	if err != nil {
		return fmt.Errorf("submit: %w", err)
	}

	if subFile == "" {
		subFile, err = a.defaultSubmissionFile()
		if err != nil {
			return fmt.Errorf("submit: %w", err)
		}
	} else if !filepath.IsAbs(subFile) {
		subFile, err = filepath.Abs(subFile)
		if err != nil {
			return fmt.Errorf("submit: %w", err)
		}
	}

	content, err := os.ReadFile(subFile)
	if err != nil {
		return fmt.Errorf("submit: %w", err)
	}

	if _, err := client.Submit(ctx, &submitv1.SubmitRequest{
		ProblemId: ws.Problem().Id,
		FileName:  filepath.Base(subFile),
		Content:   content,
	}); err != nil {
		return fmt.Errorf("submit: %w", err)
	}

	return nil
}
