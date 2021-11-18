package parameters

import (
	"fmt"
	"time"

	"github.com/fighterlyt/log"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type parameterService struct {
	client      *redis.Client
	db          *gorm.DB
	logger      log.Logger
	redisExpire time.Duration
	model       *Parameter
}

func newParameterService(client *redis.Client, db *gorm.DB, logger log.Logger, redisExpire time.Duration) *parameterService {
	return &parameterService{
		client:      client,
		db:          db,
		logger:      logger,
		redisExpire: redisExpire,
		model:       &Parameter{},
	}
}

func (p parameterService) redisKey(key string) string {
	return fmt.Sprintf(`parameters_%s`, key)
}

func (p parameterService) GetParameters(keys ...string) (parameters map[string]*Parameter, err error) {
	parameters = make(map[string]*Parameter, len(keys))

	for _, key := range keys {
		var parameter *Parameter

		if parameter, err = p.get(key); err != nil {
			return nil, errors.Wrapf(err, `获取[%s]失败`, key)
		}

		parameters[key] = parameter
	}

	return parameters, nil
}

func (p parameterService) get(key string) (parameter *Parameter, err error) {
	if parameter, err = p.getFromRedis(p.redisKey(key)); err != nil {
		return nil, errors.Wrapf(err, `从REDIS 中获取[%s]`, key)
	}

	if parameter != nil {
		return parameter, nil
	}

	if parameter, err = p.getFromMysql(key); err != nil {
		return nil, errors.Wrapf(err, `从MYSQL 获取[%s]`, key)
	}

	if parameter == nil {
		return nil, nil
	}

	if err = p.updateRedis(parameter); err != nil {
		p.logger.Warn(`更新REDIS失败`, zap.String(`错误`, err.Error()))
	}

	return parameter, nil
}

func (p parameterService) updateRedis(parameter *Parameter) error {
	redisKey := p.redisKey(parameter.Key)
	if err := p.client.Set(bg, redisKey, parameter, p.redisExpire).Err(); err != nil {
		return errors.Wrapf(err, `REDIS SET %s 值 %d`, redisKey, p.redisExpire)
	}

	return nil
}

func (p parameterService) getFromRedis(key string) (parameter *Parameter, err error) {
	parameter = &Parameter{}
	if err = p.client.Get(bg, key).Scan(parameter); err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, errors.Wrapf(err, `REDIS GET %s`, key)
	}

	return parameter, nil
}

func (p parameterService) getFromMysql(key string) (parameter *Parameter, err error) {
	parameter = &Parameter{}
	if err = p.db.Model(p.model).Where(`elemKey = ?`, key).First(parameter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, errors.Wrapf(err, `key=%s`, key)
	}

	return parameter, nil
}

func (p parameterService) Modify(key, value string) error {
	parameter, err := p.get(key)
	if err != nil {
		return errors.Wrapf(err, `获取[%s]验证规则`, key)
	}

	if parameter == nil {
		return fmt.Errorf(`参数[%s]不存在`, key)
	}

	parameter.Value = value
	if err = parameter.Validate(); err != nil {
		return fmt.Errorf(`错误信息: [%v],值[%s]不满足要求[%s]`, err, parameter.Value, parameter.Description)
	}

	if err := p.deleteFromRedis(key); err != nil {
		return errors.Wrap(err, `删除缓存`)
	}

	if err := p.updateMYSQL(key, parameter.Value); err != nil {
		return errors.Wrap(err, `更新MYSQL`)
	}

	return nil
}

func (p parameterService) deleteFromRedis(key string) error {
	redisKey := p.redisKey(key)

	if err := p.client.Del(bg, redisKey).Err(); err != nil {
		return errors.Wrapf(err, `REDIS DEL %s`, redisKey)
	}

	return nil
}

func (p parameterService) updateMYSQL(key, value string) error {
	if err := p.db.Model(p.model).Where(`elemKey = ?`, key).Updates(map[string]interface{}{
		`value`:      value,
		`lock`:       false,
		`updateTime`: time.Now().Unix(),
	}).Error; err != nil {
		return errors.Wrapf(err, `更新MYSQL key[%s],value[%s]`, key, value)
	}

	return nil
}

func (p parameterService) Save(parameter *Parameter) error {
	if err := parameter.Validate(); err != nil {
		return errors.Wrapf(err, `校验失败`)
	}

	return p.db.Model(p.model).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoNothing: true,
	}).Create(parameter).Error
}
