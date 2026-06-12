package dbmodel

// Position 当前持仓，由 Go 在成交后维护。
type Position struct {
	Code      string  `gorm:"column:code;primaryKey" json:"code"`
	Qty       int     `gorm:"column:qty" json:"qty"`
	AvgCost   float64 `gorm:"column:avg_cost" json:"avg_cost"`
	UpdatedAt string  `gorm:"column:updated_at" json:"updated_at"`
}

func (Position) TableName() string {
	return "position_tab"
}
