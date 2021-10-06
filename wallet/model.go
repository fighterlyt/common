package wallet

import (
	"github.com/shopspring/decimal"
)

// UserBalance 用户钱包余额
type UserBalance struct {
	ID       int64           `gorm:"primary_key" json:"id"`
	UserID   int64           `gorm:"column:userID;uniqueIndex:user_protocol;comment:'所属用户ID'" json:"userID"`
	Protocol string          `gorm:"column:protocol;uniqueIndex:user_protocol;type:varchar(50);comment:'协议类型'" json:"protocol"`
	Symbol   string          `gorm:"column:symbol;uniqueIndex:user_protocol;type:varchar(50);comment:'币种'" json:"symbol"`
	Balance  decimal.Decimal `gorm:"column:balance;type:decimal(20,8);comment:'账户余额'" json:"balance"`
}

func (w *UserBalance) TableName() string {
	return "user_balance"
}
