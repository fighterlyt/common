package es

import (
	"fmt"
	"os"
	"testing"

	"github.com/fighterlyt/common/helpers"
	"github.com/fighterlyt/log"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

var (
	logger log.Logger
	err    error
	client *Client
)

func TestMain(m *testing.M) {
	if logger, err = log.NewEasyLogger(true, false, ``, `es`); err != nil {
		panic(err.Error())
	}

	args := NewClientArgument(logger, []string{`http://localhost:9200`}, ``, ``)

	if client, err = NewClient(args); err != nil {
		panic(err.Error())
	}

	os.Exit(m.Run())
}
func TestNewClient(t *testing.T) {
	type args struct {
		argument *ClientArgument
	}
	tests := []struct {
		name       string
		args       args
		wantClient *Client
	}{
		{
			name: "正确无需用户名",
			args: args{NewClientArgument(logger, []string{`http://localhost:9200`}, ``, ``)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = NewClient(tt.args.argument)
			require.NoError(t, err)
		})
	}
}

func TestClient_Add(t *testing.T) {
	require.NoError(t, client.Add(TestData{
		ID:    2,
		Str:   "3",
		Value: decimal.New(50, 1),
		Bool:  false,
		Time:  helpers.Now(),
	}), `写入文档`)

}

type TestData struct {
	ID    int64           `json:"id"`
	Str   string          `json:"str"`
	Value decimal.Decimal `json:"value"`
	Bool  bool            `json:"bool"`
	Time  helpers.Time    `json:"time"`
}

func (TestData) Index() string {
	return `test`
}

func (t TestData) GetID() string {
	return fmt.Sprintf(`%d`, t.ID)
}
