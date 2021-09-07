package twofactor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	c   Auth
	err error
)

func TestNewConfig(t *testing.T) {
	c, err = NewAuth(`12345678901234567890155`)
	require.NoError(t, err)
}

func TestConfig_QR(t *testing.T) {
	TestNewConfig(t)

	qr, data, err := c.QR(`admin`)
	require.NoError(t, err)
	t.Log(`qr`, qr, data)
}

func TestConfig_Validate(t *testing.T) {
	TestNewConfig(t)

	ok, err := c.Validate(`675052`)
	require.NoError(t, err)

	require.True(t, ok)
}
