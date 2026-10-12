package cmd

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List servers or profiles",
}

func init() {
	rootCmd.AddCommand(listCmd)
}
