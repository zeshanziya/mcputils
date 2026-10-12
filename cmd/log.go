package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show enable/disable history",
	RunE: func(cmd *cobra.Command, args []string) error {
		hist := newHistoryLogger()
		entries, err := hist.List()
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		if len(entries) == 0 {
			fmt.Fprintln(out, "no history")
			return nil
		}

		for _, e := range entries {
			if e.SweptProjects > 0 {
				fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\tswept_projects=%d\n",
					e.Timestamp.Format("2006-01-02T15:04:05Z"), e.Scope, e.Trigger, e.Server, e.State, e.SweptProjects)
			} else {
				fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\n",
					e.Timestamp.Format("2006-01-02T15:04:05Z"), e.Scope, e.Trigger, e.Server, e.State)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
