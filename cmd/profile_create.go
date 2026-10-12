package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zeshanziya/mcputils/internal/core"
)

var profileCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new empty profile",
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
		if err := core.CreateProfile(cat, args[0]); err != nil {
			return err
		}
		return cat.Save(sc.CatalogPath)
	},
}

func init() {
	profileCmd.AddCommand(profileCreateCmd)
}
