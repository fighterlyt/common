package durablemetrics

import (
	"testing"
	"time"
)

func TestCounter_Inc(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "测试增加"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counter.Inc()
		})
	}

	time.Sleep(time.Minute * 5)
}
