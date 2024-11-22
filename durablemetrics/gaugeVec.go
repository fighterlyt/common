package durablemetrics

import (
	"context"
	"strings"

	"github.com/fighterlyt/common/helpers"
	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type GaugeVec struct {
	gaugeVec   *prometheus.GaugeVec
	logger     log.Logger
	name       string
	labelNames []string
}

func NewGaugeVec(name, help string, labelNames []string, logger log.Logger) (*GaugeVec, error) { // nolint:golint,dupl
	gaugeVec := &GaugeVec{
		logger:     logger,
		name:       name,
		labelNames: labelNames,
		gaugeVec: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: argument.Namespace,
			Name:      name,
			Help:      help}, labelNames),
	}

	gaugeVecValues, err := loadMetricsVec(generateRedisKey(name))
	if err != nil {
		return nil, errors.Wrap(err, "初始化监控数据失败")
	}

	for i := range gaugeVecValues {
		gaugeVec.gaugeVec.WithLabelValues(gaugeVecValues[i].labelValues...).Add(gaugeVecValues[i].value)
	}

	return gaugeVec, nil
}

func (g GaugeVec) WithLabelValuesAdd(f float64, lvs ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	g.gaugeVec.WithLabelValues(lvs...).Add(f)

	keys := g.lvsToKeys(lvs...)

	helpers.IgnoreError(g.logger, "redis操作失败", func() error {
		return metricsRedisClient.HIncrByFloat(ctx, generateRedisKey(g.name), strings.Join(keys, ";"), f).Err()
	})
}

func (g GaugeVec) lvsToKeys(lvs ...string) []string {
	var keys []string

	for i := range lvs {
		keys = append(keys, strings.Join([]string{g.labelNames[i], lvs[i]}, connector))
	}

	return keys
}

func (g GaugeVec) WithLabelValuesSet(f float64, lvs ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	g.gaugeVec.WithLabelValues(lvs...).Set(f)

	keys := g.lvsToKeys(lvs...)

	helpers.IgnoreError(g.logger, "redis操作失败", func() error {
		return metricsRedisClient.HSet(ctx, generateRedisKey(g.name), strings.Join(keys, ";"), f).Err()
	})
}
