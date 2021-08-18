package parameters

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_Modify(t *testing.T) {
	userID := int64(3)
	key := `test:a:b`
	newValue := `2`

	parameter := NewParameter(key, `测试`, `1`, `测试1`, `positiveInteger`)
	require.NoError(t, testService.Create(parameter))

	require.NoError(t, testService.Modify(map[string]string{key: newValue}, userID), `修改`)

	newParameter, err := testService.GetParameters(key)
	require.NoError(t, err, `获取修改后的`)

	require.EqualValues(t, newValue, newParameter[key].Value, `值相同`)

	var (
		history []History
	)

	_, history, err = testService.GetHistory(key, 0, 0, 0, 10)

	require.NoError(t, err, `获取一条变更记录`)

	require.EqualValues(t, 1, len(history), `只有一条`)

	require.EqualValues(t, newValue, history[0].Value, `值相同`)
	require.EqualValues(t, userID, history[0].UserID, `操作人员相同`)
}

func TestService_AddParameters(t *testing.T) {
	key := `test:1`

	parameter := NewParameter(key, `测试`, `1`, `测试1`, `numeric`)
	require.NoError(t, testService.AddParameters(parameter))

	newParameter, err := testService.GetParameters(parameter.Key)
	require.NoError(t, err, `获取修改后的`)

	require.EqualValues(t, parameter, newParameter[parameter.Key], `值相同`)

	var (
		history []History
	)

	_, history, err = testService.GetHistory(parameter.Key, 0, 1, 10, 10)

	require.NoError(t, err, `获取一条变更记录`)

	require.EqualValues(t, 1, len(history), `只有一条`)

	require.EqualValues(t, parameter.Value, history[0].Value, `值相同`)
	require.EqualValues(t, -1, history[0].UserID, `操作人员相同`)
}
