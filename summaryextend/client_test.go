package summary

import (
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

var (
	dayClient     Client
	historyClient Client
)

func TestDayClient(t *testing.T) {
	dayClient, err = NewClient(`summary_1`, SlotDay, logger, db)
	require.NoError(t, err, `构建Client`)
}

func TestHistoryClient(t *testing.T) {
	historyClient, err = NewClient(`historyIncome`, SlotWhole, logger, db)
	require.NoError(t, err, `构建Client`)
}

func TestClient_Summarize(t *testing.T) {
	TestDayClient(t)
	require.NoError(t, dayClient.Summarize(1, decimal.New(1, 0)))
}

func BenchmarkClient_Summarize(b *testing.B) {
	dayClient, err = NewClient(`summary_1`, SlotDay, logger, db)
	require.NoError(b, err, `构建Client`)

	b.ResetTimer()

	times := 10

	count := 1000

	wg := &sync.WaitGroup{}
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			for j := 0; j < times; j++ {
				require.NoError(b, dayClient.Summarize(1, decimal.New(1, 0)))
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestClient_Get(t *testing.T) {
	TestHistoryClient(t)
	now := time.Now().Unix()
	records, err := historyClient.GetSummary([]int64{2}, now, now+1)

	require.NoError(t, err)
	t.Log(records)
}
