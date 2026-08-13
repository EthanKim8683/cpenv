package command

import (
	"github.com/EthanKim8683/cpenv/internal/cli"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB
var c *cli.CLI

var rootCmd = &cobra.Command{
	Use:   "cpenv",
	Short: "Competitive programming environment CLI.",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		// cfg, err := config.Load()
		// if err != nil {
		// 	return err
		// }

		// db, err = bolt.Open(filepath.Join(xdg.StateHome, "cpenv", "cpenv.db"), 0600, nil)
		// if err != nil {
		// 	return err
		// }

		// focusStore := &focus.FileStore{Path: filepath.Join(xdg.StateHome, "cpenv", "focus.json")}
		// submissionStore := &submission.DBStore{DB: db}
		// submitClient := submitv1connect.NewSubmitServiceClient(
		// 	http.DefaultClient,
		// 	"http://localhost:"+cfg.Port,
		// )

		// c = &cli.CLI{
		// 	Cfg:             cfg,
		// 	FocusStore:      focusStore,
		// 	SubmissionStore: submissionStore,
		// 	SubmitClient:    submitClient,
		// }

		// cwd, err := os.Getwd()
		// if err != nil {
		// 	return err
		// }
		// w, err = cli.NewWorkspace(
		// 	cwd,
		// 	submissionStore,
		// 	submitClient,
		// 	c.NewHome(),
		// )
		// if err != nil && errors.Is(err, os.ErrNotExist) {
		// 	return err
		// }

		return nil
	},
	PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
		return db.Close()
	},
}

func Execute() error {
	return rootCmd.Execute()
}
