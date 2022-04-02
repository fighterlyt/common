package helpers

import (
	"bytes"
	"encoding/json"
	"image"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/owner888/resize"
	"github.com/pkg/errors"
)

type MergeParam struct {
	Width         int  `json:"width"`          // 宽度
	Height        int  `json:"height"`         // 高度
	Distance      int  `json:"distance"`       // 间隔
	Vertical      bool `json:"vertical"`       // 合并方向(是否垂直）
	FirstDistance bool `json:"first_distance"` // 第一幅图前，是否需要间隔
	LastDistance  bool `json:"last_distance"`  // 最后一幅图后，是否需要间隔
}

func (m MergeParam) Key() string {
	data, _ := json.Marshal(m)

	return string(data)
}

/*MergeImages 合并图形
参数:
*	images     	[]image.Image	图形
*   bg          image.Image     背景图，可以为空
*	width      	int          	宽度
*	height     	int          	高度
*	distance   	int          	间隔
*	vertical   	bool         	是否垂直
返回值:
*	*image.RGBA	*image.RGBA  	结果
*	error      	error        	错误
*/
func MergeImages(images []image.Image, bg image.Image, param *MergeParam) (*image.RGBA, error) {
	var (
		finalHeight int
		finalWidth  int
		rect        image.Rectangle
	)

	width := param.Width
	height := param.Height
	distance := param.Distance

	finalWidth, finalHeight = mergeCalFinalWidthHeight(param, len(images))

	des := image.NewRGBA(image.Rect(0, 0, finalWidth, finalHeight)) // 底板

	if bg != nil {
		bg = resize.Resize(uint(width), uint(height), bg, resize.Bicubic)
	}

	for i := range images {
		images[i] = resize.Resize(uint(width), uint(height), images[i], resize.Bicubic)

		if param.Vertical {
			rect = image.Rectangle{
				Min: image.Point{X: 0, Y: i * height},
				Max: image.Point{X: width, Y: (i + 1) * height},
			}

			distanceCount := i
			if param.FirstDistance {
				distanceCount += 1
			}

			rect.Min.Y += distanceCount * distance
			rect.Max.Y += distanceCount * distance
		} else {
			rect = image.Rectangle{
				Min: image.Point{X: i * width, Y: 0},
				Max: image.Point{X: (i + 1) * width, Y: height},
			}

			distanceCount := i
			if param.FirstDistance {
				distanceCount += 1
			}

			rect.Min.X += distanceCount * distance
			rect.Max.X += distanceCount * distance
		}

		if bg != nil {
			draw.Draw(des, rect, bg, image.Pt(0, 0), draw.Src)
		}

		draw.Draw(des, rect, images[i], image.Pt(0, 0), draw.Over)
	}

	return des, nil
}

func mergeCalFinalWidthHeight(param *MergeParam, imageCount int) (width int, height int) {
	width, height = param.Width, param.Height

	if param.Vertical { // 垂直
		height = param.Height*imageCount + param.Distance*(imageCount-1)

		if param.FirstDistance && imageCount >= 1 {
			height += param.Distance
		}

		if param.LastDistance && imageCount > 1 {
			height += param.Distance
		}

		return width, height
	}

	width = param.Width*imageCount + param.Distance*(imageCount-1)

	if param.FirstDistance && imageCount >= 1 {
		width += param.Distance
	}

	if param.LastDistance && imageCount > 1 {
		width += param.Distance
	}

	return width, height
}

func IsImageSpecificType(reader io.Reader, decodeFunc func(reader2 io.Reader) (image.Image, error)) bool {
	if _, err := decodeFunc(reader); err != nil {
		return false
	}

	return true
}

/*IsPNG 判断是否是png
参数:
*	reader	io.Reader	参数1
返回值:
*	bool  	bool     	返回值1
*/
func IsPNG(reader io.Reader) bool {
	return IsImageSpecificType(reader, png.Decode)
}

/*IsImageRadioMatch 判断图片比例是否正确，比如宽高比为3:4,那么就应该传入IsImageRadio(x,x,3,4)
参数:
*	reader    	io.Reader                           	图片
*	decodeFunc	func(io.Reader) (image.Image, error)	解析器
*	width     	int                                 	宽度比例
*	height    	int                                 	高度比例
返回值:
*	bool      	bool                                	是否吻合
*/
func IsImageRadioMatch(reader io.Reader, decodeFunc func(io.Reader) (image.Image, error), width, height int) (match bool, actualWidth, actualHeight int) {
	var (
		img image.Image
		err error
	)

	if img, err = decodeFunc(reader); err != nil {
		return false, 0, 0
	}

	bounds := img.Bounds()

	return bounds.Max.X*height == bounds.Max.Y*width, bounds.Max.X, bounds.Max.Y
}

/*IsPNGRadioMatch 判断png 比例，具体查看IsImageRadioMatch
参数:
*	reader	io.Reader	参数1
*	width 	int      	参数2
*	height	int      	参数3
返回值:
*	bool  	bool     	返回值1
*/
func IsPNGRadioMatch(reader io.Reader, width, height int) (match bool, actualWidth, actualHeight int) {
	return IsImageRadioMatch(reader, png.Decode, width, height)
}

/*DownloadAndOpenAsType 下载并且以指定文件格式(打开
参数:
*	imageURL  	    string                                     	下载路径
*	validateFunc	func(reader io.Reader)  error	            解析器
*   client          *http.Client                                http客户端
返回值:
*   img             image.Image                                 图片
*	reader    	    io.Reader                                  	数据
*	err       	    error                                      	错误
*/
func DownloadAndOpenAsType(imageURL string, validateFunc func(reader io.Reader) (image.Image, error), client *http.Client) (img image.Image, reader io.Reader, err error) { //nolint:lll
	var (
		resp   *http.Response
		buffer []byte
	)

	if client == nil {
		client = &http.Client{}
	}

	if resp, err = client.Get(imageURL); err != nil {
		return nil, nil, errors.Wrap(err, `下载文件`)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if buffer, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, nil, errors.Wrap(err, `读取文件`)
	}

	if img, err = validateFunc(bytes.NewBuffer(buffer)); err != nil {
		return nil, nil, errors.Wrap(err, `解析`)
	}

	return img, bytes.NewBuffer(buffer), nil
}
