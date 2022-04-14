package telegram

import (
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTelegram(t *testing.T) {
	testTelegram, err = NewTelegram("test", `5383989770:AAEo8p96mLwOm24CPfaP0ztgrk0kYlMfzJg`, logger, -1001586947225)
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

func TestSendFileWithURL(t *testing.T) {
	TestNewTelegram(t)

	require.NoError(t, testTelegram.SendFileFromURL(`https://dubai-real.oss-accelerate-overseas.aliyuncs.com/report/20220413/484392322344_202203_cur.xlsx`, `报表.xlsx`), `SendFileWithURL`)
}
