package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fighterlyt/gotron-sdk/pkg/common"
	"github.com/pkg/errors"
)

// 定义trc20全部的方法签名
const (
	/*
		transfer(address,uint256)： 0xa9059cbb

		balanceOf(address)：0x70a08231

		decimals()：0x313ce567

		allowance(address,address)： 0xdd62ed3e

		symbol()：0x95d89b41

		totalSupply()：0x18160ddd

		name()：0x06fdde03

		approve(address,uint256)：0x095ea7b3

		transferFrom(address,address,uint256)： 0x23b872dd
	*/
	transferMethodID     = "a9059cbb"
	balanceOfMethodID    = "70a08231"
	decimalsMethodID     = "313ce567"
	allowanceMethodID    = "dd62ed3e"
	symbolMethodID       = "95d89b41"
	TotalSupplyMethodID  = "18160ddd"
	nameMethodID         = "06fdde03"
	approveMethodID      = "095ea7b3"
	transferFromMethodID = "23b872dd"
)

// trc20MethodType trc20方法类型
type trc20MethodType int

const (
	trc20Transfer trc20MethodType = iota + 1
	trc20BalanceOf
	trc20Decimals
	trc20Allowance
	trc20Symbol
	trc20TotalSupply
	trc20Name
	trc20Approve
	trc20TransferFrom
	trc20Unknown = 100
	trc20InValid = 101
	hex          = 16
	boundary     = 100000000
)

var (
	idToType = map[string]trc20MethodType{
		transferMethodID:     trc20Transfer,
		balanceOfMethodID:    trc20BalanceOf,
		decimalsMethodID:     trc20Decimals,
		allowanceMethodID:    trc20Allowance,
		symbolMethodID:       trc20Symbol,
		TotalSupplyMethodID:  trc20TotalSupply,
		nameMethodID:         trc20Name,
		approveMethodID:      trc20Approve,
		transferFromMethodID: trc20TransferFrom,
	}
)

type trc20Abi struct {
}

/*MethodType 从合约原始数据判断方法类型
参数:
*	data	string
返回值:
*	trc20MethodType	trc20MethodType
*/
func (t trc20Abi) MethodType(data string) trc20MethodType {
	if len(data) < methodIDLength {
		return trc20InValid
	}

	if result, exist := idToType[data[:methodIDLength]]; exist {
		return result
	}

	return trc20Unknown
}

func (t trc20Abi) getMoneyFromTransfer(data string) string {
	var (
		money string
	)

	if len(data) > zeroValueLength {
		if len(data) > valueStartIndex {
			if len(data) >= valueStartIndex+valueLength {
				money = data[valueStartIndex : valueStartIndex+valueLength]
			} else {
				money = data[valueStartIndex:]
			}
		}

		money = strings.TrimLeft(money, "0")
	}

	return money
}

/*UnpackTransfer 解析trc20协议中的transfer方法,该方法签名为 function transfer(address _to, uint _value) returns (bool success)
算法:
	* 最终的data包括3部分,依次是:methodID to value
		1.  方法签名是固定的4个字节
		2.  address 32个字节，左填充0,hexString
		3.  value 32个字节,做填充0,16进制
注意 d8c3c9833e2f55286858b79c7c76ef90ecefc1f3814e2e802ca75fb3f14e2c03 最终多了一些附加的信息
参数:
*	data	[]byte
返回值:
*	to   	string
*	value	int64
*	err  	error
*/
func (t trc20Abi) UnpackTransfer(data string) (to string, value int64, err error) {
	if data[:methodIDLength] != transferMethodID {
		return "", 0, fmt.Errorf("并非交易数据[%s]", data[:methodIDLength])
	}

	money := t.getMoneyFromTransfer(data)

	to = data[valueStartIndex-42 : valueStartIndex]

	// 必须是41开头
	if to[:2] != "41" {
		to = "41" + to[2:]
	}

	var address []byte

	if address, err = common.Hex2Bytes(to); err != nil {
		return "", 0, errors.Wrapf(err, "解析地址Hex错误[%s]", to)
	}

	to = common.EncodeCheck(address)

	if money == `` {
		return to, 0, nil
	}

	if value, err = strconv.ParseInt(money, hex, 64); err != nil {
		return "", 0, errors.Wrapf(err, "解析金额[%s]", money)
	}

	return to, value, nil
}

