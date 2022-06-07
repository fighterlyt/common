package capture

import (
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	target Service
)

func TestNewService(t *testing.T) {
	//bg, err := os.Open(`bg.png`)
	//require.NoError(t, err, `打开背景图`)
	var err error
	target, err = NewCaptureService(`images`, 0.95, nil)
	require.NoError(t, err)
}
func TestService_Capture(t *testing.T) {
	TestNewService(t)

	require.NotNil(t, target)

	pic, block, move, err := target.Capture(1200, 543, 100, 100)

	require.NoError(t, err)

	require.NoError(t, output(pic, `pic.png`))
	require.NoError(t, output(block, `block.png`))
	t.Log(move)
}

func output(reader io.ReadCloser, name string) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	var file = "images/" + name
	if has, _ := fileExist(file); has {
		_ = os.Remove(file)
	}
	return ioutil.WriteFile(file, data, fs.ModePerm)
}
func fileExist(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
