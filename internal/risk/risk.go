// Package risk 实现下单前的风控检查。Python 只产出信号，
// 任何信号必须通过本层检查才能到达下单通道 —— 策略代码没有能力直接亏钱。
package risk

import (
	"fmt"

	"quantify/internal/config"
	"quantify/internal/db"
)

// Checker 风控检查器。基于持仓表与配置做事前检查。
type Checker struct {
	cfg       config.RiskConfig
	positions db.PositionRepo
}

func NewChecker(cfg config.RiskConfig, positions db.PositionRepo) *Checker {
	return &Checker{cfg: cfg, positions: positions}
}

// usedCapital 返回当前全部持仓占用的资金（按成本计）。
func (c *Checker) usedCapital() (float64, error) {
	all, err := c.positions.FindAll()
	if err != nil {
		return 0, err
	}
	var used float64
	for _, p := range all {
		used += p.AvgCost * float64(p.Qty)
	}
	return used, nil
}

// CheckBuy 买入前检查：可用资金是否充足、单票仓位是否超限。
// TODO: 日亏损上限（max_daily_loss_pct）需要实时行情计算当日浮亏，接入 QMT 行情后实现。
func (c *Checker) CheckBuy(code string, price float64, qty int) error {
	cost := price * float64(qty)

	used, err := c.usedCapital()
	if err != nil {
		return fmt.Errorf("risk: load positions: %w", err)
	}
	if available := c.cfg.Capital - used; cost > available {
		return fmt.Errorf("insufficient funds: need %.2f, available %.2f", cost, available)
	}

	pos, err := c.positions.Get(code)
	if err != nil {
		return fmt.Errorf("risk: get position: %w", err)
	}
	posValue := cost
	if pos != nil {
		posValue += pos.AvgCost * float64(pos.Qty)
	}
	if limit := c.cfg.Capital * c.cfg.MaxPositionPct; posValue > limit {
		return fmt.Errorf("position limit exceeded: %s would hold %.2f, limit %.2f (%.0f%%)",
			code, posValue, limit, c.cfg.MaxPositionPct*100)
	}

	return nil
}

// CheckSell 卖出前检查：持仓是否足够。
func (c *Checker) CheckSell(code string, qty int) error {
	pos, err := c.positions.Get(code)
	if err != nil {
		return fmt.Errorf("risk: get position: %w", err)
	}
	if pos == nil || pos.Qty < qty {
		held := 0
		if pos != nil {
			held = pos.Qty
		}
		return fmt.Errorf("insufficient position: sell %d %s but hold %d", qty, code, held)
	}
	return nil
}
