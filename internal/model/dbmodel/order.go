package dbmodel

// Order 实盘订单记录，由 Go 写入（唯一下单出口）。
type Order struct {
	ID        int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SignalID  int64   `gorm:"column:signal_id" json:"signal_id"`
	Code      string  `gorm:"column:code" json:"code"`
	Action    string  `gorm:"column:action" json:"action"`
	Price     float64 `gorm:"column:price" json:"price"`
	Qty       int     `gorm:"column:qty" json:"qty"`
	Status    string  `gorm:"column:status" json:"status"`
	CreatedAt string  `gorm:"column:created_at" json:"created_at"`
}

func (Order) TableName() string {
	return "order_tab"
}
