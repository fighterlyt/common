package parameters

import (
	"fmt"
	"strconv"
	"strings"

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

func (s *service) GetInts(key, delimiter string) (value []int64, err error) {
	var (
		result string
		temp   int64
	)

	if result, err = s.GetString(key); err != nil {
		return value, err
	}

	fields := strings.Split(result, delimiter)

	for _, field := range fields {
		if temp, err = strconv.ParseInt(field, 10, 64); err != nil {
			return nil, err
		}
		value = append(value, temp)
	}

	return value, nil
}

func (s *service) GetInt(key string) (value int64, err error) {
	var (
		result string
	)

	if result, err = s.GetString(key); err != nil {
		return value, err
	}

	return strconv.ParseInt(result, 10, 64)
}
