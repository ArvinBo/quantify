package db

import (
	"fmt"

	"gorm.io/gorm"

	"quantify/internal/model/dbmodel"
)

// Stats 日线数据概览统计信息。
type Stats struct {
	SymbolCount int
	TotalRows   int
	MinDate     string
	MaxDate     string
}

// DailyKlineRepo 日线数据仓储接口，定义数据写入与查询方法。
type DailyKlineRepo interface {
	// Upsert 写入或替换单条日线数据（INSERT OR REPLACE）。
	Upsert(kline *dbmodel.DailyKline) error
	// BatchUpsert 批量写入或替换日线数据，使用事务保证原子性。
	BatchUpsert(klines []*dbmodel.DailyKline) error

	// FindByCode 按标的代码查询全部历史日线，按日期升序返回。
	FindByCode(code string) ([]*dbmodel.DailyKline, error)
	// FindByCodeAndDateRange 按标的代码和日期区间查询日线，按日期升序返回。
	FindByCodeAndDateRange(code, startDate, endDate string) ([]*dbmodel.DailyKline, error)
	// GetAllCodes 返回数据库中所有标的代码。
	GetAllCodes() ([]string, error)
	// GetLatestTradeDate 返回指定标的的最新交易日。
	GetLatestTradeDate(code string) (string, error)

	// GetStats 返回数据库概览统计（标的数、总行数、日期范围）。
	GetStats() (*Stats, error)
	// GetTopVolumeByDate 返回指定日期成交量最高的 N 条日线，按成交量降序。
	GetTopVolumeByDate(date string, limit int) ([]*dbmodel.DailyKline, error)
}

type dailyKlineRepo struct {
	db *gorm.DB
}

// NewDailyKlineRepo 创建日线数据仓储的 GORM 实现。
func NewDailyKlineRepo(db *gorm.DB) DailyKlineRepo {
	return &dailyKlineRepo{db: db}
}

const upsertSQL = `INSERT OR REPLACE INTO daily_kline_tab 
	(code, trade_date, open, high, low, close, volume, amount) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

func (r *dailyKlineRepo) Upsert(kline *dbmodel.DailyKline) error {
	return r.db.Exec(upsertSQL,
		kline.Code, kline.TradeDate, kline.Open, kline.High,
		kline.Low, kline.Close, kline.Volume, kline.Amount,
	).Error
}

func (r *dailyKlineRepo) BatchUpsert(klines []*dbmodel.DailyKline) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, k := range klines {
			if err := tx.Exec(upsertSQL,
				k.Code, k.TradeDate, k.Open, k.High,
				k.Low, k.Close, k.Volume, k.Amount,
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *dailyKlineRepo) FindByCode(code string) ([]*dbmodel.DailyKline, error) {
	var klines []*dbmodel.DailyKline
	if err := r.db.Where("code = ?", code).Order("trade_date asc").Find(&klines).Error; err != nil {
		return nil, fmt.Errorf("find by code: %w", err)
	}
	return klines, nil
}

func (r *dailyKlineRepo) FindByCodeAndDateRange(code, startDate, endDate string) ([]*dbmodel.DailyKline, error) {
	var klines []*dbmodel.DailyKline
	if err := r.db.
		Where("code = ? AND trade_date >= ? AND trade_date <= ?", code, startDate, endDate).
		Order("trade_date asc").
		Find(&klines).Error; err != nil {
		return nil, fmt.Errorf("find by code and date range: %w", err)
	}
	return klines, nil
}

func (r *dailyKlineRepo) GetAllCodes() ([]string, error) {
	var codes []string
	if err := r.db.Model(&dbmodel.DailyKline{}).Distinct("code").Pluck("code", &codes).Error; err != nil {
		return nil, fmt.Errorf("get all codes: %w", err)
	}
	return codes, nil
}

func (r *dailyKlineRepo) GetLatestTradeDate(code string) (string, error) {
	var date string
	if err := r.db.Model(&dbmodel.DailyKline{}).
		Where("code = ?", code).
		Select("MAX(trade_date)").
		Scan(&date).Error; err != nil {
		return "", fmt.Errorf("get latest trade date: %w", err)
	}
	return date, nil
}

func (r *dailyKlineRepo) GetStats() (*Stats, error) {
	s := &Stats{}

	var symbolCount int64
	if err := r.db.Model(&dbmodel.DailyKline{}).Distinct("code").Count(&symbolCount).Error; err != nil {
		return nil, fmt.Errorf("count symbols: %w", err)
	}
	s.SymbolCount = int(symbolCount)

	var totalRows int64
	if err := r.db.Model(&dbmodel.DailyKline{}).Count(&totalRows).Error; err != nil {
		return nil, fmt.Errorf("count rows: %w", err)
	}
	s.TotalRows = int(totalRows)

	var dr struct {
		MinDate string
		MaxDate string
	}
	if err := r.db.Model(&dbmodel.DailyKline{}).
		Select("MIN(trade_date) as min_date, MAX(trade_date) as max_date").
		Scan(&dr).Error; err != nil {
		return nil, fmt.Errorf("date range: %w", err)
	}
	s.MinDate = dr.MinDate
	s.MaxDate = dr.MaxDate

	return s, nil
}

func (r *dailyKlineRepo) GetTopVolumeByDate(date string, limit int) ([]*dbmodel.DailyKline, error) {
	var klines []*dbmodel.DailyKline
	if err := r.db.Where("trade_date = ?", date).Order("volume desc").Limit(limit).Find(&klines).Error; err != nil {
		return nil, fmt.Errorf("get top volume by date: %w", err)
	}
	return klines, nil
}
