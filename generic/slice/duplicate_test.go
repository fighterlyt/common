package slice

import (
	"testing"
)

func TestRemoveDuplicate(t *testing.T) {
	t.Log(RemoveDuplicate[int]([]int{1, 2, 3, 2, 1}))
	t.Log(RemoveDuplicate[string]([]string{`a`, `b`, `c`, `b`, `a`}))
}
