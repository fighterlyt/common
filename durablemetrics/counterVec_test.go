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
			args: args{lvs: []string{"label1"}},
		},
		{
			name: "测试写入2",
			args: args{lvs: []string{"test1"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counterVec.WithLabelValuesInc(tt.args.lvs...)
		})
	}

	time.Sleep(time.Minute)
}
