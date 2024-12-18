package parameters

import (
	"context"

	"github.com/fighterlyt/common/twofactor"
	"github.com/shopspring/decimal"

	"github.com/fighterlyt/common/model"
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
	HelperService
	SetTwoFactorAuth(needTwoFactorKeys []string, auth twofactor.Auth)
	SetValidate(validate ParameterValidate)
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

type HelperService interface {
	GetInts(key, delimiter string) (value []int64, err error)
	GetString(key string) (value string, err error)
	GetDecimal(key string) (value decimal.Decimal, err error)
	GetInt(key string) (value int64, err error)
}
