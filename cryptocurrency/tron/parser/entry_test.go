package parser

import (
	"os"
	"testing"

	"github.com/fighterlyt/log"
)

var (
	logger log.Logger
	err    error
)

func TestMain(m *testing.M) {
	if logger, err = log.NewEasyLogger(true, false, ``, `测试`); err != nil {
		panic(err.Error())
	}

	os.Exit(m.Run())
}
