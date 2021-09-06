package ktexclient

import (
	"crypto/aes"
	"crypto/cipher"
)

type Crypto interface {
	EncryptWithIV(plainText []byte, iv []byte) ([]byte, error)
	DecryptWithIV(cipherText []byte, iv []byte) ([]byte, error)
}

type CryptoGCM struct {
	block cipher.Block
	key   []byte
}

/*EncryptWithIV 加密
参数:
*	plainText	[]byte	原始数据
*	iv       	[]byte	向量
返回值:
*	[]byte   	[]byte	二进制
*	error    	error 	错误
*/
func (c *CryptoGCM) EncryptWithIV(plainText, iv []byte) ([]byte, error) {
	crypto, err := cipher.NewGCMWithNonceSize(c.block, len(iv))
	if err != nil {
		return nil, err
	}

	cipherText := crypto.Seal(nil, iv, plainText, nil)

	return cipherText, nil
}

/*DecryptWithIV 解密
参数:
*	cipherText	[]byte	加密二进制
*	iv        	[]byte	参数2
返回值:
*	[]byte    	[]byte	返回值1
*	error     	error 	返回值2
*/
func (c *CryptoGCM) DecryptWithIV(cipherText, iv []byte) ([]byte, error) {
	crypto, err := cipher.NewGCMWithNonceSize(c.block, len(iv))
	if err != nil {
		return nil, err
	}

	plainText, err := crypto.Open(nil, iv, cipherText, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}

func NewAESCryptoGCM(key []byte) (Crypto, error) {
	b, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	r := &CryptoGCM{
		block: b,
		key:   key,
	}

	return r, nil
}
