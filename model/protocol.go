package model

// 币种
const (
	TRX = "TRX"
	// SGMT shasta 网络用于测试的token
	SGMT = "SGMT"
	//  USDT
	USDT = "USDT"
	//  ETH
	ETH = "ETH"
	FLY = "FLY"
)

// Protocol 协议
type Protocol string

const (
	// Trc20 协议
	Trc20 Protocol = "trc20"
	// Erc20 协议
	Erc20 Protocol = "erc20"
)

// Support 是否支持协议,目前只支持trc20
func (p Protocol) Support() bool {
	return p == Trc20 || p == Erc20
}

func (p Protocol) ContractLocator() ContractLocator {
	switch p {
	case Trc20:
		return tronContractLocator
	case Erc20:
		return ethContractLocator
	default:
		return nil
	}
}

func (p Protocol) String() string {
	return string(p)
}
