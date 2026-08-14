package command

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/EthanKim8683/cpenv/internal/cli"
	"github.com/EthanKim8683/cpenv/internal/config"
	extension "github.com/EthanKim8683/cpenv/internal/daemon"
	"github.com/EthanKim8683/cpenv/internal/gen/submit/v1/submitv1connect"
	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB
var c *cli.CLI

var rootCmd = &cobra.Command{
	Use:   "cpenv",
	Short: "Competitive programming environment CLI.",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		dbPath := filepath.Join(xdg.StateHome, "cpenv", "cpenv.db")
		if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
			return err
		}

		db, err = bolt.Open(dbPath, 0600, nil)
		if err != nil {
			return err
		}

		submitClient := submitv1connect.NewSubmitServiceClient(
			http.DefaultClient,
			fmt.Sprintf("http://localhost:%d", cfg.Port),
		)

		focusedProblemLoader := &extension.FocusedProblemLoader{DB: db}
		submissionsTailer := &extension.SubmissionsTailer{DB: db}
		submitter := &extension.Submitter{Client: submitClient}

		c = &cli.CLI{
			Cfg:            cfg,
			DB:             db,
			CWD:            cwd,
			FocusedProblem: focusedProblemLoader,
			Submissions:    submissionsTailer,
			Submitter:      submitter,
		}
		return nil
	},
	PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
		return db.Close()
	},
}

func Execute() error {
	return rootCmd.Execute()
}
