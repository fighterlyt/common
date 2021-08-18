package helpers

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func TestRedisDelInBatch(t *testing.T) {
	redisOption := &redis.Options{
		Addr:     "",
		DB:       0,
		Password: ``,
	}

	client := redis.NewClient(redisOption)

	count := 100
	batchSize := 10

	keys := make([]string, 0, count)

	for i := 0; i < count; i++ {
		key := fmt.Sprintf(`TestRedisDelInBatch_%d`, i+1)

		require.NoError(t, client.Set(bg, key, `a`, time.Minute).Err(), `写入数据`)

		keys = append(keys, key)
	}

	require.NoError(t, RedisDelInBatch(bg, client, keys, batchSize), `不应报错`)

	var (
		err error
	)

	keys, err = client.Keys(bg, `TestRedisDelInBatch_*`).Result()
	require.NoError(t, err, `KEYS 命令`)

	require.EqualValues(t, 0, len(keys), `删除后应该没有数据`)
}
