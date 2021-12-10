package summaryextend

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

var (
	chainClient Client
)

func TestChain(t *testing.T) {
	TestDayClient(t)
	TestHistoryClient(t)

	chainClient, err = Chain(dayClient, historyClient)
	require.NoError(t, err, `chain`)
}

func TestChain_Add(t *testing.T) {
	TestChain(t)
	require.NoError(t, chainClient.Summarize(`1`, decimal.New(1, 0)))
}
