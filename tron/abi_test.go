package tron

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	// "github.com/fighterlyt/gotron-sdk/pkg/common"
	"github.com/stretchr/testify/require"
)

var (
	abi = Trc20Abi{}
)

func TestTrc20Abi_UnpackTransfer(t *testing.T) {
	datas := []string{
		// "a9059cbb00000000000000000000000006f68705166a03d60f103703bed0d87a71571048000000000000000000000000000000000000000000000000000000000f1c1a60", "a9059cbb00000000000000000000004179309abcff2cf531070ca9222a1f72c4a513687400000000000000000000000000000000000000000000000000000000047868c0", //nolint:golint,lll
		`a9059cbb000000000000000000000000bb58f5af2aa510f048819c53fe7de47bbf60dded000000000000000000000000000000000000000000000000000000000ddbab20eba2ab18bfff44eeabee630d05479266`,
	}
	for _, data := range datas {
		to, value, err := abi.UnpackTransfer(data)
		require.NoError(t, err, "unpack")
		t.Log(to)
		t.Log(value)
	}
}

func TestBytes2Hex(t *testing.T) {
	t.Log(common.Bytes2Hex([]byte(`TT3ovxdKLaodgS63GACZ3vvtpWyE3bgsq7`)))
}
