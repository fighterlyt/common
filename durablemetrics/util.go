package durablemetrics

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

const (
	// redis超时时间
	redisTimeout = 2 * time.Second
)

var (
	// metricsRedisClient
	metricsRedisClient *redis.Client
)

/*SerMetricRedisClient 设置监控的redis客户端
参数:
*	redisClient	*redis.Client	客户端链接
返回值:
*/
func SerMetricRedisClient(redisClient *redis.Client) {
	metricsRedisClient = redisClient
}

/*loadMetricsVec 加载监控历史数据
参数:
*	name     	string    	监控的name
返回值:
*	vecValues	[]vecValue	监控的labelValue及对应的值
*	err      	error     	错误
*/
func loadMetricsVec(name string) (vecValues []vecValue, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	result, err := metricsRedisClient.HGetAll(ctx, name).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, errors.Wrap(err, "redis操作失败")
	}

	for key, value := range result {
		var labelValue *vecValue
		labelValue, err = parse(key, value)

		if err != nil {
			return nil, errors.Wrap(err, "加载监控数据失败")
		}

		vecValues = append(vecValues, *labelValue)
	}

	return vecValues, nil
}

func parse(key, value string) (*vecValue, error) {
	var datas []string

	fields := strings.Split(key, ";")

	for i := range fields {
		data := strings.Split(fields[i], connector)
		if len(data) != 2 { // nolint:golint,gomnd
			continue
		}

		datas = append(datas, data[1])
	}

	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, errors.Wrap(err, "监控数据格式转换错误")
	}

	return &vecValue{labelValues: datas, value: f}, nil
}

type vecValue struct {
	labelValues []string
	value       float64
}
