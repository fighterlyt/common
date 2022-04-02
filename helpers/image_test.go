package helpers

import (
	"bufio"
	"bytes"
	"image"
	"image/png"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeImages(t *testing.T) {
	images := make([]image.Image, 0, 10)

	var (
		err    error
		temp   image.Image
		target *image.RGBA
		bg     image.Image
		file   *os.File
	)

	file, err = os.Open(`./bg_jp.png`)
	require.NoError(t, err)

	defer file.Close()

	bg, err = png.Decode(file)
	require.NoError(t, err)

	err = filepath.WalkDir(`./合并`, func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(d.Name()) != `.png` {
			return nil
		}

		if temp, err = getImage(path); err != nil {
			return err
		}

		require.NoError(t, err)

		images = append(images, temp)

		return nil
	})

	require.NoError(t, err)

	target, err = MergeImages(images, bg, &MergeParam{
		Width:         70,
		Height:        100,
		Distance:      25,
		Vertical:      true,
		FirstDistance: true,
		LastDistance:  true,
	})

	require.NoError(t, err)
	require.NoError(t, outputImage(target))
}

func outputImage(img *image.RGBA) error {
	var (
		outFile *os.File
		err     error
	)

	if outFile, err = os.Create("gopher2.png"); err != nil {
		return err
	}

	defer outFile.Close()

	b := bufio.NewWriter(outFile)
	if err = png.Encode(b, img); err != nil {
		return err
	}

	return b.Flush()
}

func getImage(file string) (temp image.Image, err error) {
	var (
		f *os.File
	)

	if f, err = os.Open(file); err != nil {
		return nil, err
	}

	defer f.Close()

	if temp, err = png.Decode(f); err != nil {
		return nil, err
	}

	return temp, nil
}

func TestIsPNG(t *testing.T) {
	var (
		data []byte
		err  error
	)

	data, err = ioutil.ReadFile(`./gopher2.png`)
	require.NoError(t, err)

	require.True(t, IsPNG(bytes.NewReader(data)))
}

func TestIsPNGFalse(t *testing.T) {
	var (
		data []byte
		err  error
	)

	data, err = ioutil.ReadFile(`./test.go`)
	require.NoError(t, err)

	require.False(t, IsPNG(bytes.NewReader(data)))
}

func TestIsPNGRadioMatch(t *testing.T) {
	var (
		data []byte
		err  error
	)

	data, err = ioutil.ReadFile(`./gopher2.png`)
	require.NoError(t, err)

	match, _, _ := IsPNGRadioMatch(bytes.NewReader(data), 3, 4)
	require.False(t, match)
}

func TestDownloadAndOpenAsType(t *testing.T) {
	type args struct {
		imageURL     string
		validateFunc func(reader io.Reader) (image.Image, error)
	}

	pngValidate := png.Decode

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: `合法png`,
			args: args{
				imageURL:     "https://dubai-real.oss-accelerate-overseas.aliyuncs.com/first/telegram.png",
				validateFunc: pngValidate,
			},
			wantErr: false,
		},
		{
			name: `非png文件`,
			args: args{
				imageURL:     "https://dubai-common.oss-me-east-1.aliyuncs.com/%E5%A4%B4%E5%83%8F/%E7%94%B7%E6%80%A7/10.jpeg",
				validateFunc: pngValidate,
			},
			wantErr: true,
		},
		{
			name: "路径错误",
			args: args{
				imageURL:     "https://dubai-common.oss-me-east-1.aliyuncs.com/%E5%A4%B4%E5%83%8F/%E7%94%B7%E6%80%A7/10.jpe",
				validateFunc: pngValidate,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, _, err := DownloadAndOpenAsType(tt.args.imageURL, tt.args.validateFunc, nil)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, img)
			} else {
				require.NotNil(t, img)
				require.NoError(t, err)
			}
		})
	}
}
