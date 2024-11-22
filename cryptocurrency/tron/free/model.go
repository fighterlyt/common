package free

import (
	"github.com/fighterlyt/common/helpers"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// FreezeRecord trx 质押记录
type FreezeRecord struct {
	ID           int64           `gorm:"column:id;primaryKey;comment:id" json:"id"`
	From         string          `gorm:"column:from;type:varchar(64);comment:TRX的提供方"`
	To           string          `gorm:"column:to;type:varchar(64);comment:能量的受益方"`
	FreezeTxID   string          `gorm:"column:freeze_tx_id;type:varchar(128);comment:冻结交易ID"`
	UnFreezeTxID string          `gorm:"column:unfreeze_tx_id;type:varchar(128);comment:解冻交易ID"` //  解冻时，因为可能对应一笔记录，需要
	Amount       decimal.Decimal `gorm:"column:amount;type:decimal(20,6);comment:质押的TRX"`
	Time         helpers.Time    `gorm:"column:time;type:int(10);comment:质押时间"`
	UnFreezeTime helpers.Time    `gorm:"column:unfreeze_time;type:int(10);comment:解冻时间"`
}

func NewFreezeRecord(from, to, freezeTxID string, amount decimal.Decimal, time helpers.Time) *FreezeRecord {
	return &FreezeRecord{
		From:       from,
		To:         to,
		FreezeTxID: freezeTxID,
		Amount:     amount,
		Time:       time,
	}
}

func (FreezeRecord) TableName() string {
	return `trx_freeze_record`
}

// GetRecordsConditionFilter 用于过滤记录的条件
type GetRecordsConditionFilter struct {
	From     string                       `json:"from"`     // 发出账号
	To       string                       `json:"to"`       // 收益账号
	Amount   helpers.DecimalRangeArgument `json:"amount"`   // 金额范围
	Time     helpers.Range                `json:"time"`     // 时间范围
	Unfreeze bool                         `json:"unfreeze"` // 是否已经赎回
}

func (g GetRecordsConditionFilter) ForSQL() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if !helpers.IsStringEmpty(g.From) {
			db = db.Where(`from = ?`, g.From)
		}

		if !helpers.IsStringEmpty(g.To) {
			db = db.Where(`to = ?`, g.To)
		}

		db = db.Scopes(g.Amount.Scope(`amount`))
		db = db.Where(g.Time.Scope(`time`))

		if g.Unfreeze {
			db = db.Where(`unfreeze_tx_id = ?`, ``)
		} else {
			db = db.Where(`unfreeze_tx_id != ?`, ``)
		}

		return db
	}
}

// GetRecordFilter 获取条件
type GetRecordFilter struct {
	GetRecordsConditionFilter
	Start int `json:"start"`
	Limit int `json:"limit"`
}

// FailOperation 操作失败记录
type FailOperation struct {
	ID     int64           `gorm:"column:id;primaryKey;comment:id" json:"id"`
	From   string          `gorm:"column:from;type:varchar(64);comment:TRX的提供方"`
	To     string          `gorm:"column:to;type:varchar(64);comment:能量的受益方"`
	Error  string          `gorm:"column:error;type:varchar(2048);comment:错误信息"`
	Amount decimal.Decimal `gorm:"column:amount;type:decimal(20,6);comment:质押的TRX"`
	Time   helpers.Time    `gorm:"column:time;type:int(10);comment:操作时间"`
	Freeze bool            `gorm:"column:freeze;comment:是否是冻结"`
}

/*
NewFailOperation 新建操作失败记录
参数:
*	from          	string         	质押来源地址
*	to            	string         	质押收益地址
*	error         	string         	错误信息
*	amount        	decimal.Decimal	质押金额
*	time          	helpers.Time   	发生时间
*	freeze        	bool           	是否为质押
返回值:
*	*FailOperation	*FailOperation 	失败
*/
func NewFailOperation(from, to, error string, amount decimal.Decimal, time helpers.Time, freeze bool) *FailOperation {
	return &FailOperation{
		From:   from,
		To:     to,
		Error:  error,
		Amount: amount,
		Time:   time,
		Freeze: freeze,
	}
}

func (FailOperation) TableName() string {
	return `free_fail_operation`
}
