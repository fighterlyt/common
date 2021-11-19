package parameters

import (
	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type historyService struct {
	db     *gorm.DB
	logger log.Logger
	model  *History
}

func newHistoryService(db *gorm.DB, logger log.Logger) *historyService {
	return &historyService{db: db, logger: logger, model: &History{}}
}

func (h historyService) Save(key, value string, userID int64) error {
	history := NewHistory(key, value, userID)

	if err := h.db.Create(history).Error; err != nil {
		return errors.Wrap(err, `保存数据错误`)
	}

	return nil
}

func (h historyService) Get(key string, startTime, endTime int64, start, limit int) (allCount int64, histories []History, err error) {
	h.logger.Debug(`获取数据`, zap.String(`elemKey`, key), zap.Ints(`开始/数量`, []int{start, limit}))

	query := h.db.Model(h.model).Debug()

	if key != "" {
		query = query.Where(`elemKey= ?`, key)
	}

	if startTime != 0 {
		query = query.Where(`updateTime >= ? `, startTime)
	}

	if endTime != 0 {
		query = query.Where(`updateTime <= ? `, endTime)
	}

	if err = query.Limit(limit).
		Offset(start).
		Order(`updateTime desc`).Find(&histories).Error; err != nil {
		return 0, nil, errors.Wrapf(err, `获取数据,key=[%s],start=[%d],limit=[%d]`, key, start, limit)
	}

	if len(histories) < limit {
		h.logger.Debug(`数量已知，不需要进行SQL操作`)
		return int64(start + len(histories)), histories, nil
	}

	if err = query.Count(&allCount).Error; err != nil {
		return 0, nil, errors.Wrap(err, `统计数量`)
	}

	return allCount, histories, nil
}
