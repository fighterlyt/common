package free

import (
	"github.com/shopspring/decimal"
)

// Service 服务
type Service interface {
	// SetUp 设置质押来源和对应的私钥
	SetUp(from, privateKey string) error
	// Freeze 质押，to 收益地址 trxAmount  质押TRX 金额
	Freeze(to string, trxAmount decimal.Decimal) error
	// FreezeForTransfer 质押用于转账，会自动计算需要的TRX
	FreezeForTransfer(to string) error
	// UnFreeze 解冻/解除质押
	UnFreeze(to string) error
	// GetRecords 获取全部记录，filter 是记录,needAllCount 是否需要全部计数， totalCount 总数量,records 总记录 err 错误
	GetRecords(filter GetRecordFilter, needAllCount bool) (totalCount int64, records []FreezeRecord, err error)
}
