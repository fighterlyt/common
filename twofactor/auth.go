package twofactor

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"image/png"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/dgryski/dgoogauth"
	"github.com/pkg/errors"
)

type Auth interface {
	Validate(password string) (ok bool, err error)
	QR(name string) (qrcode, data string, err error)
}

// auth 动态验证
type auth struct {
	secret string
	config *dgoogauth.OTPConfig
}

const (
	width  = 200
	height = 200
	length = 20
)

/*NewAuth 新建动态验证
参数:
*	secret	string	密钥，字符串，至少20位
返回值:
*	auth   	*auth   动态验证
*	err   	error 	错误
*/
func NewAuth(secret string) (target *auth, err error) {
	if len(strings.TrimSpace(secret)) < length {
		return nil, fmt.Errorf(`至少非空[%d]位`, length)
	}

	secret = base32.StdEncoding.EncodeToString([]byte(secret))

	target = &auth{
		secret: secret,
		config: &dgoogauth.OTPConfig{
			Secret: secret,
			UTC:    true,
		},
	}

	return target, nil
}

/*QR 返回谷歌验证的二维码，二维码以base64返回
参数:
*	name  	string	用户名
返回值:
*	qrCode	string	base64编码的二维码
*	err   	error 	错误
*/
func (c auth) QR(name string) (qrcode, data string, err error) {
	var (
		barCode barcode.Barcode
		buffer  = &bytes.Buffer{}
	)

	data = strings.TrimRight(c.config.ProvisionURI(name), `%3D`)

	if barCode, err = qr.Encode(data, qr.M, qr.Auto); err != nil {
		return qrcode, data, errors.Wrap(err, `编码二维码失败`)
	}

	if barCode, err = barcode.Scale(barCode, width, height); err != nil {
		return qrcode, data, errors.Wrap(err, `二维码放大失败`)
	}

	if err = png.Encode(buffer, barCode); err != nil {
		return qrcode, data, errors.Wrap(err, `png编码`)
	}

	return strings.TrimRight(base64.StdEncoding.EncodeToString(buffer.Bytes()), `=`), data, nil
}

func (c auth) Validate(password string) (ok bool, err error) {
	return c.config.Authenticate(password)
}
