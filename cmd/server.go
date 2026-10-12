package cmd

import "github.com/spf13/cobra"

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage individual servers",
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
