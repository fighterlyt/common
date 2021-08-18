package parameters

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomValid(t *testing.T) {
	type fields struct {
		validFunc func(i, j interface{}) bool
		value     string
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: `key`,
			fields: fields{
				validFunc: keyValid,
				value:     `tron.chargeCheckInterval`,
			},
			wantErr: false,
		},
		{
			name: `usdt正金额5位小数`,
			fields: fields{
				validFunc: usdtPositiveValue,
				value:     `1.23456`,
			},
			wantErr: false,
		},
		{
			name: `usdt正金额6位小数`,
			fields: fields{
				validFunc: usdtPositiveValue,
				value:     `1.234567`,
			},
			wantErr: false,
		},
		{
			name: `usdt正金额7位小数`,
			fields: fields{
				validFunc: usdtPositiveValue,
				value:     `1.2345678`,
			},
			wantErr: true,
		},
		{
			name: `usdt负金额6`,
			fields: fields{
				validFunc: usdtPositiveValue,
				value:     `-1`,
			},
			wantErr: true,
		},
		{
			name: `tronAddresses`,
			fields: fields{
				validFunc: tronAddresses,
				value:     `TBkbH9yKoBtmPtsH5gcxgmn6rWzSREAoUU,TWN9sjAWrUEUmrCoyo3EbrDgno5ye8SyQN`,
			},
			wantErr: false,
		},
		{
			name: `tronAddress`,
			fields: fields{
				validFunc: tronAddresses,
				value:     `TBkbH9yKoBtmPtsH5gcxgmn6rWzSREAoUC`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.EqualValues(t, !tt.wantErr, tt.fields.validFunc(tt.fields.value, nil), `[%s] 验证[%s]`, tt.name, tt.fields.value)
		})
	}
}
