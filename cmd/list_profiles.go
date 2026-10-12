package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List all profiles with active/inactive state",
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

		for _, p := range cat.Profiles {
			state := "inactive"
			if st.IsActive(p.Name) {
				state = "active"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", p.Name, state)
		}
		return nil
	},
}

func init() {
	listCmd.AddCommand(listProfilesCmd)
}
