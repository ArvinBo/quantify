package cmd

import (
	"fmt"

	"quantify/internal/config"
	"quantify/internal/db"
)

func InitDB(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	database, err := db.Open(cfg.DB.Path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close(database)

	fmt.Printf("database connection ok: %s\n", cfg.DB.Path)
	return nil
}
