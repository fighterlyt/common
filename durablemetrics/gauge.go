package durablemetrics

import (
	"context"

	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.com/nova_dubai/common/helpers"
)

type Gauge struct {
	prometheus.Gauge
	name   string
	logger log.Logger
}

func NewGauge(name, help string, logger log.Logger) (*Gauge, error) {
	gauge := &Gauge{Gauge: promauto.NewGauge(prometheus.GaugeOpts{
		Name:      name,
		Help:      help,
		Namespace: argument.Namespace,
	}), name: name, logger: logger}

	value, err := getValueFromRedis(name)
	if err != nil {
		if !errors.Is(err, errNotFound) {
			return nil, errors.Wrap(err, "redis操作失败")
		}

		return gauge, nil
	}

	gauge.Gauge.Set(value)

	return gauge, nil
}

func (g Gauge) Set(f float64) {
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	g.Gauge.Set(f)

	helpers.IgnoreError(g.logger, "redis操作失败", func() error {
		return metricsRedisClient.Set(ctx, generateRedisKey(g.name), f, -1).Err()
	})
}

func (g Gauge) Inc() {
	g.Add(1)
}

func (g Gauge) Add(f float64) {
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	g.Gauge.Add(f)

	helpers.IgnoreError(g.logger, "redis操作失败", func() error {
		return metricsRedisClient.IncrByFloat(ctx, generateRedisKey(g.name), f).Err()
	})
}

func (g Gauge) Sub(f float64) {
	g.Add(-f)
}
