package helpers

import (
	"github.com/fighterlyt/log"
	"go.uber.org/zap"
)

/*IgnoreError 忽略方法的错误 ,如果fun执行返回错误且logger!=nil,那么日志输出
参数:
*	logger	log.Logger      日志器
*	msg   	string          相关信息
*	fun   	func() error    方法
返回值:
*/
func IgnoreError(logger log.Logger, msg string, fun func() error) {
	err := fun()
	if err != nil && logger != nil {
		logger.Error(msg, zap.String(`错误`, err.Error()))
	}
}
