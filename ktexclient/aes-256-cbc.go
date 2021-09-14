package ktexclient

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

type CryptoCBC struct {
	block cipher.Block
	key   []byte
}

func NewAESCryptoCBC(key []byte) (Crypto, error) {
	b, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	r := &CryptoCBC{
		block: b,
		key:   key,
	}

	return r, nil
}

func __pkcs7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func __pkcsUnPadding(text []byte) []byte {
	n := len(text)
	if n == 0 {
		return text
	}
	paddingSize := int(text[n-1])
	return text[:n-paddingSize]
}

func (a *CryptoCBC) EncryptWithIV(plainText []byte, iv []byte) ([]byte, error) {
	plainText = __pkcs7Padding(plainText, a.block.BlockSize())
	cipherText := make([]byte, len(plainText))
	crypto := cipher.NewCBCEncrypter(a.block, iv)
	crypto.CryptBlocks(cipherText, plainText)

	return cipherText, nil
}

func (a *CryptoCBC) DecryptWithIV(cipherText []byte, iv []byte) ([]byte, error) {
	plainText := make([]byte, len(cipherText))
	crypto := cipher.NewCBCDecrypter(a.block, iv)
	crypto.CryptBlocks(plainText, cipherText)
	plainText = __pkcsUnPadding(plainText)

	return plainText, nil
}
