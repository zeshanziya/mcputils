package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listServersCmd = &cobra.Command{
	Use:   "servers",
	Short: "List all catalog servers with enabled/disabled state",
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

		for _, s := range cat.Servers {
			enabled, err := ad.IsEnabled(sc, s.Name)
			if err != nil {
				return err
			}
			state := "disabled"
			if enabled {
				state = "enabled"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", s.Name, state)
		}
		return nil
	},
}

func init() {
	listCmd.AddCommand(listServersCmd)
}
