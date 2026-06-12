package db

import (
	"fmt"

	"gorm.io/gorm"

	"quantify/internal/model/dbmodel"
)

// OrderRepo 订单仓储接口。
type OrderRepo interface {
	// Create 写入一条订单记录。
	Create(order *dbmodel.Order) error
	// FindBySignalID 按信号 ID 查询订单。
	FindBySignalID(signalID int64) ([]*dbmodel.Order, error)
}

type orderRepo struct {
	db *gorm.DB
}

// NewOrderRepo 创建订单仓储的 GORM 实现。
func NewOrderRepo(db *gorm.DB) OrderRepo {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(order *dbmodel.Order) error {
	if err := r.db.Omit("created_at").Create(order).Error; err != nil {
		return fmt.Errorf("create order: %w", err)
	}
	return nil
}

func (r *orderRepo) FindBySignalID(signalID int64) ([]*dbmodel.Order, error) {
	var orders []*dbmodel.Order
	if err := r.db.Where("signal_id = ?", signalID).Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("find orders by signal id: %w", err)
	}
	return orders, nil
}
