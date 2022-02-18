package durablemetrics

import (
	"testing"
	"time"
)

func TestCounterVec_WithLabelValues(t *testing.T) {
	type args struct {
		lvs []string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试写入",
			args: args{lvs: []string{`url1`, `method1`, `sendOK1`, `statusCode1`, `local1`}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counterVec.WithLabelValuesInc(tt.args.lvs...)
		})
	}

	time.Sleep(time.Minute)
}

func TestCounterVec_WithLabelValuesAdd(t *testing.T) {
	type args struct {
		value float64
		lvs   []string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试add",
			args: args{
				value: 12,
				lvs:   []string{`url1`, `method1`, `sendOK1`, `statusCode1`, `local1`},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counterVec.WithLabelValuesAdd(tt.args.value, tt.args.lvs...)
		})
	}

	time.Sleep(time.Minute)
}
