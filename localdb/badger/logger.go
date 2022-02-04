package badger

import (
	"fmt"

	"github.com/fighterlyt/log"
	"go.uber.org/zap"
)

// 日志器相关处理

type logger struct {
	logger log.Logger
}

func newLogger(originLogger log.Logger) *logger {
	return &logger{logger: originLogger}
}

/*Errorf 实现github.com/dgraph-io/badger/v3.Logger接口
参数:
*	s	string        	格式化字符
*	i	...interface{}	值
返回值:
*/
func (l logger) Errorf(s string, i ...interface{}) {
	l.logger.Error(`Error`, zap.String(`内容`, fmt.Sprintf(s, i...)))
}

/*Warningf 实现github.com/dgraph-io/badger/v3.Logger接口
参数:
*	s	string        	格式化字符
*	i	...interface{}	值
返回值:
*/
func (l logger) Warningf(s string, i ...interface{}) {
	l.logger.Warn(`Warning`, zap.String(`内容`, fmt.Sprintf(s, i...)))
}

/*Infof 实现github.com/dgraph-io/badger/v3.Logger接口
参数:
*	s	string        	格式化字符
*	i	...interface{}	值
返回值:
*/
func (l logger) Infof(s string, i ...interface{}) {
	l.logger.Info(``, zap.String(`内容`, fmt.Sprintf(s, i...)))
}

/*Debugf 实现github.com/dgraph-io/badger/v3.Logger接口
参数:
*	s	string        	格式化字符
*	i	...interface{}	值
返回值:
*/
func (l logger) Debugf(s string, i ...interface{}) {
	l.logger.Debug(``, zap.String(`内容`, fmt.Sprintf(s, i...)))
}
