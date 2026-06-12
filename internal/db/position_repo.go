package db

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"quantify/internal/model/dbmodel"
)

// PositionRepo 持仓仓储接口。
type PositionRepo interface {
	// Get 查询单只标的持仓，无持仓返回 nil。
	Get(code string) (*dbmodel.Position, error)
	// FindAll 返回全部持仓。
	FindAll() ([]*dbmodel.Position, error)
	// ApplyFill 按成交更新持仓：买入加仓重算成本，卖出减仓，清零删除。
	ApplyFill(code, action string, price float64, qty int) error
}

type positionRepo struct {
	db *gorm.DB
}

// NewPositionRepo 创建持仓仓储的 GORM 实现。
func NewPositionRepo(db *gorm.DB) PositionRepo {
	return &positionRepo{db: db}
}

func (r *positionRepo) Get(code string) (*dbmodel.Position, error) {
	var pos dbmodel.Position
	err := r.db.Where("code = ?", code).First(&pos).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get position: %w", err)
	}
	return &pos, nil
}

func (r *positionRepo) FindAll() ([]*dbmodel.Position, error) {
	var positions []*dbmodel.Position
	if err := r.db.Order("code asc").Find(&positions).Error; err != nil {
		return nil, fmt.Errorf("find all positions: %w", err)
	}
	return positions, nil
}

func (r *positionRepo) ApplyFill(code, action string, price float64, qty int) error {
	pos, err := r.Get(code)
	if err != nil {
		return err
	}

	switch action {
	case dbmodel.ActionBuy:
		newQty := qty
		newCost := price
		if pos != nil {
			newQty = pos.Qty + qty
			newCost = (pos.AvgCost*float64(pos.Qty) + price*float64(qty)) / float64(newQty)
		}
		return r.db.Exec(
			`INSERT OR REPLACE INTO position_tab (code, qty, avg_cost, updated_at)
			 VALUES (?, ?, ?, datetime('now', 'localtime'))`,
			code, newQty, newCost,
		).Error

	case dbmodel.ActionSell:
		if pos == nil || pos.Qty < qty {
			return fmt.Errorf("sell %d %s but position is insufficient", qty, code)
		}
		remain := pos.Qty - qty
		if remain == 0 {
			return r.db.Where("code = ?", code).Delete(&dbmodel.Position{}).Error
		}
		return r.db.Exec(
			`UPDATE position_tab SET qty = ?, updated_at = datetime('now', 'localtime') WHERE code = ?`,
			remain, code,
		).Error

	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}
