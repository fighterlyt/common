package alert

import (
	"github.com/fighterlyt/log"
	"gitlab.com/nova_dubai/common/telegram"
	"go.uber.org/zap"
)

type Service interface {
	SendText(msg string)
	SendMarkDown(md string)
}

type telegramService struct {
	service telegram.Telegram
	logger  log.Logger
}

/*NewTelegramService 新建基于telegram的服务
参数:
*	service         	telegram.Telegram	telegram服务
*	logger          	log.Logger       	日志器
返回值:
*	Service	            Service 	        服务
*/
func NewTelegramService(service telegram.Telegram, logger log.Logger) Service {
	return &telegramService{service: service, logger: logger}
}

func (t telegramService) SendText(msg string) {
	err := t.service.SendMsg(msg)
	if err != nil {
		t.logger.Error(`发送文本失败`, zap.Error(err), zap.String(`内容`, msg))
	}
}

// SendMarkDown 注意语法略有不同
func (t telegramService) SendMarkDown(md string) {
	err := t.service.SendMarkdown(md)
	if err != nil {
		t.logger.Error(`发送markdown失败`, zap.Error(err), zap.String(`内容`, md))
	}
}
