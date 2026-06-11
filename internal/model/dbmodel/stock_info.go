package dbmodel

// StockInfo 股票基本信息。
type StockInfo struct {
	Code     string `gorm:"column:code;primaryKey" json:"code"`
	Name     string `gorm:"column:name" json:"name"`
	Market   string `gorm:"column:market" json:"market"`
	Industry string `gorm:"column:industry" json:"industry"`
	ListDate string `gorm:"column:list_date" json:"list_date"`
}

func (StockInfo) TableName() string {
	return "stock_info_tab"
}
