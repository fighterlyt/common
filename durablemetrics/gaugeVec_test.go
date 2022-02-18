package durablemetrics

import (
	"testing"
	"time"
)

func TestGaugeVec_WithLabelValuesAdd(t *testing.T) {
	type args struct {
		f   float64
		lvs []string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试写入",
			args: args{
				f:   123.4,
				lvs: []string{"usdt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gaugeVec.WithLabelValuesAdd(tt.args.f, tt.args.lvs...)
		})
	}

	time.Sleep(time.Minute)
}
