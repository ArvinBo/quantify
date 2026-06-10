package dbmodel

type DailyKline struct {
	Code      string  `gorm:"column:code;primaryKey" json:"code"`
	TradeDate string  `gorm:"column:trade_date;primaryKey" json:"trade_date"`
	Open      float64 `gorm:"column:open" json:"open"`
	High      float64 `gorm:"column:high" json:"high"`
	Low       float64 `gorm:"column:low" json:"low"`
	Close     float64 `gorm:"column:close" json:"close"`
	Volume    float64 `gorm:"column:volume" json:"volume"`
	Amount    float64 `gorm:"column:amount" json:"amount"`
}

func (DailyKline) TableName() string {
	return "daily_kline_tab"
}
