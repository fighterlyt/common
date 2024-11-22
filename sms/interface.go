package sms

import (
	"github.com/fighterlyt/common/model"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type Service interface {
	// DirectSend 直接发送，内容直接发送
	DirectSend(target, content string) error
	// TemplateSend 模板发送，内容需要和模板匹配
	TemplateSend(target, content, id string) error
	// Support 是否支持某种发送
	Support(supported Supported) bool
	Balance() (balance decimal.Decimal, err error)
}

type Supported int

const (
	// SupportDirectSend 支持直接发送
	SupportDirectSend Supported = 1
	// SupportTemplateSend 支持模板发送
	SupportTemplateSend Supported = 2
)

var (
	// ErrNotSupported 不支持的发送方式
	ErrNotSupported = errors.New(`not supported`)
)

type RecordAccess interface {
	SetFinish(id string, err error) error
	GetFinishStatus(id string) (status SendStatus, err error)
}

// SendStatus 发送状态
type SendStatus int

const (
	// SendAll 所有
	SendAll SendStatus = 0
	// SendSuccess 发送成功
	SendSuccess SendStatus = 1
	// SendFail 发送失败
	SendFail SendStatus = 2
	// SendUnknown 发送结果未知，尚未知晓
	SendUnknown SendStatus = 3
)

func (s SendStatus) Value() int {
	return int(s)
}

func (s SendStatus) Text() string {
	switch s {
	case SendAll:
		return "所有"
	case SendSuccess:
		return `发送成功`
	case SendFail:
		return `发送失败`
	case SendUnknown:
		return `发送结果未知`
	default:
		return "未知状态"
	}
}

func (s SendStatus) MarshalJSON() ([]byte, error) {
	return model.MarshalJSON(s)
}
