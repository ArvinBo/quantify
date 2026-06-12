package dbmodel

// Signal 状态流转：PENDING → EXECUTED / REJECTED。
const (
	SignalStatusPending  = "PENDING"
	SignalStatusExecuted = "EXECUTED"
	SignalStatusRejected = "REJECTED"

	ActionBuy  = "BUY"
	ActionSell = "SELL"
)

// Signal 策略信号，由 Python 写入，Go 消费。
type Signal struct {
	ID        int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Code      string  `gorm:"column:code" json:"code"`
	Action    string  `gorm:"column:action" json:"action"`
	Price     float64 `gorm:"column:price" json:"price"`
	Strategy  string  `gorm:"column:strategy" json:"strategy"`
	TradeDate string  `gorm:"column:trade_date" json:"trade_date"`
	Status    string  `gorm:"column:status" json:"status"`
	Reason    string  `gorm:"column:reason" json:"reason"`
	CreatedAt string  `gorm:"column:created_at" json:"created_at"`
}

func (Signal) TableName() string {
	return "signal_tab"
}
