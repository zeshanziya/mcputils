package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show combined profiles, servers, and effective enabled state",
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

		out := cmd.OutOrStdout()
		for _, p := range cat.Profiles {
			profileState := "inactive"
			if st.IsActive(p.Name) {
				profileState = "active"
			}
			fmt.Fprintf(out, "profile %s: %s\n", p.Name, profileState)
			for _, serverName := range p.Servers {
				enabled, err := ad.IsEnabled(sc, serverName)
				if err != nil {
					return err
				}
				serverState := "disabled"
				if enabled {
					serverState = "enabled"
				}
				fmt.Fprintf(out, "  %s: %s\n", serverName, serverState)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
