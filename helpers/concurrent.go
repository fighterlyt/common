package helpers

import (
	"fmt"
	stdLog "log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/fighterlyt/log"
	"gitlab.com/nova_dubai/common/telegram"
	"go.uber.org/zap"
)

var (
	telegramClient telegram.Telegram
)

func SetTeleGramClient(client telegram.Telegram) {
	telegramClient = client
}

/*EnsureGo 并发函数，确保在返回前已经开始执行
参数:
*	functions	...func()
返回值:
*/
func EnsureGo(logger log.Logger, functions ...func()) {
	wg := &sync.WaitGroup{}
	wg.Add(len(functions))

	for i := range functions {
		function := functions[i]
		go recoverRun(logger, wg, function)
	}

	wg.Wait()
}

func recoverRun(logger log.Logger, wg *sync.WaitGroup, function func()) {
	defer func() {
		if err := recover(); err != nil {
			if logger != nil {
				logger.Error("发生panic", zap.Any("错误信息", err), zap.ByteString("堆栈", debug.Stack()))
			} else {
				stdLog.Println("发生panic")
			}

			// 发送告警
			IgnoreError(logger, "发送告警", func() error {
				msgInfo := fmt.Sprintf("发生panic,发生时间[%s],错误信息[%v],堆栈信息[%s]",
					time.Now().UTC().Format("2006-01-02 15:04:05"), err, string(debug.Stack()))

				return telegramClient.SendMsg(msgInfo)
			})
		}
	}()
	wg.Done()
	function()
}
