package parameters

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func (s *service) GetString(key string) (wallet string, err error) {
	var (
		result map[string]*Parameter
	)

	if result, err = s.GetParameters(key); err != nil {
		return ``, errors.Wrapf(err, `获取业务参数[%s]`, key)
	}

	if result[key] == nil {
		return "", fmt.Errorf("key[%s]不存在", key)
	}

	return result[key].Value, nil
}

func (s *service) GetDecimal(key string) (value decimal.Decimal, err error) {
	var (
		result string
	)

	if result, err = s.GetString(key); err != nil {
		return value, err
	}

	return decimal.NewFromString(result)
}
