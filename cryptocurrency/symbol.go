package cryptocurrency

import "fmt"

// Symbol 币种符号
type Symbol string

const (
	// USDTSymbol USDT
	USDTSymbol Symbol = `USDT`
)

func (s Symbol) String() string {
	return string(s)
}

func (s Symbol) Validate() error {
	switch s {
	case USDTSymbol:
		return nil
	default:
		return fmt.Errorf(`不支持的币种[%s]`, s.String())
	}
}
