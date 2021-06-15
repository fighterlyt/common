package telegram

import (
	"os"
	"testing"

	"github.com/fighterlyt/log"
)

var (
	testTelegram Telegram
	err          error
	logger       log.Logger
)

func TestMain(m *testing.M) {
	if logger, err = log.NewEasyLogger(true, false, ``, `telegram`); err != nil {
		panic(`telegram` + err.Error())
	}

	os.Exit(m.Run())
}
