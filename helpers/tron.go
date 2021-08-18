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
)

/*ValidateAddress 验证钱包地址是否合法
参数:
*	addr	string	参数1
返回值:
*	bool	bool  	返回值1
*/
func ValidateAddress(addr string) bool {
	if len(addr) != 34 {
		return false
	}

	if addr[0:1] != "T" {
		return false
	}

	_, err := common.DecodeCheck(addr)

	return err == nil
}

func IsPrivateKeyMatched(addr, privateKeyHex string) (matched bool, err error) {
	var privateKey *ecdsa.PrivateKey

	if privateKey, err = crypto.HexToECDSA(privateKeyHex); err != nil {
		return false, errors.Wrapf(err, "私钥非法")
	}

	var fromAddress = tronAddress.PubkeyToAddress(privateKey.PublicKey)

	return fromAddress.String() == addr, nil
}

func ValidatePrivateKey(key string) bool {
	_, err := crypto.HexToECDSA(key)
	return err == nil
}
