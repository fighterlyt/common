package balance

import (
	"github.com/fighterlyt/log"
	"gitlab.com/nova_dubai/common/durablemetrics"
)

type metrics struct {
	collectWalletBalance  *durablemetrics.GaugeVec
	withdrawWalletBalance *durablemetrics.GaugeVec
}

func newMetrics(logger log.Logger) (*metrics, error) {
	collectGaugeVec, err := durablemetrics.NewGaugeVec("collect_balances", "归集钱包余额", []string{"currency"}, logger)
	if err != nil {
		return nil, err
	}

	withdrawGaugeVec, err := durablemetrics.NewGaugeVec("withdraw_balances", "提现钱包钱包余额", []string{"currency"}, logger)
	if err != nil {
		return nil, err
	}

	return &metrics{collectWalletBalance: collectGaugeVec, withdrawWalletBalance: withdrawGaugeVec}, nil
}
