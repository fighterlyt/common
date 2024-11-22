package yunpian

import (
	"sync"
	"testing"
	"time"

	"github.com/fighterlyt/common/sms"
	"github.com/fighterlyt/log"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	service sms.Service
)

func TestNewService(t *testing.T) {
	debug = true
	logger, _ := log.NewEasyLogger(true, false, ``, `测试`)
	service = NewService(`267ae9382419204c77ec80d2cce15301`, time.Second*2, logger, time.Second, newAccessor(), 3)
}

func TestService_DirectSend(t *testing.T) {
	TestNewService(t)

	type args struct {
		in0 string
		in1 string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: `不支持支持发送`,
			args: args{
				in0: "971585119862",
				in1: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DirectSend(tt.args.in0, tt.args.in1)
			require.Error(t, err)
			t.Log(err.Error())
		})
	}
}

func TestService_TemplateSend(t *testing.T) {
	TestNewService(t)

	type args struct {
		target  string
		content string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: `迪拜手机`,
			args: args{
				target:  "971-585119862",
				content: `[Macy's] Your verification code is:123456, verification code 2 minutes valid, please do not disclose to others`,
			},
		},
		{
			name: "非法手机号",
			args: args{
				target:  "",
				content: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.TemplateSend(tt.args.target, tt.args.content, primitive.NewObjectID().Hex())
			require.NoError(t, err)
		})
	}
}

func TestService_Balance(t *testing.T) {
	TestNewService(t)

	var (
		balance decimal.Decimal
		err     error
	)

	tests := []struct {
		name string
	}{
		{
			name: `迪拜手机`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance, err = service.Balance()
			require.NoError(t, err)
			t.Log(balance.String())
		})
	}
}

type accessor struct {
	record map[string]string
	lock   *sync.RWMutex
}

func (a *accessor) SetFinish(id string, err error) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	var (
		errMsg = ``
	)

	if err != nil {
		errMsg = err.Error()
	}

	a.record[id] = errMsg

	return nil
}

func (a *accessor) GetFinishStatus(id string) (status sms.SendStatus, err error) {
	a.lock.RLock()
	defer a.lock.RUnlock()

	if data, exist := a.record[id]; exist {
		if data == `` {
			return sms.SendSuccess, nil
		}

		return sms.SendFail, nil
	}

	return sms.SendUnknown, nil
}

func newAccessor() *accessor {
	return &accessor{
		record: make(map[string]string, 100),
		lock:   &sync.RWMutex{},
	}
}

func TestServicePullStatus(t *testing.T) {
	TestNewService(t)

	debug = true

	var (
		status *pullStatusResult
		err    error
	)

	status, err = service.(*Service).pullStatus()

	require.NoError(t, err)

	t.Log(status)
}
