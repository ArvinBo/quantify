package db

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"quantify/internal/model/dbmodel"
)

func Open(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	return db, nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	return sqlDB.Close()
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&dbmodel.DailyKline{}); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

type Stats struct {
	SymbolCount int
	TotalRows   int
	MinDate     string
	MaxDate     string
}

func GetStats(db *gorm.DB) (*Stats, error) {
	s := &Stats{}

	var symbolCount int64
	if err := db.Model(&dbmodel.DailyKline{}).Distinct("code").Count(&symbolCount).Error; err != nil {
		return nil, fmt.Errorf("count symbols: %w", err)
	}
	s.SymbolCount = int(symbolCount)

	var totalRows int64
	if err := db.Model(&dbmodel.DailyKline{}).Count(&totalRows).Error; err != nil {
		return nil, fmt.Errorf("count rows: %w", err)
	}
	s.TotalRows = int(totalRows)

	var dr struct {
		MinDate string
		MaxDate string
	}
	if err := db.Model(&dbmodel.DailyKline{}).
		Select("MIN(trade_date) as min_date, MAX(trade_date) as max_date").
		Scan(&dr).Error; err != nil {
		return nil, fmt.Errorf("date range: %w", err)
	}
	s.MinDate = dr.MinDate
	s.MaxDate = dr.MaxDate

	return s, nil
}
