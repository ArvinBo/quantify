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

type TushareConfig struct {
	Token string `yaml:"token"`
}

type DataConfig struct {
	Source    string        `yaml:"source"`
	StartDate string        `yaml:"start_date"`
	Symbols   []string      `yaml:"symbols"`
	Tushare   TushareConfig `yaml:"tushare"`
}

type BacktestConfig struct {
	Capital        float64 `yaml:"capital"`
	CommissionRate float64 `yaml:"commission_rate"`
	CommissionMin  float64 `yaml:"commission_min"`
	StampTaxRate   float64 `yaml:"stamp_tax_rate"`
}

type RiskConfig struct {
	Capital         float64 `yaml:"capital"`
	MaxPositionPct  float64 `yaml:"max_position_pct"`
	MaxDailyLossPct float64 `yaml:"max_daily_loss_pct"`
}

type TradeConfig struct {
	Mode string `yaml:"mode"` // dry-run | qmt
}

type Config struct {
	DB       DBConfig       `yaml:"db"`
	Data     DataConfig     `yaml:"data"`
	Backtest BacktestConfig `yaml:"backtest"`
	Risk     RiskConfig     `yaml:"risk"`
	Trade    TradeConfig    `yaml:"trade"`
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
