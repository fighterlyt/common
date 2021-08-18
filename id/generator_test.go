package id

import (
	"testing"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/require"
)

var (
	err           error
	testGenerator Generator
	testID        snowflake.ID
)

func TestNewGenerator(t *testing.T) {
	type args struct {
		id int64
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    `0`,
			args:    args{id: 0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testGenerator, err = NewGenerator(tt.args.id)
			if tt.wantErr {
				require.Error(t, err, tt.name)
			} else {
				require.NoError(t, err, tt.name)
			}
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	TestNewGenerator(t)

	testID = testGenerator.Generate()

	t.Log(`原生`, testID)
	t.Log(`64进制`, testID.Base64())
	t.Log(`36进制`, testID.Base32())
}
