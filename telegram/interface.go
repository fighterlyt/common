package telegram

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/tucnak/telebot.v2"
)

const (
	initCapacity = 100
)

type Handler func(message *telebot.Message) string
type Telegram interface {
	SendMsg(msg string) error
	SendMarkdown(msg string) error
	Handle(string, Handler) error
	Start()
	SendFile(reader io.Reader, fileName string) error
	SendFileFromURL(url, fileName string) error
}

type telegram struct {
	serviceName string
	token       string
	bot         *telebot.Bot
	logger      log.Logger
	chatID      int64
	group       telebot.ChatID
	handlers    map[string]struct{}
	lock        *sync.Mutex
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

	b.Handle(telebot.OnText, func(tb *telebot.Message) {
		logger.Info(`收到文本命令`, zap.Any(`命令`, tb))
		_, _ = b.Reply(tb, `收到`)
	})

	b.Handle(`/hello`, func(m *telebot.Message) {
		_, _ = b.Send(m.Sender, `hello`)
	})

	return &telegram{
		serviceName: serviceName,
		token:       token,
		bot:         b,
		logger:      logger,
		group:       telebot.ChatID(chatID),
		chatID:      chatID,
		handlers:    make(map[string]struct{}, initCapacity),
		lock:        &sync.Mutex{},
	}, nil
}

var maxLen = 4096

func (t telegram) Start() {
	go t.bot.Start()
}
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

func (t telegram) SendFile(reader io.Reader, fileName string) error {
	_, err := t.bot.Send(t.group, &telebot.Document{
		File:      telebot.FromReader(reader),
		Thumbnail: nil,
		Caption:   "",
		MIME:      "",
		FileName:  fileName,
	})

	return err
}

func (t telegram) SendFileFromURL(fileURL, fileName string) error {
	resp, err := http.Get(fileURL)
	if err != nil {
		return errors.Wrap(err, `下载文件`)
	}

	defer resp.Body.Close()

	return t.SendFile(resp.Body, fileName)
}

func (t telegram) getMessage(msg string) string {
	var msgLen = len(msg)
	if msgLen > maxLen {
		msgLen = maxLen
	}

	return msg[:msgLen]
}

func (t telegram) SendMarkdown(msg string) error {
	msg = fmt.Sprintf("服务[%s]", t.serviceName) + msg

	_, err := t.bot.Send(t.group, msg, telebot.ModeMarkdown)

	return err
}

func (t *telegram) Handle(endPoint string, handler Handler) error {
	if endPoint == `` || endPoint[0:1] != `/` {
		return fmt.Errorf(`endPoint 不能为空且必须以/开头`)
	}

	if handler == nil {
		return fmt.Errorf(`handler不能为空`)
	}

	t.lock.Lock()
	defer t.lock.Unlock()

	if _, exist := t.handlers[endPoint]; exist {
		return fmt.Errorf(`%s 已经注册`, endPoint)
	}

	t.handlers[endPoint] = struct{}{}

	t.bot.Handle(endPoint, func(m *telebot.Message) {
		result := handler(m)
		_, _ = t.bot.Reply(m, escape(result))
	})

	return nil
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

func (n noneTelegram) SendMsg(_ string) error {
	return nil
}

func (n noneTelegram) SendMarkdown(_ string) error {
	return nil
}

func (n noneTelegram) Handle(string, Handler) error {
	return nil
}

func (n noneTelegram) Start() {
}

func (n noneTelegram) SendFile(_ io.Reader, _ string) error {
	return nil
}

func (n noneTelegram) SendFileFromURL(_, _ string) error {
	return nil
}
