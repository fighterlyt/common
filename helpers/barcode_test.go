package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBase64BarCode(t *testing.T) {
	result, err := Base64BarCode(`操你`, 100, 100)
	require.NoError(t, err)
	t.Log(result)
}
