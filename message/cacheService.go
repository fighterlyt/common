package message

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"gitlab.com/nova_dubai/cache"
	"gorm.io/gorm"
)

type cacheService struct {
	service *service
	client  cache.Client
}

func NewCacheService(db *gorm.DB, logger log.Logger, manager cache.Manager) (result *cacheService, err error) {
	var (
		service *service
		typ     cache.Type
		load    cache.Load
		client  cache.Client
	)

	if service, err = NewService(db, logger); err != nil {
		return nil, errors.Wrap(err, `构建service`)
	}

	load = func(ctx context.Context, key interface{}) (interface{}, error) {
		str, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf(`不支持key类型为[%s]`, reflect.TypeOf(key).Kind().String())
		}

		str = strings.TrimPrefix(str, Record{}.TableName()+cache.Delimiter)

		var (
			result []string
			getErr error
		)

		if result, getErr = service.Get(str); getErr != nil {
			return nil, errors.Wrap(getErr, `db记载`)
		}

		return messages(result), nil
	}

	if typ, err = cache.NewTypeTmpl(Record{}.TableName(), load, func() interface{} {
		return &messages{}
	}); err != nil {
		return nil, errors.Wrap(err, `构建类型`)
	}

	if client, err = manager.Register(typ, time.Minute, cache.OnlyRedis); err != nil {
		return nil, errors.Wrap(err, `注册到缓存服务`)
	}

	return &cacheService{
		service: service,
		client:  client,
	}, nil
}

/*Get 获取指定分类的消息，基于缓存
参数:
*	key    	string  	分类
返回值:
*	message	[]string	消息
*	err    	error   	错误
*/
func (c cacheService) Get(key string) (message []string, err error) {
	var (
		result interface{}
	)

	if result, err = c.client.Get(key); err != nil {
		return nil, errors.Wrap(err, `从缓存获取`)
	}

	if _, ok := result.(*messages); ok {
		return *result.(*messages), nil
	}

	if _, ok := result.([]string); ok {
		return result.([]string), nil
	}

	return nil, fmt.Errorf(`数据类型为[%s],%v`, reflect.TypeOf(result).Kind().String(), result)
}

func (c cacheService) Exist(key, message string) (exists bool, err error) {
	return c.service.Exist(key, message)
}

func (c cacheService) Add(ctx context.Context, key, message string) error {
	if err := c.service.Add(ctx, key, message); err != nil {
		return errors.Wrap(err, `数据库新增失败`)
	}

	_ = c.client.Invalidate(key)

	return nil
}

type messages []string

func (m *messages) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m messages) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}
