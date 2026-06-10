package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type DBConfig struct {
	Path string `yaml:"path"`
}

type DataConfig struct {
	Source    string   `yaml:"source"`
	StartDate string   `yaml:"start_date"`
	Symbols   []string `yaml:"symbols"`
}

type Config struct {
	DB   DBConfig   `yaml:"db"`
	Data DataConfig `yaml:"data"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if !filepath.IsAbs(cfg.DB.Path) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working dir: %w", err)
		}
		cfg.DB.Path = filepath.Join(cwd, cfg.DB.Path)
	}

	return cfg, nil
}
