package ktexclient

import (
	"bytes"
	"encoding/base64"
	"github.com/stretchr/testify/require"
	"net/http"
	"reflect"
	"testing"
)

func TestNewClient(t *testing.T) {

	s := "https://8e11-94-205-10-241.ngrok.io"
	key = []byte("k1QbNxsE1nuO1k+QL8Zd4ePSLUXHM9KO") // 32
	iv = []byte("EDOCnElJ2EzTZWER")                  // 16

	client, err := NewClient(nil, key, iv)
	require.NoError(t, err)
	body := "我是JSON请求Body"
	bodyByte := []byte(body)
	request, err := http.NewRequest(http.MethodPost, s, bytes.NewReader(bodyByte))
	require.NoError(t, err)
	response, err := client.Do(request)
	require.NoError(t, err)
	err = response.Body.Close()
	require.NoError(t, err)

	sign := request.Header.Get("Sign")
	str := request.Header.Get("Nonce")
	unixStr := request.Header.Get("Timestamp")
	aesCryptoGCM, err := NewAESCryptoGCM(key)
	require.NoError(t, err)
	decodeString, err := base64.StdEncoding.DecodeString(sign)
	require.NoError(t, err)
	decryptText, err := aesCryptoGCM.DecryptWithIV(decodeString, iv)
	require.NoError(t, err)
	verificationString, err := BuildVerificationString(unixStr, str, bodyByte)
	require.NoError(t, err)
	if !reflect.DeepEqual(verificationString, decryptText) {
		t.Error("not equal")
	}
}
