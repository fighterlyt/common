package twofactor

import (
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type service struct {
	auth   Auth
	audit  Audit
	notify Notify
}

/*NewService 新建服务
参数:
*	auth    	*auth    	动态验证
*	audit   	Audit   	审核
*	notify  	Notify  	通知
返回值:
*	*service	*service	服务
*/
func NewService(auth Auth, audit Audit, notify Notify) *service {
	return &service{auth: auth, audit: audit, notify: notify}
}

/*Process 处理
参数:
*	id           	string         	记录ID
*	userID       	int64          	提现用户ID
*	notifyUserID 	int64          	通知用户ID
*	protocol     	string         	提现协议
*	symbol       	string         	提现币种
*	notifyMessage	string         	通知信息
*	amount       	decimal.Decimal	提现金额
返回值:
*	need         	bool           	是否需要动态验证，true表示需要且发出通知
*	err          	error          	错误
*/
func (s service) Process(id string, userID int64, notifyUserIDs []int64, protocol, symbol, notifyMessage string, amount decimal.Decimal) (need bool, err error) { //nolint:lll
	if need, err = s.audit.Audit(userID, protocol, symbol, amount); err != nil {
		return false, errors.Wrap(err, `审核`)
	}

	if need {
		if err = s.notify.SendTo(notifyUserIDs, id, notifyMessage); err != nil {
			return false, errors.Wrap(err, `通知`)
		}
	}

	return need, nil
}

/*Auth 动态验证
参数:
*	password	string	动态密码
返回值:
*	ok      	bool  	密码是否正确
*	err     	error 	错误
*/
func (s service) Auth(password string) (ok bool, err error) {
	return s.auth.Validate(password)
}

/*QR 返回谷歌验证的二维码，二维码以base64返回
参数:
*	user  	string	用户名
返回值:
*	qrCode	string	base64编码的二维码
*	err   	error 	错误
*/
func (s service) QR(user string) (qrCode string, err error) {
	qrCode, _, err = s.auth.QR(user)
	return qrCode, err
}

// auditBySingleAmount 基于单笔金额审核
type auditBySingleAmount struct {
	// get 获取阈值方法
	get func() (decimal.Decimal, error)
}

/*NewAuditBySingleAmount 新建基于单笔金额审核
参数:
*	get                 	func() (decimal.Decimal, error)	获取
返回值:
*	*auditBySingleAmount	*auditBySingleAmount           	服务
*/
func NewAuditBySingleAmount(get func() (decimal.Decimal, error)) *auditBySingleAmount {
	return &auditBySingleAmount{get: get}
}

/*Audit 方法说明
参数:
*	_     	int64          	用户ID
*	_     	string         	协议
*	_     	string         	币种
*	amount	decimal.Decimal	金额
返回值:
*	need  	bool           	返回值1
*	err   	error          	返回值2
*/
func (a auditBySingleAmount) Audit(_ int64, _, _ string, amount decimal.Decimal) (need bool, err error) {
	var (
		threshold decimal.Decimal
	)

	if threshold, err = a.get(); err != nil {
		return false, errors.Wrap(err, `获取最小阈值错误`)
	}

	return amount.GreaterThanOrEqual(threshold), nil
}
