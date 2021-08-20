package telegram

import (
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Core struct {
	minLevel zapcore.Level
	fields   []zapcore.Field
	encoder  zapcore.Encoder
	ch       chan string
	lock     *sync.RWMutex
	sender   *telegram
	get      func() ([]string, error)
}

func NewCore(minLevel zapcore.Level, sender *telegram, get func() ([]string, error)) *Core {
	return &Core{
		minLevel: minLevel,
		sender:   sender,
		ch:       make(chan string, 100),
		lock:     &sync.RWMutex{},
		encoder: zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			// Keys can be anything except the empty string.
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}),
		get: get,
	}
}

func (c *Core) Start() {
	go func() {
		for msg := range c.ch {
			if err := c.sender.SendMsg(msg); err != nil {
				c.sender.logger.Error(`发送消息失败`, zap.String(`错误`, err.Error()))
			}
		}
	}()
}

func (c Core) Enabled(level zapcore.Level) bool {
	return c.minLevel <= level
}

func (c *Core) With(fields []zapcore.Field) zapcore.Core {
	// Clone Core.
	clone := *c

	// Clone and append fields.
	clone.fields = make([]zapcore.Field, len(c.fields)+len(fields))
	copy(clone.fields, c.fields)
	copy(clone.fields[len(c.fields):], fields)

	// Done.
	return &clone
}

func (c *Core) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}

	return checked
}

func (c *Core) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	restart, err := c.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return errors.Wrap(err, `EncodeEntry`)
	}

	msg := restart.String()

	var (
		ignores []string
	)

	if c.get != nil {
		if ignores, err = c.get(); err != nil {
			c.ch <- err.Error()
		}
	}

	if Contains(msg, ignores...) {
		return nil
	}

	c.lock.RLock()
	defer c.lock.RUnlock()

	c.ch <- msg

	return nil
}

func (c *Core) Sync() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	for {
		time.Sleep(time.Millisecond)

		if len(c.ch) == 0 {
			return nil
		}
	}
}

func Contains(data string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(data, candidate) {
			return true
		}
	}

	return false
}
