package captcha

import (
	"github.com/pkg/errors"
	"gitlab.com/nova_dubai/cache"
	"io"
)

var (
	imagePath string
	images    []string
)

type Service struct {
	cacheClient *cache.Client
}
type ImgParam struct {
	Cid      string    `json:"cid"`
	Width    int       `json:"width"`
	Height   int       `json:"height"`
	BgImg    io.Reader `json:"bgImg"`
	FrontImg io.Reader `json:"frontImg"`
}

func NewService(imgPath string, cacheClient cache.Client) error {
	if len(imgPath) <= 0 {
		return errors.New("请设置图片存储路径")
	}
	imagePath = imgPath

	return nil
}

/*GetCaptcha 获取一个图形验证码
参数:
*	width      	int          	宽度
*	height     	int          	高度
返回值:
*	resp	*image.RGBA  	结果
*	err      	error        	错误
*/
func (s *Service) GetCaptcha(width int, height int) (resp *ImgParam, err error) {
	return nil, nil
}

func (s *Service) Check(cid string) (b bool, err error) {
	return true, nil
}
