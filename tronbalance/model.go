package tronbalance

import (
	"sync"

	"github.com/shopspring/decimal"
)

// 钱包地址
type walletBalance struct {
	balances map[string]decimal.Decimal
	lock     *sync.RWMutex
}

func newWalletBalance() *walletBalance {
	return &walletBalance{
		balances: make(map[string]decimal.Decimal, 2),
		lock:     &sync.RWMutex{},
	}
}

func (w *walletBalance) reset(balances map[string]decimal.Decimal) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.balances = balances
}

func (w *walletBalance) get(currencies ...string) map[string]decimal.Decimal {
	w.lock.RLock()
	defer w.lock.RUnlock()

	// 交易对为空查询所有
	if len(currencies) == 0 {
		return w.balances
	}

	var result = make(map[string]decimal.Decimal, len(currencies))

	for i := range currencies {
		result[currencies[i]] = w.balances[currencies[i]]
	}

	return result
}

type WalletBalanceInfo struct {
	CollectWalletInfo  WalletInfo `json:"collect_wallet_info"`  // 归集钱包信息
	WithdrawWalletInfo WalletInfo `json:"withdraw_wallet_info"` // 提款钱包信息
}

type WalletInfo struct {
	Address  string                     `json:"address"`  // 地址
	Balances map[string]decimal.Decimal `json:"balances"` // 账户余额信息
}
