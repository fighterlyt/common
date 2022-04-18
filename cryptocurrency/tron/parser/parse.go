package parser

import (
	"context"

	"gitlab.com/nova_dubai/common/cryptocurrency"
)

// TronParser 波场区块解析器,是对处理器的进一步抽象
type TronParser interface {
	Parse(ctx context.Context, blockNumber int64) (trades []*cryptocurrency.Trade, err error)
}
