package db

import (
	"fmt"

	"gorm.io/gorm"

	"quantify/internal/model/dbmodel"
)

// SignalRepo 策略信号仓储接口。Python 写入信号，Go 通过本接口消费。
type SignalRepo interface {
	// FindPending 返回全部待处理信号，按创建时间升序。
	FindPending() ([]*dbmodel.Signal, error)
	// UpdateStatus 更新信号状态与原因（EXECUTED / REJECTED）。
	UpdateStatus(id int64, status, reason string) error
}

type signalRepo struct {
	db *gorm.DB
}

// NewSignalRepo 创建信号仓储的 GORM 实现。
func NewSignalRepo(db *gorm.DB) SignalRepo {
	return &signalRepo{db: db}
}

func (r *signalRepo) FindPending() ([]*dbmodel.Signal, error) {
	var signals []*dbmodel.Signal
	if err := r.db.
		Where("status = ?", dbmodel.SignalStatusPending).
		Order("created_at asc").
		Find(&signals).Error; err != nil {
		return nil, fmt.Errorf("find pending signals: %w", err)
	}
	return signals, nil
}

func (r *signalRepo) UpdateStatus(id int64, status, reason string) error {
	if err := r.db.Model(&dbmodel.Signal{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "reason": reason}).Error; err != nil {
		return fmt.Errorf("update signal status: %w", err)
	}
	return nil
}
