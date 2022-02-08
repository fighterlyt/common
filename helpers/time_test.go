package helpers

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDataCal(t *testing.T) {
	type args struct {
		date int
		add  int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: `加负值`,
			args: args{
				date: 20210918,
				add:  -1,
			},
			want: 20210917,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DataCal(tt.args.date, tt.args.add); got != tt.want {
				t.Errorf("DataCal() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	bg = context.Background()
)

func TestBeginningOfDay(t *testing.T) {
	tests := []struct {
		name string
		want time.Time
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BeginningOfDay(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeginningOfDay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRange_UnMarshal(t *testing.T) {
	data := `{
	"a":{
		"min":1,
		"max":2
	},
	"b":{
		"min":3,
		"max":4
	}
	}
`
	type temp struct {
		A Range `json:"a"`
		B Range `json:"b"`
	}

	result := &temp{}

	require.NoError(t, json.Unmarshal([]byte(data), result))

	require.EqualValues(t, 1, result.A.Min)
	require.EqualValues(t, 2, result.A.Max)

	require.EqualValues(t, 3, result.B.Min)
	require.EqualValues(t, 4, result.B.Max)

}
