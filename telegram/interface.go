package telegram

import (
	"fmt"
	"strings"
	"time"

	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"gopkg.in/tucnak/telebot.v2"
)

type Telegram interface {
	SendMsg(msg string) error
	SendMarkdown(msg string) error
}

// 1886019351:AAGOCfqMyPC-xIbqpN0WsEkv0fkERKdnyE8
type telegram struct {
	serviceName string
	token       string
	bot         *telebot.Bot
	logger      log.Logger
	chatID      int64
	group       telebot.ChatID
}

func NewTelegram(serviceName, token string, logger log.Logger, chatID int64) (*telegram, error) {
	b, err := telebot.NewBot(telebot.Settings{
		URL:         "",
		Token:       token,
		Updates:     0,
		Poller:      &telebot.LongPoller{Timeout: 10 * time.Second},
		Synchronous: false,
		Verbose:     false,
		ParseMode:   telebot.ModeMarkdownV2,
		Reporter:    nil,
		Client:      nil,
	})
	if err != nil {
		return nil, errors.Wrap(err, `构建telebot`)
	}

	go b.Start()

	return &telegram{
		serviceName: serviceName,
		token:       token,
		bot:         b,
		logger:      logger,
		group:       telebot.ChatID(chatID),
		chatID:      chatID,
	}, nil
}

var maxLen = 4096

func (t telegram) SendMsg(msg string) error {
	msg = fmt.Sprintf("服务[%s]", t.serviceName) + msg

	for {
		sendMsg := t.getMessage(msg)

		if _, err := t.bot.Send(t.group, escape(sendMsg), telebot.ModeDefault); err != nil {
			return errors.Wrap(err, "发送消息失败")
		}

		if len(sendMsg) < maxLen {
			return nil
		}

		msg = msg[len(sendMsg)+1:]
	}
}

func (t telegram) getMessage(msg string) string {
	var msgLen = len(msg)
	if msgLen > maxLen {
		msgLen = maxLen
	}

	return msg[:msgLen]
}

func (t telegram) SendMarkdown(msg string) error {
	_, err := t.bot.Send(t.group, msg, telebot.ModeMarkdownV2)
	return err
}

var (
	needEscape = string([]byte{'_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'})
)

func escape(msg string) string {
	for i := range needEscape {
		msg = strings.ReplaceAll(msg, needEscape[i:i+1], `\`+needEscape[i:i+1])
	}

	return msg
}

// 空的结构
type noneTelegram struct {
}

func NewNoneTelegram() Telegram {
	return &noneTelegram{}
}

func (n noneTelegram) SendMsg(msg string) error {
	return nil
}

func (n noneTelegram) SendMarkdown(msg string) error {
	return nil
}
