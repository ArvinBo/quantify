package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"quantify/internal/db"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize quantify project",
	Long:  `Create required directories, database, and copy default config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		base := projectDir()

		dirs := []string{
			"data",
			"logs",
		}
		for _, d := range dirs {
			p := filepath.Join(base, d)
			if err := ensureDir(p); err != nil {
				return fmt.Errorf("create dir %s: %w", p, err)
			}
			fmt.Printf("created directory: %s/\n", d)
		}

		if err := loadConfig(); err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		database, err := db.Open(cfg.DB.Path)
		if err != nil {
			return fmt.Errorf("open db: %w", err)
		}
		defer database.Close()

		if err := db.Migrate(database); err != nil {
			return fmt.Errorf("migrate db: %w", err)
		}
		fmt.Printf("database initialized: %s\n", cfg.DB.Path)

		fmt.Println("\nproject initialized successfully!")
		return nil
	},
}
