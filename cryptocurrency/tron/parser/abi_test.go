package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrc20Abi_UnpackTransfer(t *testing.T) {
	datas := []string{
		"a9059cbb00000000000000000000000006f68705166a03d60f103703bed0d87a71571048000000000000000000000000000000000000000000000000000000000f1c1a60", "a9059cbb00000000000000000000004179309abcff2cf531070ca9222a1f72c4a513687400000000000000000000000000000000000000000000000000000000047868c0", //nolint:golint,lll
	}
	for _, data := range datas {
		to, value, err := abi.UnpackTransfer(data)
		require.NoError(t, err, "unpack")
		t.Log(to)
		t.Log(value)
	}
}
