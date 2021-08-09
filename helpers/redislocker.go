package helpers

import (
	"time"

	"github.com/fighterlyt/redislock"
	"github.com/go-redsync/redsync/v4"
	"github.com/pkg/errors"
)

// EnsureRedisLock 一定会获取到分布式锁
func EnsureRedisLock(mutex redislock.Mutex) {
	for err := mutex.Lock(); errors.Is(err, redsync.ErrFailed); err = mutex.Lock() {
		if err == nil {
			return
		}

		time.Sleep(time.Millisecond)
	}
}
