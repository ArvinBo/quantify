package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"quantify/internal/db"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show database status",
	Long:  `Display an overview of the stored daily kline data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(); err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		database, err := db.Open(cfg.DB.Path)
		if err != nil {
			return fmt.Errorf("open db: %w", err)
		}
		defer database.Close()

		stats, err := db.GetStats(database)
		if err != nil {
			return fmt.Errorf("get stats: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintf(w, "Database:\t%s\n", cfg.DB.Path)
		fmt.Fprintf(w, "Symbols:\t%d\n", stats.SymbolCount)
		fmt.Fprintf(w, "Total Rows:\t%d\n", stats.TotalRows)
		if stats.MinDate != "" {
			fmt.Fprintf(w, "Date Range:\t%s ~ %s\n", stats.MinDate, stats.MaxDate)
		} else {
			fmt.Fprintf(w, "Date Range:\t(empty)\n")
		}
		w.Flush()

		return nil
	},
}
