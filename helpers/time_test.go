package helpers

import (
	"context"
	"testing"
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