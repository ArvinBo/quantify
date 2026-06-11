package db

import (
	"fmt"

	"gorm.io/gorm"

	"quantify/internal/model/dbmodel"
)

// StockInfoRepo 股票基本信息仓储接口。
type StockInfoRepo interface {
	// Upsert 写入或替换单条股票信息（INSERT OR REPLACE）。
	Upsert(info *dbmodel.StockInfo) error
	// BatchUpsert 批量写入或替换股票信息，使用事务保证原子性。
	BatchUpsert(infos []*dbmodel.StockInfo) error

	// FindByCode 按代码查询股票基本信息。
	FindByCode(code string) (*dbmodel.StockInfo, error)
	// FindAll 返回全部股票基本信息。
	FindAll() ([]*dbmodel.StockInfo, error)
	// FindByMarket 按市场（SH / SZ / BJ）筛选。
	FindByMarket(market string) ([]*dbmodel.StockInfo, error)
	// FindByIndustry 按行业模糊匹配。
	FindByIndustry(industry string) ([]*dbmodel.StockInfo, error)
}

type stockInfoRepo struct {
	db *gorm.DB
}

// NewStockInfoRepo 创建股票信息仓储的 GORM 实现。
func NewStockInfoRepo(db *gorm.DB) StockInfoRepo {
	return &stockInfoRepo{db: db}
}

const upsertStockInfoSQL = `INSERT OR REPLACE INTO stock_info_tab 
	(code, name, market, industry, list_date) 
	VALUES (?, ?, ?, ?, ?)`

func (r *stockInfoRepo) Upsert(info *dbmodel.StockInfo) error {
	return r.db.Exec(upsertStockInfoSQL,
		info.Code, info.Name, info.Market, info.Industry, info.ListDate,
	).Error
}

func (r *stockInfoRepo) BatchUpsert(infos []*dbmodel.StockInfo) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, info := range infos {
			if err := tx.Exec(upsertStockInfoSQL,
				info.Code, info.Name, info.Market, info.Industry, info.ListDate,
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *stockInfoRepo) FindByCode(code string) (*dbmodel.StockInfo, error) {
	var info dbmodel.StockInfo
	if err := r.db.Where("code = ?", code).First(&info).Error; err != nil {
		return nil, fmt.Errorf("find stock info by code: %w", err)
	}
	return &info, nil
}

func (r *stockInfoRepo) FindAll() ([]*dbmodel.StockInfo, error) {
	var infos []*dbmodel.StockInfo
	if err := r.db.Order("code asc").Find(&infos).Error; err != nil {
		return nil, fmt.Errorf("find all stock info: %w", err)
	}
	return infos, nil
}

func (r *stockInfoRepo) FindByMarket(market string) ([]*dbmodel.StockInfo, error) {
	var infos []*dbmodel.StockInfo
	if err := r.db.Where("market = ?", market).Order("code asc").Find(&infos).Error; err != nil {
		return nil, fmt.Errorf("find stock info by market: %w", err)
	}
	return infos, nil
}

func (r *stockInfoRepo) FindByIndustry(industry string) ([]*dbmodel.StockInfo, error) {
	var infos []*dbmodel.StockInfo
	if err := r.db.Where("industry LIKE ?", "%"+industry+"%").Order("code asc").Find(&infos).Error; err != nil {
		return nil, fmt.Errorf("find stock info by industry: %w", err)
	}
	return infos, nil
}
