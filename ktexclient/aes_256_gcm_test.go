package ktexclient

import (
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	key = []byte("k1QbNxsE1nuO1k+QL8Zd4ePSLUXHM9KO") // 32
	iv  = []byte("EDOCnElJ2EzTZWER")                 // 16
)

func TestCryptoGCM_EncryptWithIV(t *testing.T) {

	cryptoGCM, err := NewAESCryptoGCM(key)
	require.NoError(t, err)

	type args struct {
		plainText []byte
		iv        []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "测试",
			args: args{
				plainText: []byte("我是请求json"),
				iv:        iv,
			},
			want:    "y8c/dW/cuMSz+/mBg7lTXg5c9jGMZobzuZmCitW4cn8=",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := cryptoGCM.EncryptWithIV(tt.args.plainText, tt.args.iv)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptWithIV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			encodeToString := base64.StdEncoding.EncodeToString(got)
			if !reflect.DeepEqual(encodeToString, tt.want) {
				t.Errorf("EncryptWithIV() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCryptoGCM_DecryptWithIV(t *testing.T) {
	cryptoGCM, err := NewAESCryptoGCM(key)
	require.NoError(t, err)
	type args struct {
		cipherTextBase64 string
		iv               []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "测试",
			args: args{
				cipherTextBase64: "y8c/dW/cuMSz+/mBg7lTXg5c9jGMZobzuZmCitW4cn8=",
				iv:               iv,
			},
			want:    []byte("我是请求json"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipherText, err := base64.StdEncoding.DecodeString(tt.args.cipherTextBase64)
			require.NoError(t, err)
			got, err := cryptoGCM.DecryptWithIV(cipherText, tt.args.iv)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptWithIV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecryptWithIV() got = %v, want %v", got, tt.want)
			}
		})
	}
}
