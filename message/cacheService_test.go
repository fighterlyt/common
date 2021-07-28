package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/nova_dubai/cache"
)

var (
	manager          cache.Manager
	testCacheService Service
	result           []string
)

func TestCache(t *testing.T) {
	manager, err = cache.NewService(testLogger, `localhost:9736`, `dubaihell`, 6)
	require.NoError(t, err, `构建缓存服务`)
}

func TestNewCacheService(t *testing.T) {
	TestCache(t)

	testCacheService, err = NewCacheService(db, testLogger, manager)
	require.NoError(t, err, `NewCacheService`)
}

func TestCacheService_Get(t *testing.T) {
	TestNewCacheService(t)

	result, err = testCacheService.Get(`a`)
	require.NoError(t, err, `get`)
	t.Log(result)
}

func TestCacheService_AddAndGet(t *testing.T) {
	TestNewCacheService(t)

	result, err = testCacheService.Get(`a`)
	require.NoError(t, err, `get`)
	t.Log(result)

	require.NoError(t, testCacheService.Add(bg, `a`, `d`))

	result, err = testCacheService.Get(`a`)
	require.NoError(t, err, `get`)
	t.Log(result)
}

var (
	bg = context.Background()
)
