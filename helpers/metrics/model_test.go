package metrics

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestNet_RUN(t *testing.T) {
	n := NewNet(21453)

	require.NotNil(t, n.pid)

	n.Run(bg)

	require.NoError(t, n.err)
	spew.Dump(n.data, n.pid)
}

func TestProcess_Run(t *testing.T) {
	p := NewProcess(459800)

	p.Run(bg)

	require.NoError(t, p.err)

	spew.Dump(p.rlimits)
}
