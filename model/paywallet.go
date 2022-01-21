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

// WalletArgument 出款钱包参数
type WalletArgument struct {
	Protocol  Protocol // 协议
	Symbol    string   // 币种
	WalletKey string   // 钱包业务参数key
}

func NewWalletArgument(protocol Protocol, symbol string, walletKey string) *WalletArgument {
	return &WalletArgument{Protocol: protocol, Symbol: symbol, WalletKey: walletKey}
}
