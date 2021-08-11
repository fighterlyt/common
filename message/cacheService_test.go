package message

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"gitlab.com/nova_dubai/cache"
	"gitlab.com/nova_dubai/common/helpers"
	"go.uber.org/zap/zapcore"
)

var (
	manager          cache.Manager
	testCacheService Service
	result           []string
)

func TestCache(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     `localhost:9736`,
		Password: `dubaihell`,
		DB:       6,
	})

	redisClient.AddHook(helpers.NewRedisLogger(testLogger.Derive(`redis`).SetLevel(zapcore.DebugLevel).AddCallerSkip(1)))
	manager, err = cache.NewServiceByRedisClient(testLogger, redisClient)
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
	time.Sleep(time.Second)
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

func Test_cacheService_Delete(t *testing.T) {
	TestNewCacheService(t)

	require.NoError(t, testCacheService.Delete("a"))
}
