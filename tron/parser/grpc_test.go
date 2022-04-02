package parser

import (
	"context"
	"testing"

	"github.com/fighterlyt/gotron-sdk/pkg/client"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gitlab.com/nova_dubai/usdtpay/model"
	"google.golang.org/grpc"
)

func TestGrpcParser_Parse(t *testing.T) {
	cl := client.NewGrpcClient("47.241.192.246:50051")
	grpcParser := NewGRPCTronScanParser(mockConcern{}, cl, resource.Logger, model.ContractTRC20USDT, nil)

	require.NoError(t, cl.Start(grpc.WithInsecure()))
	block, err := cl.GetNowBlock()
	require.NoError(t, err, `获取最新区块`)
	t.Log(block.BlockHeader.RawData.Number)
	trades, err := grpcParser.Parse(context.Background(), 36171457)
	require.NoError(t, err)

	for _, trade := range trades {
		t.Log(trade.Token, trade.From, trade.To, trade.Amount, trade.Fee, trade.BlockNum, trade.ID, trade.TradeKind)
	}
}

type mockConcern struct {
}

func (m mockConcern) FilterConcernedAccounts(from, to string, amount decimal.Decimal) (matched bool, data interface{}, err error) {
	return true, nil, nil
}
