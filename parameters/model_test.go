package parameters

import (
	"testing"
)

func TestParameter_Validate(t *testing.T) {
	type fields struct {
		Key         string
		Purpose     string
		Value       string
		Description string
		UpdateTime  int64
		ValidKey    string
		before      func()
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: `key为空`,
			fields: fields{
				Key:         "",
				Purpose:     "a",
				Value:       "a",
				Description: "a",
				UpdateTime:  0,
				ValidKey:    "email",
			},
			wantErr: true,
		},
		{
			name: `key长度非法`,
			fields: fields{
				Key:         "a.abc",
				Purpose:     "a",
				Value:       "a",
				Description: "a",
				UpdateTime:  0,
				ValidKey:    "email",
			},
			wantErr: true,
		},
		{
			name: `key长度正好`,
			fields: fields{
				Key:         "a.cdefghijklmnopqrstuvwxyzabcdef",
				Purpose:     "a",
				Value:       "a",
				Description: "a",
				UpdateTime:  0,
				ValidKey:    "email",
			},
			wantErr: false,
		},
		{
			name: `key包含数字`,
			fields: fields{
				Key:         "k8s",
				Purpose:     "a",
				Value:       "a",
				Description: "a",
				UpdateTime:  0,
				ValidKey:    "email",
			},
			wantErr: true,
		},
		{
			name: `purpose为空`,
			fields: fields{
				Key:         "a.cdefghijklmnopqrstuvwxyzabcdef",
				Purpose:     "",
				Value:       "a",
				Description: "a",
				UpdateTime:  0,
				ValidKey:    "email",
			},
			wantErr: true,
		},
		{
			name: `validKey为空`,
			fields: fields{
				Key:         "a.cdefghijklmnopqrstuvwxyzabcdef",
				Purpose:     "a",
				Value:       "a",
				Description: "a",
				UpdateTime:  0,
				ValidKey:    "",
			},
			wantErr: true,
		},
		{
			name: `validKey不存在`,
			fields: fields{
				Key:         "a.defghijklmnopqrstuvwxyzabcdef",
				Purpose:     "a",
				Value:       "a",
				Description: "a",
				UpdateTime:  0,
				ValidKey:    "alpha1",
			},
			wantErr: true,
		},
		{
			name: `值无法通过验证`,
			fields: fields{
				Key:         "a.cdefghijklmnopqrstuvwxyzabcdef",
				Purpose:     "a",
				Value:       "a",
				Description: "a",
				UpdateTime:  0,
				ValidKey:    "email",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Parameter{
				Key:         tt.fields.Key,
				Purpose:     tt.fields.Purpose,
				Value:       tt.fields.Value,
				Description: tt.fields.Description,
				UpdateTime:  tt.fields.UpdateTime,
				ValidKey:    tt.fields.ValidKey,
			}

			if tt.fields.before != nil {
				tt.fields.before()
			}

			if err := p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
