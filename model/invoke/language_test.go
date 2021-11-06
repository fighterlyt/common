package invoke

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadLanguage(t *testing.T) {
	require.NoError(t, loadLanguage(`.`))
	t.Log(language)
}
