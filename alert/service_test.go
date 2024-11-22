package alert

import (
	"testing"

	"github.com/fighterlyt/common/telegram"
	"github.com/fighterlyt/log"
	"github.com/stretchr/testify/require"
)

var (
	service Service
)

func TestNewTelegramService(t *testing.T) {
	logger, err := log.NewEasyLogger(true, false, ``, `通知`)
	require.NoError(t, err, `构建日志器`)

	tele, err := telegram.NewTelegram(`测试`, `1920269352:AAFc-xSggJoSYjDY7LNw_JWHIFX-k_CcgPQ`, logger, -1001405517538)
	require.NoError(t, err, `构建飞机`)

	service = NewTelegramService(tele, logger)

}
func Test_telegramService_SendMarkDown(t *testing.T) {
	TestNewTelegramService(t)

	type args struct {
		data string
		md   bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: `普通文本`,
			args: args{data: `123`, md: false},
		},
		{
			name: `markdown`,
			args: args{data: `md*123*`, md: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			if tt.args.md {
				service.SendMarkDown(tt.args.data)
			} else {
				service.SendText(tt.args.data)
			}
		})
	}
}
