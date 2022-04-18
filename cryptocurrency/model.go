package cryptocurrency

import (
	"fmt"
	"net/url"
	"time"

	"github.com/shopspring/decimal"
)

// TransactionRecord 交易记录
type TransactionRecord interface {
	GetTxID() string       // 交易id
	GetProtocol() Protocol // 协议
	GetTransactionDetail() *TransactionDetail
	SetTransactionDetail(detail *TransactionDetail)
	SetSendResult(txID string, err error)
	SetCheckResult(blockNumber int64, fee decimal.Decimal, reason string)
}

// TransactionDetail 交易详情
type TransactionDetail struct {
	Amount      decimal.Decimal `gorm:"column:amount;type:decimal(40,6);comment:'充值金额'" json:"amount"`
	Protocol    Protocol        `gorm:"column:protocol;comment:'协议类型'" json:"protocol"`
	Symbol      string          `gorm:"column:symbol;comment:'币种'" json:"symbol"`
	From        string          `gorm:"column:from;comment:'转账地址'" json:"from"`
	To          string          `gorm:"column:to;comment:'接收地址'" json:"to"`
	BlockNumber int64           `gorm:"column:block_number;comment:'所在区块'" json:"blockNumber"`
	TxID        string          `gorm:"column:tx_id;comment:'交易号'" json:"txID"`
	Fee         decimal.Decimal `gorm:"column:fee;type:decimal(28,18);comment:'手续费'" json:"fee"`
	Reason      string          `gorm:"column:reason;comment:'失败原因'" json:"reason"`
	TimeStamp   int64           `gorm:"column:timestamp;type bigint(20);comment:时间戳" json:"timeStamp"`
	TradeKind   TradeKind       `gorm:"trade_kind;comment:交易类型" json:"tradeKind"`
}

func NewFullTransactionDetail(amount decimal.Decimal, protocol Protocol, symbol, payAddress, receiveAddress string, blockNumber int64, txID string, fee decimal.Decimal, timestamp int64, kind TradeKind) *TransactionDetail { //nolint:lll
	return &TransactionDetail{
		Amount:      amount,
		Protocol:    protocol,
		Symbol:      symbol,
		From:        payAddress,
		To:          receiveAddress,
		BlockNumber: blockNumber,
		TxID:        txID,
		Fee:         fee,
		TimeStamp:   timestamp,
		TradeKind:   kind,
	}
}

func NewTransactionDetail(amount decimal.Decimal, protocol Protocol, currency, payAddress, receiveAddress string) *TransactionDetail {
	return &TransactionDetail{
		Amount:   amount,
		Protocol: protocol,
		Symbol:   currency,
		From:     payAddress,
		To:       receiveAddress,
	}
}

func (t TransactionDetail) Encode() url.Values {
	value := url.Values{}

	value.Set(`amount`, t.Amount.String())
	value.Set(`protocol`, t.Protocol.String())
	value.Set(`symbol`, t.Symbol)
	value.Set(`from`, t.From)
	value.Set(`to`, t.To)
	value.Set(`blockNumber`, fmt.Sprintf(`%d`, t.BlockNumber))
	value.Set(`txID`, t.TxID)
	value.Set(`fee`, t.Fee.String())
	value.Set(`reason`, t.Reason)
	value.Set(`timeStamp`, fmt.Sprintf(`%d`, t.TimeStamp))

	return value
}

// UserWallet 用户钱包
type UserWallet struct {
	ID         int64    `gorm:"primary_key"`
	UserID     int64    `gorm:"column:user_id;uniqueIndex:userID_protocol;comment:'所属用户ID'"`
	Protocol   Protocol `gorm:"column:protocol;type:varchar(50);uniqueIndex:userID_protocol;comment:'协议类型'"`
	Mnemonic   string   `gorm:"column:mnemonic;comment:'助记词'"`
	PrivateKey string   `gorm:"column:private_key;comment:'私钥'"`
	Address    string   `gorm:"column:address;comment:'地址'"`
}

func (w *UserWallet) TableName() string {
	return "user_wallet"
}

type TradeKind int

const (
	// TradeTransfer 普通转账
	TradeTransfer TradeKind = 1
	// TradeApprove 授权
	TradeApprove TradeKind = 2
	// TradeRevokeApprove 撤销授权
	TradeRevokeApprove TradeKind = 3
)

// Trade tron交易
type Trade struct {
	ID        string          `gorm:"column:ID;primary Key;varchar(64);comment:'ID,对应链上的tx Hash'"` // 交易ID,对应链上的tx Hash
	Protocol  Protocol        `gorm:"column:protocol;comment:'协议'"`                                // 协议
	From      string          `gorm:"column:FROM;comment:'来源地址'"`                                  // 来源地址
	To        string          `gorm:"column:TO;comment:'目标地址'"`                                    // 目标地址
	Amount    decimal.Decimal `gorm:"column:AMOUNT;type:decimal(40,6);comment:'交易金额'"`             // 金额
	Token     string          `gorm:"column:TOKEN;comment:'交易币种'"`                                 // 币种
	Time      int64           `gorm:"column:TIME;comment:'交易时间'"`                                  // 交易时间,这是毫秒,区块链原始数据如此
	BlockNum  int64           `gorm:"column:BLOCK_NUM;comment:'对应的区块编号'"`                          // 对应的区块号码
	Fee       decimal.Decimal `gorm:"column:FEE;type:decimal(30,18);comment:'手续费TRX'"`             // 手续费 trc20是TRX,erc20是gas
	TradeKind TradeKind       `gorm:"column:TRADE_KIND;comment:交易类型"`                              // 交易类型
}

func NewTrade(protocol Protocol, from, to string, amount decimal.Decimal, token, id string, tradeTime, blockNum int64, fee decimal.Decimal, tradeKind TradeKind) *Trade { // nolint:golint,lll
	return &Trade{Protocol: protocol,
		From:      from,
		To:        to,
		Amount:    amount,
		Token:     token,
		ID:        id,
		Time:      tradeTime,
		BlockNum:  blockNum,
		Fee:       fee,
		TradeKind: tradeKind,
	}
}

func (t Trade) TableName() string {
	return "tron_trades"
}

/* -------------------------------区块链事务相关-------------------------------*/

// Transaction 区块链事务
type Transaction interface {
	ContractKind() string // 合约类型
	Success() bool        // 交易是否成功
	Hash() string         // 交易hash
	From() string         // 交易发送方账户地址
	To() string           // 交易接收方账户地址
	// todo： 这两个Value 意义何在
	Value() decimal.Decimal   // 以太币金额
	Token() decimal.Decimal   // 代币金额
	GasUsed() decimal.Decimal // 消耗gas
	StartDate() time.Time     // 交易开始日期
}

// TransactionService 事务查询
type TransactionService interface {
	// FindByHash  查询指定hash对应的交易
	FindByHash(txHash string) (Transaction, error)

	// FindByStartDate 查询最接近交易开始日期(>=startDate)的交易
	FindByStartDate(fromAddress, toAddress string, contract Contract, startDate time.Time) (Transaction, error)
}
