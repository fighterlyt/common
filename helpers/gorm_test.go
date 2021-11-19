package helpers

import (
	"testing"

	"gorm.io/gorm"
)

var (
	db = &gorm.DB{}
)

func TestBuildScope(t *testing.T) {
	type args struct {
		data     interface{}
		excludes []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "纯粹数字",
			args: args{
				data: testInt64Struct{
					A: 10,
					B: 7,
				},
				excludes: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			BuildScope(tt.args.data, tt.args.excludes)

		})
	}
}

type testInt64Struct struct {
	A int `json:"a"`
	B int `json:"b"`
}
