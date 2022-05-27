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
	Hooks
}

// Hook 钩子信息
type Hook interface {
	Key() string
	// BeforeFreeze 冻结前
	BeforeFreeze(info *FreezeInfo)
	// AfterFreeze 冻结后
	AfterFreeze(info *FreezeInfo, err error)
	// BeforeUnfreeze 解冻前
	BeforeUnfreeze(info *FreezeInfo)
	// AfterUnfreeze 解冻后
	AfterUnfreeze(info *FreezeInfo, err error)
}

// FreezeInfo 冻结/解冻信息
type FreezeInfo struct {
	From   string
	To     string
	Amount decimal.Decimal
}

func NewFreezeInfo(from, to string, amount decimal.Decimal) *FreezeInfo {
	return &FreezeInfo{
		From:   from,
		To:     to,
		Amount: amount,
	}
}

// Hooks 钩子管理器
type Hooks interface {
	// Add 添加钩子
	Add(hook Hook) error
	// Remove 删除钩子
	Remove(key string)
	// EveryBeforeFreeze 冻结前
	EveryBeforeFreeze(info *FreezeInfo)
	// EveryAfterFreeze 冻结后
	EveryAfterFreeze(info *FreezeInfo, err error)
	// EveryBeforeUnfreeze 解冻前
	EveryBeforeUnfreeze(info *FreezeInfo)
	// EveryAfterUnfreeze 解冻后
	EveryAfterUnfreeze(info *FreezeInfo, err error)
}
