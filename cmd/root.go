package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	globalFlag  bool
	projectFlag string
)

var rootCmd = &cobra.Command{
	Use:   "mcputils",
	Short: "Manage MCP server profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&globalFlag, "global", false, "operate on the global scope")
	rootCmd.PersistentFlags().StringVar(&projectFlag, "project", "", "operate on a project scope (optionally an explicit path)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
