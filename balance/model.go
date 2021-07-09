package balance

import (
	"github.com/shopspring/decimal"
	"sync"
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
