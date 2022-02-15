package helpers

import (
	"context"

	"github.com/fighterlyt/log"
)

func PutLogger(ctx context.Context, logger log.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

/*GetLogger 从上下文中获取日志器，如果其中的日志器==nil,那么返回默认日志器
参数:
*	ctx          	context.Context	上下文
*	defaultLogger	log.Logger     	默认日志器
返回值:
*	log.Logger   	log.Logger     	日志器
*/
func GetLogger(ctx context.Context, defaultLogger log.Logger) log.Logger {
	data := ctx.Value(loggerKey)
	if data == nil {
		return defaultLogger
	}

	if data.(log.Logger) == nil {
		return defaultLogger
	}

	return data.(log.Logger)
}

type loggerKeyType string

var (
	loggerKey loggerKeyType = `logger`
)
