package twofactor

import (
	"github.com/shopspring/decimal"
)

// Notify 通知服务
type Notify interface {
	// SendTo 通知,参数分别为通知用户ID,记录id,通知内容
	SendTo(userIDs []int64, id, message string) error
}

// Audit 审核服务
type Audit interface {
	// Audit 审核,用户ID,协议、币种、金额
	Audit(userID int64, protocol, symbol string, amount decimal.Decimal) (need bool, err error)
}

// Service 服务
type Service interface {
	// Process 处理，参数分别为 记录id、提现用户id、通知用户ID、协议、币种、通知信息、金额,返回need==true 表示已经需要动态校验发出通知
	Process(id string, userID int64, notifyUserIDs []int64, protocol, symbol, notifyMessage string, amount decimal.Decimal) (need bool, err error) //nolint:lll
	// Auth 校验
	Auth(password string) (ok bool, err error)
	// QR 获取二维码
	QR(name string) (qrcode string, err error)
}
