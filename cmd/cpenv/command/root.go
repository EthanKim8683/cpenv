package command

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/EthanKim8683/cpenv/internal/cli"
	"github.com/EthanKim8683/cpenv/internal/config"
	focusv1connect "github.com/EthanKim8683/cpenv/internal/gen/Focus/v1/Focusv1connect"
	"github.com/EthanKim8683/cpenv/internal/gen/active_problem/v1/active_problemv1connect"
	"github.com/EthanKim8683/cpenv/internal/gen/submissions/v1/submissionsv1connect"
	"github.com/EthanKim8683/cpenv/internal/gen/submit/v1/submitv1connect"
	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var c *cli.CLI
var db *bolt.DB

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

		dbPath := filepath.Join(xdg.StateHome, "cpenv", "cli.db")
		if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
			return err
		}

		db, err = bolt.Open(dbPath, 0600, nil)
		if err != nil {
			return err
		}

		baseURL := fmt.Sprintf("http://localhost:%d", cfg.Port)
		activeProblemClient := active_problemv1connect.NewActiveProblemServiceClient(http.DefaultClient, baseURL)
		focusClient := focusv1connect.NewFocusServiceClient(http.DefaultClient, baseURL)
		submissionsClient := submissionsv1connect.NewSubmissionsServiceClient(http.DefaultClient, baseURL)
		submitClient := submitv1connect.NewSubmitServiceClient(http.DefaultClient, baseURL)
		prefs := &cli.DBPreferences{DB: db}

		c = &cli.CLI{
			Cfg:                 cfg,
			CWD:                 cwd,
			ActiveProblemClient: activeProblemClient,
			FocusClient:         focusClient,
			SubmissionsClient:   submissionsClient,
			SubmitClient:        submitClient,
			Preferences:         prefs,
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
