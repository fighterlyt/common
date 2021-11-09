package parameters

import (
	"context"
	"gitlab.com/nova_dubai/common/twofactor"

	"gitlab.com/nova_dubai/common/model"
)

var (
	bg = context.Background()
)

// Service 服务
type Service interface {
	GetParameters(keys ...string) (parameters map[string]*Parameter, err error)
	AddParameters(parameters ...*Parameter) error
	Modify(keyValue map[string]string, userID int64) error
	GetHistory(key string, startTime, endTime int64, start, limit int) (allCount int64, histories []History, err error)
	GetString(key string) (value string, err error)
	SetTwoFactorAuth(needTwoFactorKeys []string, auth twofactor.Auth)
	model.Module
}

// ParameterService 参数服务
type ParameterService interface {
	Save(parameter *Parameter) error
	GetParameters(keys ...string) (parameters map[string]*Parameter, err error)
	Modify(key, value string) error
}

// HistoryService 历史服务
type HistoryService interface {
	Save(key, value string, userID int64) error
	Get(key string, startTime, endTime int64, start, limit int) (count int64, histories []History, err error)
}
