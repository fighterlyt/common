package model

import (
	"github.com/shopspring/decimal"
)

// BalanceDetail 余额详情
type BalanceDetail struct {
	Address     string          `json:"address"`
	Protocol    Protocol        `json:"protocol"`
	USDTBalance decimal.Decimal `json:"usdt_balance"`
	FeeBalance  decimal.Decimal `json:"fee_balance"`
}
