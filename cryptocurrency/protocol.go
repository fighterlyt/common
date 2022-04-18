package cryptocurrency

// Protocol 协议
type Protocol string

const (
	// Trc20 协议
	Trc20 Protocol = "trc20"
	// Erc20 协议
	Erc20 Protocol = "erc20"
)

// Support 是否支持协议,目前只支持trc20和ERC20
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

/*GetSymbol 获取币种
参数:
*	test	bool  	是否为
返回值:
*	symbol	symbol	币种
*/
func GetSymbol(protocol Protocol, test bool) Symbol {
	switch protocol {
	case Trc20:
		if test {
			return SGMT
		}

		return USDT
	case Erc20:
		if test {
			return FLY
		}

		return USDT
	default:
		return USDT
	}
}
