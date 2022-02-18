package durablemetrics

import (
	"context"
	"strings"

	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.com/nova_dubai/common/helpers"
)

type CounterVec struct {
	counterVec *prometheus.CounterVec
	logger     log.Logger
	name       string
	labelNames []string
}

func NewCounterVec(name, help string, labelNames []string, logger log.Logger) (*CounterVec, error) { // nolint:golint,dupl
	counterVec := &CounterVec{
		logger:     logger,
		name:       name,
		labelNames: labelNames,
		counterVec: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: argument.Namespace,
			Name:      name, Help: help,
		}, labelNames),
	}

	counterVecValues, err := loadMetricsVec(generateRedisKey(name))
	if err != nil {
		return nil, errors.Wrap(err, "初始化监控数据失败")
	}

	for i := range counterVecValues {
		counterVec.counterVec.WithLabelValues(counterVecValues[i].labelValues...).Add(counterVecValues[i].value)
	}

	return counterVec, nil
}

const (
	connector = "=" // 连接符
)

func (c CounterVec) WithLabelValuesInc(lvs ...string) {
	c.WithLabelValuesAdd(1, lvs...)
}

func (c CounterVec) WithLabelValuesAdd(value float64, lvs ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	c.counterVec.WithLabelValues(lvs...).Add(value)

	var keys []string
	for i := range lvs {
		keys = append(keys, strings.Join([]string{c.labelNames[i], lvs[i]}, connector))
	}

	helpers.IgnoreError(c.logger, "redis操作失败", func() error {
		key := strings.Join(keys, ";")

		return metricsRedisClient.HIncrByFloat(ctx, generateRedisKey(c.name), key, value).Err()
	})
}
