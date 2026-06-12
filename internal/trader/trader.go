// Package trader 是系统唯一的下单出口。
// 当前提供 dry-run 模拟通道；实盘接入 QMT xttrader（仅 Windows）时
// 新增 qmtTrader 实现同一接口即可，上层逻辑不变。
package trader

import (
	"fmt"
)

// Order 下单请求。
type Order struct {
	SignalID int64
	Code     string
	Action   string // BUY / SELL
	Price    float64
	Qty      int
}

// Trader 下单通道接口。
type Trader interface {
	// Place 执行下单，返回成交状态（如 FILLED）。
	Place(o *Order) (status string, err error)
}

// New 按配置 trade.mode 创建下单通道。
func New(mode string) (Trader, error) {
	switch mode {
	case "", "dry-run":
		return &dryRunTrader{}, nil
	case "qmt":
		return nil, fmt.Errorf("qmt trader not implemented yet (requires Windows + xtquant)")
	default:
		return nil, fmt.Errorf("unknown trade mode: %s", mode)
	}
}

// dryRunTrader 模拟下单：不触达任何券商接口，按信号价即时成交。
type dryRunTrader struct{}

func (t *dryRunTrader) Place(o *Order) (string, error) {
	fmt.Printf("[dry-run] %s %s qty=%d price=%.2f (signal #%d)\n",
		o.Action, o.Code, o.Qty, o.Price, o.SignalID)
	return "FILLED", nil
}
