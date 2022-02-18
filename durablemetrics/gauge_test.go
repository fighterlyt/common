package durablemetrics

import (
	"testing"
)

func TestGauge_Set(t *testing.T) {
	type args struct {
		f float64
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试set",
			args: args{f: 1234},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gauge.Set(tt.args.f)
		})
	}
}

func TestGauge_Add(t *testing.T) {
	type args struct {
		f float64
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试add",
			args: args{f: 111},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gauge.Add(tt.args.f)
		})
	}
}
