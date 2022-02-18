package durablemetrics

import (
	"os"
	"testing"

	"github.com/fighterlyt/log"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/nova_dubai/common/helpers"
)

var (
	counter    *Counter
	gauge      *Gauge
	counterVec *CounterVec
	gaugeVec   *GaugeVec
)

func TestMain(m *testing.M) {
	metricsRedisClient = redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379",
		Password: "123",
		DB:       3})

	logger, err := log.NewEasyLogger(true, false, "", "test_metrics")
	if err != nil {
		panic(err)
	}

	engine := gin.Default()
	engine.GET(`/metrics`, func(ctx *gin.Context) {
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})

	helpers.EnsureGo(logger, func() {
		_ = engine.Run(":1235")
	})

	counter, err = NewCounter("test_cunter", "test_help", logger)
	if err != nil {
		panic(err)
	}

	gauge, err = NewGauge("test_gauge", "测试gauge", logger)
	if err != nil {
		panic(err)
	}

	counterVec, err = NewCounterVec("test_counter_vec", "测试counter_vec", []string{`url`, `method`, `sendOK`, `statusCode`, `local`}, logger)
	if err != nil {
		panic(err)
	}

	gaugeVec, err = NewGaugeVec("test_gauge_vec", "测试gauge_vec", []string{"currency"}, logger)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}
