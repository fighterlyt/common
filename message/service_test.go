package message

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	testService, err = NewService(db, testLogger)
	require.NoError(t, err, `NewService`)
}

func TestService_Create(t *testing.T) {
	TestNewService(t)

	require.NoError(t, testService.Add(bg, `a`, `b`))

	require.NoError(t, testService.Add(bg, `a`, `b`), `重复写入`)
}

func TestService_Get(t *testing.T) {
	TestNewService(t)
	require.NoError(t, testService.clearAll(), `清理`)

	key := `a`
	messages := []string{`b`, `c`}

	for _, message := range messages {
		require.NoError(t, testService.Add(bg, key, message))
	}

	var (
		values []string
	)

	values, err = testService.Get(key)
	require.NoError(t, err, `获取`)

	require.EqualValues(t, messages, values)
}

func TestService_Exist(t *testing.T) {
	TestNewService(t)
	require.NoError(t, testService.clearAll(), `清理`)

	key := `a`
	messages := []string{`c`, `b`}

	for _, message := range messages {
		require.NoError(t, testService.Add(bg, key, message))
	}

	var (
		exsits bool
	)

	nonExist := ``

	for _, message := range messages {
		exsits, err = testService.Exist(key, message)
		require.NoError(t, err)
		require.True(t, exsits)

		nonExist += message
	}

	exsits, err = testService.Exist(key, nonExist)
	require.NoError(t, err)
	require.False(t, exsits)
}
