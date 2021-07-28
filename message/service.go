package message

import (
	"context"
	"fmt"

	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type service struct {
	db     *gorm.DB
	logger log.Logger
}

/*NewService 新建普通服务
参数:
*	db    	*gorm.DB  	数据
*	logger	log.Logger	日志器
返回值:
*	result	*service  	服务
*	err   	error     	错误
*/
func NewService(db *gorm.DB, logger log.Logger) (result *service, err error) {
	result = &service{
		db:     db.Model(&Record{}),
		logger: logger,
	}

	if err = result.start(); err != nil {
		return nil, errors.Wrap(err, `启动错误`)
	}

	return result, nil
}

func (s service) start() error {
	if err := s.db.AutoMigrate(&Record{}); err != nil {
		return errors.Wrap(err, `数据迁移失败`)
	}

	return nil
}

/*Get 获取一类信息
参数:
*	key    	string  	参数1
返回值:
*	message	[]string	返回值1
*	err    	error   	返回值2
*/
func (s service) Get(key string) (message []string, err error) {
	if err = s.db.WithContext(context.Background()).Where(`elemKey = ?`, key).Pluck(`value`, &message).Error; err != nil {
		return nil, errors.Wrap(err, `数据库操作失败`)
	}

	return message, nil
}

func (s service) Exist(key, message string) (exists bool, err error) {
	var (
		count = new(int64)
	)

	if err = s.db.Session(&gorm.Session{}).Where(`elemKey = ? and value = ?`, key, message).Count(count).Error; err != nil {
		return false, errors.Wrap(err, `数据库操作`)
	}

	return *count != 0, nil
}

func (s service) Add(ctx context.Context, key, message string) error {
	record := NewRecord(key, message)

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: `elemKey`}, {Name: `value`}},
		DoNothing: true,
	}).Create(&record).Error
}

func (s service) clearAll() error {
	if err := s.db.Exec(fmt.Sprintf(`DELETE FROM %s`, Record{}.TableName())).Error; err != nil {
		return errors.Wrapf(err, `清理失败`)
	}

	return nil
}
