package summaryextend

import (
	"github.com/shopspring/decimal"
)

// Client 客户端抽象
type Client interface {
	Model() Summary
	Summarize(ownerID string, amount decimal.Decimal, other ...decimal.Decimal) error
	SummarizeNotAddTimes(ownerID string, amount decimal.Decimal, extend ...decimal.Decimal) error
	SummarizeDay(date int64, ownerID string, amount decimal.Decimal, other ...decimal.Decimal) error
	RevertSummarizeDay(date int64, ownerID string, amount decimal.Decimal, other ...decimal.Decimal) error

	Key() string
	GetSummary(ownerIDs []string, from, to int64) (records []Summary, err error)
	GetSummaryByLike(like string, from, to int64) (records []Summary, err error)

	// GetSummarySummary 获取汇总的汇总
	GetSummarySummary(ownerIDs []string, from, to int64) (record Summary, err error)
	GetSummaryExclude(excludeOwnerID []string, from, to int64, selects ...string) (records []Summary, err error)

	SummarizeDayFirstUpdate(date int64, ownerID string, amount decimal.Decimal, other ...decimal.Decimal) error

	SummarizeFirstUpdate(ownerID string, amount decimal.Decimal, extendValue ...decimal.Decimal) error

	RevertSummarizeFirstUpdate(ownerID string, amount decimal.Decimal, extendValue ...decimal.Decimal) error
}

// Summary 记录抽象
type Summary interface {
	GetID() int64
	GetSlot() Slot
	GetOwnerID() string
	GetValue() decimal.Decimal
	GetExtendValue() []decimal.Decimal
	GetSlotValue() string
	TableName() string
	GetTimes() int64
	SetValue(value decimal.Decimal)
}
