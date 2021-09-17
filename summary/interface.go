package summary

import (
	"github.com/shopspring/decimal"
)

// Client 客户端抽象
type Client interface {
	Model() Summary
	Summarize(ownerID int64, amount decimal.Decimal) error
	SummarizeDay(date int64, ownerID int64, amount decimal.Decimal) error
	Key() string
	GetSummary(ownerIDs []int64, from, to int64) (records []Summary, err error)
}

// Summary 记录抽象
type Summary interface {
	GetID() int64
	GetSlot() Slot
	GetOwnerID() int64
	GetValue() decimal.Decimal
	GetSlotValue() int64
	TableName() string
}
