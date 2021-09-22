package helpers

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/pkg/errors"
)

/*Base64BarCode base64编码的二维码，格式为png
参数:
*	data  	string	需要编码的数据
*	width 	int   	宽度
*	height	int   	高度
返回值:
*	qrcode	string	base64
*	err   	error 	错误
*/
func Base64BarCode(data string, width, height int) (qrcode string, err error) {
	var (
		barCode barcode.Barcode
		buffer  = &bytes.Buffer{}
	)

	if barCode, err = qr.Encode(data, qr.M, qr.Auto); err != nil {
		return qrcode, errors.Wrap(err, `编码二维码失败`)
	}

	if barCode, err = barcode.Scale(barCode, width, height); err != nil {
		return qrcode, errors.Wrap(err, `二维码放大失败`)
	}

	if err = png.Encode(buffer, barCode); err != nil {
		return qrcode, errors.Wrap(err, `png编码`)
	}

	return strings.TrimRight(base64.StdEncoding.EncodeToString(buffer.Bytes()), `=`), nil
}
