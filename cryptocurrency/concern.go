package cryptocurrency

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

var (
	detailFMT = `%s->%s,协议[%s],币种[%s],金额[%s],交易时间[%s],真实手续费[%s]`
)

type Concern interface {
	// FilterConcernedAccounts 过滤感兴趣的账户,交易服务仅保留感兴趣的账户对应的交易记录
	FilterConcernedAccounts(from, to string, amount decimal.Decimal) (matched bool, data interface{}, err error)

	// 获取账户余额
}

// TradeNotify 交易细节通知
type TradeNotify interface {
	Notify(protocol Protocol, details []*TransactionDetail) error
}

// TradeBusinessDetail 交易细节
type TradeBusinessDetail struct {
	From                  string          `json:"fromAddress"`             // 支付地址
	To                    string          `json:"toAddress"`               // 收款地址
	Token                 string          `json:"token"`                   // 币种
	Protocol              Protocol        `json:"protocol"`                // 协议
	Amount                decimal.Decimal `json:"amount"`                  // 金额
	RealTransactionCharge decimal.Decimal `json:"real_transaction_charge"` // 真实交易手续费,主链货币
	RealTransactionUSDT   decimal.Decimal `json:"real_transaction_usdt"`   // 真实交易手续费,USDT
	DealTime              time.Time       `json:"-"`                       // 交易时间
	Time                  int64           `json:"time"`                    // 交易时间
	Data                  interface{}     `json:"data"`                    // 额外数据
}

func NewTradeBusinessDetail(from, to, token string, protocol Protocol, amount, realTransactionCharge decimal.Decimal, dealTime time.Time, data interface{}) *TradeBusinessDetail { // nolint:golint,lll
	return &TradeBusinessDetail{
		From:                  from,
		To:                    to,
		Token:                 token,
		Protocol:              protocol,
		Amount:                amount,
		RealTransactionCharge: realTransactionCharge,
		DealTime:              dealTime,
		Data:                  data,
	}
}

func (t TradeBusinessDetail) String() string {
	return fmt.Sprintf(detailFMT, t.From, t.To, t.Protocol, t.Token, t.Amount.String(), t.DealTime.Format("20060102 15:04:05"), t.RealTransactionCharge.String()) // nolint:golint,lll
}
