package summary

import (
	"github.com/shopspring/decimal"
)

// Client 客户端抽象
type Client interface {
	Model() Summary
	Summarize(ownerID string, amount decimal.Decimal) error
	SummarizeDay(date int64, ownerID string, amount decimal.Decimal) error
	Key() string
	GetSummary(ownerIDs []string, from, to int64) (records []Summary, err error)
}

// Summary 记录抽象
type Summary interface {
	GetID() int64
	GetSlot() Slot
	GetOwnerID() string
	GetValue() decimal.Decimal
	GetSlotValue() string
	TableName() string
	GetTimes() int64
}
