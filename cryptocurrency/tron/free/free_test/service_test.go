package free_test

import (
	"testing"

	"github.com/fighterlyt/gotron-sdk/pkg/client"
	"github.com/fighterlyt/log"
	"github.com/fighterlyt/redislock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gitlab.com/nova_dubai/common/cryptocurrency/tron/free"
	"gorm.io/gorm"
)

func TestNewService(t *testing.T) {
	type args struct {
		db         *gorm.DB
		logger     log.Logger
		tronClient *client.GrpcClient
		locker     redislock.Locker
	}

	tests := []struct {
		name    string
		args    args
		wantS   free.Service
		wantErr bool
	}{
		{
			name:    `全空`,
			args:    args{},
			wantErr: true,
		},
		{
			name: `全有`,
			args: args{
				db:         db,
				logger:     logger,
				tronClient: g,
				locker:     locker,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = free.NewService(tt.args.db, tt.args.logger, tt.args.tronClient, tt.args.locker)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewService1(t *testing.T) {
	service, err = free.NewService(db, logger, g, locker)
	require.NoError(t, service.Add(testHook{}))
	require.Error(t, service.Add(testHook{}))
	require.NoError(t, err)
}

func TestService_SetUp1(t *testing.T) {
	TestNewService1(t)

	require.NoError(t, service.SetUp(from, privateKey))
}

func TestService_Freeze(t *testing.T) {
	TestService_SetUp1(t)

	// 冻结金额太大
	require.Error(t, service.Freeze(`TQDKRosJu7ktdER5h4e864cXFxGePMiy9P`, decimal.New(10000, 1)))
	require.True(t, beforeFreeze)
	require.True(t, afterFreeze)
	require.Error(t, afterFreezeError)

	require.NoError(t, service.Freeze(`TQDKRosJu7ktdER5h4e864cXFxGePMiy9P`, decimal.New(10, 0)))
}

func Test_service_SetUp(t *testing.T) {
	TestNewService1(t)

	type args struct {
		from       string
		privateKey string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    `空私钥和地址，报错`,
			wantErr: true,
		},
		{
			name: `正确的，成功`,
			args: args{
				from:       from,
				privateKey: privateKey,
			},
		},
		{
			name: `不匹配的私钥和地址，报错`,
			args: args{
				from:       from,
				privateKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = service.SetUp(tt.args.from, tt.args.privateKey)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_Unfreeze(t *testing.T) {
	TestService_SetUp1(t)

	// 冻结金额太大
	require.Error(t, service.UnFreeze(`TYjBaCYBgngDA3nMpBD76Qk7qBx8twvDqY`))
	require.True(t, beforeUnFreeze)
	require.True(t, afterUnFreeze)
	require.Error(t, afterUnFreezeError)
}

var (
	beforeFreeze       = false
	afterFreeze        = false
	afterFreezeError   error
	beforeUnFreeze     = false
	afterUnFreeze      = false
	afterUnFreezeError error
)

type testHook struct{}

func (t testHook) Key() string {
	return `test`
}

func (t testHook) BeforeFreeze(_ *free.FreezeInfo) {
	beforeFreeze = true
}

func (t testHook) AfterFreeze(_ *free.FreezeInfo, err error) {
	afterFreeze = true
	afterFreezeError = err
}

func (t testHook) BeforeUnfreeze(_ *free.FreezeInfo) {
	beforeUnFreeze = true
}

func (t testHook) AfterUnfreeze(_ *free.FreezeInfo, err error) {
	afterUnFreeze = true
	afterUnFreezeError = err
}
