package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zeshanziya/mcputils/internal/core"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import existing MCP server definitions into the catalog",
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

		result, err := core.ImportServers(cat, ad, sc)
		if err != nil {
			return err
		}

		if len(result.Imported) > 0 {
			if err := cat.Save(sc.CatalogPath); err != nil {
				return err
			}
		}

		out := cmd.OutOrStdout()
		if len(result.Imported) == 0 && len(result.Skipped) == 0 {
			fmt.Fprintln(out, "nothing to import")
			return nil
		}
		if len(result.Imported) > 0 {
			fmt.Fprintf(out, "imported: %s\n", joinNames(result.Imported))
		}
		if len(result.Skipped) > 0 {
			fmt.Fprintf(out, "skipped: %s\n", joinSkipped(result.Skipped))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
