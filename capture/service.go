package capture

import (
	"bytes"
	"github.com/disintegration/imaging"
	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

var (
	ErrNotFound = errors.New(`没有图片`)
)

type Service interface {
	Capture(bgWidth, bgHeight, blockWidth, blockHeight int) (pic, block io.ReadCloser, movement int, err error)
}

type captureService struct {
	Pictures   []image.Image
	PicDirPath string
	Opacity    float64
	Random     *rand.Rand
	Logger     log.Logger
}
type imgBuffer struct {
	data *bytes.Buffer
}

/*Capture 生成一次验证
参数:
*	picWidth   	int             原图宽度
*	picHeight  	int          	原图高度
*	blockWidth 	int          	切块宽度
*	blockHeight	int          	切块大小
返回值:
*	pic        	io.ReadCloser	背景图
*	block      	io.ReadCloser	切块
*	movement   	int          	x轴位移
*	err        	error        	错误
*/
func (c *captureService) Capture(picWidth, picHeight, blockWidth, blockHeight int) (bg, block io.ReadCloser, movement int, err error) {
	var (
		chosenImg image.Image
		blockImg  image.Image
	)

	//1.随机挑选一张图片
	if chosenImg, err = c.findRandomPic(); err != nil {
		return nil, nil, 0, errors.Wrap(err, `未找到图片`)
	}

	// 2. 计算切割小图的大小及位置
	fromY := picHeight/2 - blockHeight/2
	toY := picHeight/2 + blockHeight/2

	movement = c.Random.Intn(picWidth - blockWidth)

	// 3. 切割
	if chosenImg, blockImg, err = c.crop(chosenImg, movement, fromY, movement+blockWidth, toY); err != nil {
		return nil, nil, 0, errors.Wrap(err, `图片切割失败`)
	}

	// 4. 类型转换
	bg = newImgBuffer()
	if err = png.Encode(bg.(io.Writer), chosenImg); err != nil {
		return bg, block, 0, err
	}

	block = newImgBuffer()
	if err = png.Encode(block.(io.Writer), blockImg); err != nil {
		return bg, block, 0, err
	}
	return bg, block, movement, err
}

/*NewCaptureService 构建服务
参数:
*	picDirPath	string    	图片路径
*	logger 	    log.Logger	日志器
*	opacity		float64   	遮罩透明度
返回值:
*	target 	*captureService   	服务
*	err    	error     			错误
*/
func NewCaptureService(picDirPath string, opacity float64, logger log.Logger) (*captureService, error) {
	target := &captureService{
		PicDirPath: picDirPath,
		Opacity:    opacity,
		Logger:     logger,
		Random:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	if err := target.readPngPictures(); err != nil {
		return nil, errors.Wrap(err, "读取png图片失败")
	}

	if len(target.Pictures) <= 0 {
		return nil, ErrNotFound
	}

	return target, nil
}

//readPngPictures 读取目录下的png图片
func (c *captureService) readPngPictures() error {
	var (
		file *os.File
		img  image.Image
	)
	return filepath.WalkDir(c.PicDirPath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		if file, err = os.Open(path); err != nil {
			return errors.Wrap(err, "读取目录下图片失败")
		}
		defer func() { _ = file.Close() }()
		if img, err = png.Decode(file); err != nil {
			return errors.Wrapf(err, `[%s]不是png图片`, path)
		}
		c.Pictures = append(c.Pictures, img)
		return nil
	})
}

//findRandomPic 随机出一张图片
func (c *captureService) findRandomPic() (pic image.Image, err error) {
	index := c.Random.Intn(len(c.Pictures))
	//c.Logger.Info("选用图片：", zap.Int("索引", index))
	return c.Pictures[index], nil
}

//crop 根据随机图片和坐标，裁剪出两张图片
func (c *captureService) crop(from image.Image, fromX, fromY, toX, toY int) (bgPic, blockPic image.Image, err error) {
	bgPic = imaging.Overlay(from, createMaskPng(toX-fromX, toY-fromY), image.Point{
		X: fromX,
		Y: fromY,
	}, c.Opacity)

	blockPic = imaging.Crop(from, image.Rectangle{
		Min: image.Point{
			X: fromX,
			Y: fromY,
		},
		Max: image.Point{
			X: toX,
			Y: toY,
		},
	})

	return bgPic, blockPic, err
}

//createMaskPng 创建遮罩png图片
func createMaskPng(width, height int) image.Image {
	topLeft := image.Point{
		X: 0,
		Y: 0,
	}
	bottomRight := image.Point{
		X: width,
		Y: height,
	}

	var maskPng = image.NewRGBA(image.Rectangle{
		Min: topLeft,
		Max: bottomRight,
	})
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			maskPng.Set(x, y, color.Black)
		}
	}
	return maskPng
}

func (i *imgBuffer) Write(p []byte) (int, error) {
	return i.data.Write(p)
}

func (i *imgBuffer) Read(p []byte) (int, error) {
	return i.data.Read(p)
}

func (i *imgBuffer) Close() error {
	i.data.Reset()
	return nil
}

func newImgBuffer() *imgBuffer {
	return &imgBuffer{
		data: &bytes.Buffer{},
	}
}
