package badger

import (
	"encoding/json"
	"testing"

	"github.com/fighterlyt/common/localdb"
	"github.com/fighterlyt/log"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

var (
	service    localdb.Service
	err        error
	dbPath     = `db`
	testLogger log.Logger
	testData   = testStruct{
		ID: "1",
		A:  2,
		B:  false,
		C:  "3",
		D:  decimal.New(4, 0),
	}
)

func TestNewService(t *testing.T) {
	testLogger, err = log.NewEasyLogger(true, false, ``, `badger`)
	require.NoError(t, err, `NewEasyLogger`)

	service, err = NewService(dbPath, testLogger)

	require.NoError(t, err, `NewService`)
}
func TestService_Write(t *testing.T) {
	TestNewService(t)

	type args struct {
		data localdb.Item
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: `测试`,
			args: args{data: &testData},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, service.Write(tt.args.data))
			another := &testStruct{}

			require.NoError(t, service.Read(testData.Key(), another))
			require.EqualValues(t, testData, *another)
		})
	}
}

func TestService_Get(t *testing.T) {
	TestNewService(t)

	another := &testStruct{}

	require.NoError(t, service.Read([]byte(`1`), another))
	t.Log(another)
}

func TestService_Delete(t *testing.T) {
	TestNewService(t)

	key := []byte(testData.ID)
	another := &testStruct{}

	err = service.Read(key, another) // 先拿

	if err != nil {
		require.True(t, service.IsNotFound(err), `未找到是合理的`)
		require.NoError(t, service.Write(&testData), `写入数据`)
	}

	require.NoError(t, service.Delete(key), `删除`)

	err = service.Read(key, another)
	require.Error(t, err, `删除后获取必然报错`)

	require.True(t, service.IsNotFound(err))
}

type testStruct struct {
	ID string
	A  int
	B  bool
	C  string
	D  decimal.Decimal
}

func (t testStruct) Key() []byte {
	return []byte(t.ID)
}

func (t testStruct) Encode() ([]byte, error) {
	return json.Marshal(t)
}

func (t *testStruct) Decode(bytes []byte) error {
	return json.Unmarshal(bytes, t)
}
