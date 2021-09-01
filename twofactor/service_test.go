package twofactor

import (
	"testing"

	"github.com/fighterlyt/log"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

var (
	testService Service
)

func TestNewService(t *testing.T) {
	audit := NewAuditBySingleAmount(func() (decimal.Decimal, error) {
		return decimal.New(100, 0), nil
	})

	logger, err := log.NewEasyLogger(true, false, ``, `动态验证`)
	require.NoError(t, err)

	TestNewConfig(t)
	testService = NewService(c, audit, newMockNotify(logger))

	need, err := testService.Process(`1`, 1, -1, ``, ``, `测试`, decimal.New(100, 0))
	require.NoError(t, err)
	require.True(t, need)

	need, err = testService.Process(`1`, 1, -1, ``, ``, `测试`, decimal.New(90, 0))
	require.NoError(t, err)
	require.False(t, need)
}
