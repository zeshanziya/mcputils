package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zeshanziya/mcputils/internal/core"
)

var serverEnableCmd = &cobra.Command{
	Use:   "enable [name]",
	Short: "Enable a single server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sc, err := resolveScope()
		if err != nil {
			return err
		}
		cat, err := loadCatalog(sc)
		if err != nil {
			return err
		}
		ad := newAdapter(cat)
		hist := newHistoryLogger()
		return core.EnableServer(cat, ad, hist, sc, args[0])
	},
}

func init() {
	serverCmd.AddCommand(serverEnableCmd)
}
