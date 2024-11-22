package free

import (
	"github.com/fighterlyt/common/helpers"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var (
	modelRecord = &FreezeRecord{}
	modelFail   = &FailOperation{}
)

/*
CreateRecord 创建记录
参数:
*	to    	string         	质押收益地址
*	txID  	string         	质押交易Hash
*	amount	decimal.Decimal	质押TRX金额
返回值:
*	record	*FreezeRecord  	记录
*	err   	error          	错误
*/
func (s service) CreateRecord(to, txID string, amount decimal.Decimal) (record *FreezeRecord, err error) {
	record = NewFreezeRecord(s.from, to, txID, amount, helpers.Now())

	db := s.db.Model(modelRecord)

	if err = db.Create(record).Error; err != nil {
		return nil, errors.Wrap(err, `保存`)
	}

	return record, nil
}

/*
UpdateUnfreezeInfo 更新解冻信息
参数:
*	txID       	    string      	解冻交易ID
*	unfreeTime 	    helpers.Time	解冻时间
*	freezeTimeMax	helpers.Time	最大冻结时间
返回值:
*	error      	    error       	错误
重点:

	由于解冻是批量解冻，所有符合条件(冻结72小时)的都被解冻
*/
func (s service) UpdateUnfreezeInfo(txID string, unfreeTime, freezeTimeMax helpers.Time) error {
	db := s.db.Model(modelRecord)

	return db.Where(`unfreeze_tx_id = ? and freeze_time <= ?`, ``, freezeTimeMax).Updates(map[string]interface{}{
		`unfreeze_tx_id`: txID,
		`unfreeze_time`:  unfreeTime,
	}).Error
}

func (s service) findRecords(filter GetRecordFilter, needAllCount bool) (totalCount int64, records []FreezeRecord, err error) {
	db := s.db.Model(modelRecord)

	db = filter.ForSQL()(db)

	if err = db.Offset(filter.Start).Limit(filter.Limit).Find(&records).Error; err != nil {
		return 0, nil, errors.Wrap(err, `查询`)
	}

	if needAllCount {
		if err = db.Count(&totalCount).Error; err != nil {
			return 0, nil, errors.Wrap(err, `计数`)
		}
	}

	return totalCount, records, nil
}

/*
CreateFailRecord 创建操作鼠标记录
参数:
*	to    	string         	收益方
*	errMsg	string         	错误信息
*	amount	decimal.Decimal	金额
*	freeze	bool           	是否冻结
返回值:
*	record	*FailOperation 	记录
*	err   	error          	错误
*/
func (s service) CreateFailRecord(to, errMsg string, amount decimal.Decimal, freeze bool) (record *FailOperation, err error) {
	record = NewFailOperation(s.from, to, errMsg, amount, helpers.Now(), freeze)

	db := s.db.Model(modelFail)

	if err = db.Create(record).Error; err != nil {
		return nil, errors.Wrap(err, `保存`)
	}

	return record, nil
}
