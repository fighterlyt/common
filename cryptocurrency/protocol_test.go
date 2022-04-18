package cryptocurrency

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSymbol(t *testing.T) {
	type args struct {
		protocol Protocol
		test     bool
	}

	tests := []struct {
		name string
		args args
		want Symbol
	}{
		{
			name: "波场测试链",
			args: args{
				protocol: Trc20,
				test:     false,
			},
			want: USDT,
		},
		{
			name: `波场真实`,
			args: args{
				protocol: Trc20,
				test:     true,
			},
			want: SGMT,
		},
		{
			name: "以太坊测试链",
			args: args{
				protocol: Erc20,
				test:     false,
			},
			want: USDT,
		},
		{
			name: `以太坊真实`,
			args: args{
				protocol: Erc20,
				test:     true,
			},
			want: FLY,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSymbol(tt.args.protocol, tt.args.test)
			require.Equal(t, tt.want, got)
		})
	}
}
