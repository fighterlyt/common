package telegram

import (
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTelegram(t *testing.T) {
	testTelegram, err = NewTelegram("test", `1886019351:AAGOCfqMyPC-xIbqpN0WsEkv0fkERKdnyE8`, logger, -1001376524055)
	require.NoError(t, err, `NewTelegram`)
}

func TestSendMsg(t *testing.T) {
	TestNewTelegram(t)

	require.NoError(t, testTelegram.SendMsg(`测试`), `SendMsg`)
}

func TestSendMarkdown(t *testing.T) {
	TestNewTelegram(t)

	require.NoError(t, testTelegram.SendMarkdown(`**测试**`), `SendMarkdown`)
}

func TestPanic(t *testing.T) {
	TestNewTelegram(t)

	defer func() {
		if x := recover(); x != nil {
			var stackMsg string
			for i := 0; i < 10; i++ {
				stackMsg = stackMsg + string(debug.Stack())
			}
			require.NoError(t, testTelegram.SendMsg(fmt.Sprintf("服务[%s]发生panic,错误信息[%v],堆栈信息[%s]", `serviceName`, err, stackMsg)))
		}
	}()

	panic(`1`)
}
