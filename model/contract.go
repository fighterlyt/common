package model

import (
	"fmt"
)

// 定义了和协议相同的一些数据
// 以太坊合约地址
const (
	USDTContractAddress = "0xdac17f958d2ee523a2206206994597c13d831ec7"
	FLYContractAddress  = "0x4E7491942F165262cc329C37f7e587A305b18292"
)

//  波场合约地址
const (
	TronUSDTContractAddress = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	TronSGMTContractAddress = "TRZTJyNpKVevp959982XBbkjm7qrLxYTWi"
)

// Contract 表示一个合约
type Contract interface {
	Address() string  // 合约地址
	Kind() string     // 代码
	Token() string    // 币种
	Precision() int32 // 精度
}

type contract struct {
	address   string
	kind      string
	token     string
	precision int32
}

func newContract(address, kind, token string, precision int32) Contract {
	return &contract{address: address, kind: kind, token: token, precision: precision}
}

func (c contract) Address() string {
	return c.address
}

func (c contract) Kind() string {
	return c.kind
}
func (c contract) Token() string {
	return c.token
}

func (c contract) Precision() int32 {
	return c.precision
}

// ContractLocator 合约定位器
type ContractLocator interface {
	GetContract(currency string) (Contract, error)
}

// mapContractLocator 基于map的以太坊合约地址定位器
type mapContractLocator struct {
	contracts map[string]Contract // 合约代号->合约地址，由于只读，不需要锁
}

var (
	// tron合约实例
	tronContractLocator ContractLocator
	ethContractLocator  ContractLocator
)

/*InitContractLocator 初始化合约
参数:
*	ethSupports 	[]string	eth支持的合约
*	tronSupports	[]string	tron支持的合约
返回值:
*/
func InitContractLocator(ethSupports, tronSupports []string) {
	tronContractLocator = newTronContractLocator(tronSupports)
	ethContractLocator = newEthContractLocator(ethSupports)
}

/*GetContractAddress 获取合约地址,如果合约不存在或者未定义，返回空字符串
参数:
*	currency	string
返回值:
*	string	string
*/
func (m mapContractLocator) GetContract(currency string) (Contract, error) {
	contract := m.contracts[currency]

	if contract == nil {
		return nil, fmt.Errorf("不支持的合约[%s]", currency)
	}

	return contract, nil
}

func newTronContractLocator(supports []string) ContractLocator {
	locator := &mapContractLocator{
		contracts: make(map[string]Contract),
	}

	for _, support := range supports {
		switch support {
		case USDT:
			locator.contracts[USDT] = ContractTRC20USDT
		case SGMT:
			locator.contracts[SGMT] = ContractTRC20SGMT
		}
	}

	return locator
}

func newEthContractLocator(supports []string) ContractLocator {
	locator := &mapContractLocator{
		contracts: make(map[string]Contract),
	}

	for _, support := range supports {
		switch support {
		case USDT:
			locator.contracts[USDT] = ContractERC20USDT
		case FLY:
			locator.contracts[FLY] = ContractERC20Fly
		}
	}

	return locator
}

var (
	ContractTRC20SGMT = newContract(TronSGMTContractAddress, "test", SGMT, 6)
	ContractTRC20USDT = newContract(TronUSDTContractAddress, "production", USDT, 6)

	ContractERC20Fly  = newContract(FLYContractAddress, "test", SGMT, 8)
	ContractERC20USDT = newContract(USDTContractAddress, "production", USDT, 6)
)
