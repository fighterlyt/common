package helpers

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var a *AutoAppendSlice

func TestNewAutoAppendSlice(t *testing.T) {
	a = NewAutoAppendSlice(10000, 10000)
}

func TestAutoAppendSlice_GetElement(t *testing.T) {
	TestNewAutoAppendSlice(t)

	result, err := a.GetElement(10001)
	require.NoError(t, err)

	t.Log(result)
}

func TestAutoAppendSlice_SetElement(t *testing.T) {
	TestAutoAppendSlice_GetElement(t)

	err = a.SetElement(10001, "abc")
	require.NoError(t, err)

	result, err := a.GetElement(10001)
	require.NoError(t, err)

	t.Log(result)

	err = a.SetElement(10001, "123")
	require.NoError(t, err)

	result, err = a.GetElement(10001)
	require.NoError(t, err)

	t.Log(result)

}
