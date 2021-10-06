package helpers

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	tronAddress "github.com/fighterlyt/gotron-sdk/pkg/address"
	"github.com/fighterlyt/gotron-sdk/pkg/common"
	"github.com/pkg/errors"
)

const (
	// Energy 已激活的钱包，转账消耗能量
	Energy = 14631
	// Bandwidth TRC-20转账消耗带宽
	Bandwidth = 345
	// UnActivatedEnergy 未激活的钱包收款，需要消耗能量
	UnActivatedEnergy = Energy + 15000
	// legalAddrLen 合法的地址长度
	legalAddrLen = 34
)

/*ValidateAddress 验证钱包地址是否合法
参数:
*	addr	string	参数1
返回值:
*	bool	bool  	返回值1
*/
func ValidateAddress(addr string) bool {
	if len(addr) != legalAddrLen {
		return false
	}

	if addr[0:1] != "T" {
		return false
	}

	_, err := common.DecodeCheck(addr)

	return err == nil
}

/*IsPrivateKeyMatched 公钥私钥是否为一对
参数:
*	addr         	string	待验证的公钥
*	privateKeyhex	string	私钥十六进制字符串
返回值:
*	matched      	bool  	是一对
*	err          	error 	验证过程中发生错误
*/
func IsPrivateKeyMatched(addr, privateKeyHex string) (matched bool, err error) {
	var privateKey *ecdsa.PrivateKey

	if privateKey, err = crypto.HexToECDSA(privateKeyHex); err != nil {
		return false, errors.Wrapf(err, "私钥非法")
	}

	var fromAddress = tronAddress.PubkeyToAddress(privateKey.PublicKey)

	return fromAddress.String() == addr, nil
}

/*ValidatePrivateKey 验证是否为私钥
参数:
*	key 	string	私钥十六进制字符串
返回值:
*	bool	bool  	是否为私钥
*/
func ValidatePrivateKey(key string) bool {
	_, err := crypto.HexToECDSA(key)
	return err == nil
}
