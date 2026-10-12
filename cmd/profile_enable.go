package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zeshanziya/mcputils/internal/core"
)

var profileEnableCmd = &cobra.Command{
	Use:   "enable [name]",
	Short: "Enable all servers in a profile",
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
		st, err := loadState(sc)
		if err != nil {
			return err
		}
		ad := newAdapter(cat)
		hist := newHistoryLogger()
		if err := core.EnableProfile(cat, st, ad, hist, sc, args[0]); err != nil {
			return err
		}
		return st.Save(sc.StatePath)
	},
}

func init() {
	profileCmd.AddCommand(profileEnableCmd)
}
