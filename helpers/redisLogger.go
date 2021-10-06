package helpers

import (
	"context"

	"github.com/fighterlyt/log"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisLogger  实现了redis.Hook
type RedisLogger struct {
	logger log.Logger
}

func NewRedisLogger(logger log.Logger) *RedisLogger {
	return &RedisLogger{logger: logger}
}

func (r RedisLogger) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	r.logger.Debug(`执行命令`, zap.String(`命令`, cmd.String()), zap.Any(`db`, ctx.Value(`db`)))
	return ctx, nil
}

func (r RedisLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	r.logger.Debug(`执行命令完成`, zap.String(`命令`, cmd.String()))
	return nil
}

func (r RedisLogger) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	r.logger.Debug(`执行pipeline命令`)
	return ctx, nil
}

func (r RedisLogger) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	r.logger.Debug(`执行pipeline命令完成`)
	return nil
}
