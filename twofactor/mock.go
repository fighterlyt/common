package twofactor

import (
	"github.com/fighterlyt/log"
	"go.uber.org/zap"
)

type mockNotify struct {
	logger log.Logger
}

func newMockNotify(logger log.Logger) *mockNotify {
	return &mockNotify{logger: logger}
}

func (m mockNotify) SendTo(userID int64, id, message string) error {
	m.logger.Info(`发起通知`, zap.Int64(`用户ID`, userID), zap.Strings(`id/消息`, []string{id, message}))
	return nil
}
