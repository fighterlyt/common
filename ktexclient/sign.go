package ktexclient

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"math/rand"
	"time"
)

/*
BuildVerificationString
获取验签名串
格式为：
	应答时间戳\n
	应答随机串\n
	应答报文主体\n
*/
func BuildVerificationString(timestamp, nonce string, body []byte) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	bufw := bufio.NewWriter(buffer)
	_, _ = bufw.WriteString(timestamp)
	_ = bufw.WriteByte('\n')
	_, _ = bufw.WriteString(nonce)
	_ = bufw.WriteByte('\n')

	if len(body) != 0 {
		_, _ = bufw.Write(body)
	}

	_ = bufw.WriteByte('\n')

	err := bufw.Flush()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// NonceStr 生成随机字符串
func NonceStr() string {
	rand.Seed(time.Now().UnixNano())
	byteLen := 16
	randBytes := make([]byte, byteLen)

	for i := 0; i < byteLen; i++ {
		randBytes[i] = byte(rand.Intn(256)) //nolint:gosec
	}

	return hex.EncodeToString(randBytes)
}
