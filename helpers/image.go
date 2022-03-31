package helpers

import (
	"image"
	"image/draw"
	"image/png"
	"io"

	"github.com/owner888/resize"
)

/*MergeImages 合并图形
参数:
*	images     	[]image.Image	图形
*	width      	int          	宽度
*	height     	int          	高度
*	distance   	int          	间隔
*	vertical   	bool         	是否垂直
返回值:
*	*image.RGBA	*image.RGBA  	结果
*	error      	error        	错误
*/
func MergeImages(images []image.Image, width, height, distance int, vertical bool) (*image.RGBA, error) {
	var (
		finalHeight = height
		finalWidth  = width
		rect        image.Rectangle
	)

	if vertical { // 垂直
		finalHeight = height*len(images) + distance*(len(images)-1)
	} else {
		finalWidth = width*len(images) + distance*(len(images)-1)
	}

	des := image.NewRGBA(image.Rect(0, 0, finalWidth, finalHeight)) // 底板

	for i := range images {
		images[i] = resize.Resize(uint(width), uint(height), images[i], resize.Bicubic)

		// bounds := images[i].Bounds()
		//
		// temp := image.NewRGBA(bounds)
		// draw.Draw(temp, bounds, images[i], image.Pt(0, 0), draw.Over)
		//
		// images[i] = temp.SubImage(image.Rect(0, 0, width, height))

		if vertical {
			rect = image.Rectangle{
				Min: image.Point{X: 0, Y: i * height},
				Max: image.Point{X: width, Y: (i + 1) * height},
			}

			if i > 0 {
				rect.Min.Y += i * distance
				rect.Max.Y += i * distance
			}
		} else {
			rect = image.Rectangle{
				Min: image.Point{X: i * width, Y: 0},
				Max: image.Point{X: (i + 1) * width, Y: height},
			}

			if i > 0 {
				rect.Min.X += i * distance
				rect.Max.X += i * distance
			}
		}

		draw.Draw(des, rect, images[i], image.Pt(0, 0), draw.Src)
	}

	return des, nil
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
func IsImageRadioMatch(reader io.Reader, decodeFunc func(io.Reader) (image.Image, error), width, height int) bool {
	var (
		img image.Image
		err error
	)

	if img, err = decodeFunc(reader); err != nil {
		return false
	}

	bounds := img.Bounds()

	return bounds.Max.X*height == bounds.Max.Y*width
}

/*IsPNGRadioMatch 判断png 比例，具体查看IsImageRadioMatch
参数:
*	reader	io.Reader	参数1
*	width 	int      	参数2
*	height	int      	参数3
返回值:
*	bool  	bool     	返回值1
*/
func IsPNGRadioMatch(reader io.Reader, width, height int) bool {
	return IsImageRadioMatch(reader, png.Decode, width, height)
}
