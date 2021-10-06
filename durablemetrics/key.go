package durablemetrics

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

const (
	prefix = "metric_"
)

func generateRedisKey(originalKey string) string {
	return prefix + originalKey
}

var errNotFound = errors.New("not found from redis")

func getValueFromRedis(name string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	value, err := metricsRedisClient.Get(ctx, generateRedisKey(name)).Float64()
	if err != nil {
		if err != redis.Nil {
			return 0, errors.Wrap(err, "redis操作失败")
		}

		return 0, errNotFound
	}

	return value, nil
}
