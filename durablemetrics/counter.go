package durablemetrics

import (
	"context"
	"github.com/go-redis/redis/v8"

	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.com/nova_dubai/common/helpers"
)

// Counter 计数器
type Counter struct {
	prometheus.Counter
	name   string
	logger log.Logger
}

func NewCounter(name, help string, logger log.Logger) (*Counter, error) {
	counter := &Counter{
		Counter: promauto.NewCounter(prometheus.CounterOpts{
			Name: name, Help: help,
		}),
		name:   name,
		logger: logger,
	}

	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	value, err := metricsRedisClient.Get(ctx, generateRedisKey(name)).Float64()
	if err != nil {
		if err != redis.Nil {
			return nil, errors.Wrap(err, "redis操作失败")
		}

		return counter, nil
	}

	counter.Counter.Add(value)

	return counter, nil
}

func (c Counter) Inc() {
	c.Add(1)
}

func (c Counter) Add(f float64) {
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	c.Counter.Add(f)

	helpers.IgnoreError(c.logger, "redis操作失败", func() error {
		return metricsRedisClient.IncrByFloat(ctx, generateRedisKey(c.name), f).Err()
	})
}
