package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zeshanziya/mcputils/internal/core"
)

var profileRemoveServerCmd = &cobra.Command{
	Use:   "remove-server [profile] [server]",
	Short: "Remove a server from a profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sc, err := resolveScope()
		if err != nil {
			return err
		}
		cat, err := loadCatalog(sc)
		if err != nil {
			return err
		}
		if err := core.RemoveServerFromProfile(cat, args[0], args[1]); err != nil {
			return err
		}
		return cat.Save(sc.CatalogPath)
	},
}

func init() {
	profileCmd.AddCommand(profileRemoveServerCmd)
}
