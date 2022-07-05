package summaryextend

import (
	"sync"
	"testing"
	"time"

	"gitlab.com/nova_dubai/common/helpers"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

var (
	dayClient     Client
	historyClient Client
)

func TestDayClient(t *testing.T) {
	dayClient, err = NewClient(`summary_extend`, SlotDay, logger, db)
	require.NoError(t, err, `构建Client`)
}

func TestHistoryClient(t *testing.T) {
	historyClient, err = NewClient(`summary_offline_charge_day`, SlotWhole, logger, db)
	require.NoError(t, err, `构建Client`)
}

func TestClient_Summarize(t *testing.T) {
	TestDayClient(t)
	require.NoError(t, dayClient.Summarize(`1`, decimal.New(1, 0), decimal.New(2, 0), decimal.New(3, 0), decimal.New(4, 0)))
}

func BenchmarkClient_Summarize(b *testing.B) {
	dayClient, err = NewClient(`summary_extend`, SlotDay, logger, db)
	require.NoError(b, err, `构建Client`)

	b.ResetTimer()

	times := 10

	count := 1000

	wg := &sync.WaitGroup{}
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			for j := 0; j < times; j++ {
				require.NoError(b, dayClient.Summarize(`2`, decimal.New(1, 0)))
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestClient_Get(t *testing.T) {
	TestHistoryClient(t)

	now := time.Now().Unix()

	var (
		records []Summary
	)

	records, err = historyClient.GetSummary(nil, now, now+1)

	require.NoError(t, err)
	t.Log(records)
}

func TestClient_GetDay(t *testing.T) {
	TestDayClient(t)

	now := time.Now().Unix()

	var (
		records []Summary
	)

	records, err = dayClient.GetSummary(nil, now, now+1)

	require.NoError(t, err)
	t.Log(records)
}
func Test_client_GetSummarySummary(t *testing.T) {
	TestHistoryClient(t)

	var (
		record Summary
	)

	record, err = historyClient.GetSummarySummary(nil, 0, 0)

	require.NoError(t, err)

	t.Logf("%+v", record)
}

func TestClient_SummarizeDayFirstUpdate(t *testing.T) {
	TestDayClient(t)
	require.NoError(t, dayClient.SummarizeDayOptimism(1, `1`, decimal.New(1, 0), decimal.New(2, 0), decimal.New(3, 0), decimal.New(4, 0)))
}

func TestClient_SummarizeFirstUpdate(t *testing.T) {
	TestDayClient(t)

	helpers.SetTimeZone(helpers.GetBeiJin())
	require.NoError(t, dayClient.SummarizeOptimism(`1`, decimal.New(1, 0), decimal.New(2, 0), decimal.New(3, 0), decimal.New(4, 0)))

	helpers.SetTimeZone(helpers.GetBeiJin())
	require.NoError(t, dayClient.SummarizeOptimism(`2`, decimal.New(10, 0), decimal.New(20, 0), decimal.New(30, 0), decimal.New(40, 0)))
}
