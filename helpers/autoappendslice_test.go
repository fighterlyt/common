package helpers

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

var a *AutoAppendSlice

func TestNewAutoAppendSlice(t *testing.T) {
	a, err = NewAutoAppendSlice(10000, 10000)
	require.NoError(t, err)
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

func TestAutoAppendSlice_SetElement1(t *testing.T) {
	TestNewAutoAppendSlice(t)
	wg := sync.WaitGroup{}

	for i := 1; i <= 100000; i++ {
		wg.Add(1)
		go func(index int) {
			err = a.SetElement(int64(10000+index), fmt.Sprintf("%d", index))
			require.NoError(t, err)

			wg.Done()
		}(i)
	}

	for i := 1; i <= 100000; i++ {
		wg.Add(1)
		go func(index int) {
			_, err = a.GetElement(int64(10000 + index))
			require.NoError(t, err)
			wg.Done()
			// t.Log(index, ":", result)
		}(i)
	}

	wg.Wait()
}
