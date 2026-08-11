package command

import (
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var statusLimit int

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Tail submission results.",
	Long:  "Tail submission results for the problem corresponding to the current workspace, or globally if not in a workspace.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		subs, err := c.Status(statusLimit)
		if err != nil {
			return err
		}

		var rows [][]string
		for _, sub := range subs {
			rows = append(rows, []string{
				time.UnixMilli(sub.GetTimestampMs()).Format("Jan 2, 2006 3:04:05 PM"),
				sub.GetProblemId(),
				sub.GetStatus().String(),
				strconv.FormatUint(uint64(sub.GetTimeMs()), 10),
				strconv.FormatUint(uint64(sub.GetMemoryKb()), 10),
			})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"timestamp (ms)", "problem ID", "status", "time (ms)", "memory (kB)"})
		if err := table.Bulk(rows); err != nil {
			return err
		}
		if err := table.Render(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
	statusCmd.Flags().IntVarP(&statusLimit, "limit", "l", 10, "maximum number of submission results to output")
}