/*UnpackApprove 解析trc20协议中的approve方法,该方法签名为 function approve(address _to, uint _value) returns (bool success)
算法:
	* 最终的data包括3部分,依次是:methodID to value
		1.  方法签名是固定的4个字节
		2.  address 32个字节，左填充0,hexString
		3.  value 32个字节,做填充0,16进制
注意 d8c3c9833e2f55286858b79c7c76ef90ecefc1f3814e2e802ca75fb3f14e2c03 最终多了一些附加的信息
参数:
*	data	[]byte
返回值:
*	to   	string
*	value	int64
*	err  	error
*/
func (t trc20Abi) UnpackApprove(data string) (to string, value int64, err error) {
	if data[:methodIDLength] != approveMethodID {
		return "", 0, fmt.Errorf("并非授权数据[%s]", data[:methodIDLength])
	}

	money := t.getMoneyFromTransfer(data)

	to = data[valueStartIndex-42 : valueStartIndex]

	// 必须是41开头
	if to[:2] != "41" {
		to = "41" + to[2:]
	}

	var address []byte

	if address, err = common.Hex2Bytes(to); err != nil {
		return "", 0, errors.Wrapf(err, "解析地址Hex错误[%s]", to)
	}

	to = common.EncodeCheck(address)

	if money == `` {
		return to, 0, nil
	}

	if value, err = strconv.ParseInt(money, hex, 64); err != nil {
		if strings.Contains(err.Error(), "value out of range") {
			return to, boundary, nil
		}

		return "", 0, errors.Wrapf(err, "解析金额[%s]", money)
	}

	return to, value, nil
}

/*UnpackTransferFrom 解析trc20协议中的approve方法,该方法签名为 function approve(address _to, uint _value) returns (bool success)
算法:
	* 最终的data包括3部分,依次是:methodID to value
		1.  方法签名是固定的4个字节
		2.  address 32个字节，左填充0,hexString
		3.  recipient address 32个字节
		3.  value 32个字节,做填充0,16进制
注意 d8c3c9833e2f55286858b79c7c76ef90ecefc1f3814e2e802ca75fb3f14e2c03 最终多了一些附加的信息
参数:
*	data	[]byte
返回值:
*	to   	string
*	value	int64
*	err  	error
*/
func (t trc20Abi) UnpackTransferFrom(data string) (to string, value int64, err error) {
	if data[:methodIDLength] != transferFromMethodID {
		return "", 0, fmt.Errorf("并非transferFrom数据[%s]", data[:methodIDLength])
	}

	var money string

	if len(data) > zeroValueLength {
		if len(data) >= valueStartIndex+valueLength+addressLength {
			money = data[valueStartIndex+addressLength : valueStartIndex+valueLength+addressLength]
		} else {
			money = data[valueStartIndex+addressLength:]
		}

		money = strings.TrimLeft(money, "0")
	}

	to = data[valueStartIndex-42 : valueStartIndex]

	// 必须是41开头
	if to[:2] != "41" {
		to = "41" + to[2:]
	}

	var address []byte

	if address, err = common.Hex2Bytes(to); err != nil {
		return "", 0, errors.Wrapf(err, "解析地址Hex错误[%s]", to)
	}

	to = common.EncodeCheck(address)

	if money == `` {
		return to, 0, nil
	}

	if value, err = strconv.ParseInt(money, hex, 64); err != nil {
		return "", 0, errors.Wrapf(err, "解析金额[%s]", money)
	}

	return to, value, nil
}

// 这里定义了一组原始数据长度相关的常量
const (
	methodIDLength  = 8
	addressLength   = 64
	valueStartIndex = methodIDLength + addressLength
	valueLength     = 64
	// 转了0个trc20
	zeroValueLength = methodIDLength + addressLength
	trc20Length     = zeroValueLength + valueLength
)
