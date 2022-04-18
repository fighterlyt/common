package cryptocurrency

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitContractLocator(t *testing.T) {
	InitContractLocator([]string{USDT, FLY}, []string{USDT, SGMT})
}
func Test_mapContractLocator_GetContract(t *testing.T) {
	TestInitContractLocator(t)

	tests := []struct {
		name     string
		args     string
		protocol Protocol
		want     Contract
		wantErr  bool
	}{
		{
			name:     `ETH-生产`,
			protocol: Erc20,
			args:     USDT,
			want:     ContractERC20USDT,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.protocol.ContractLocator().GetContract(tt.args)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
