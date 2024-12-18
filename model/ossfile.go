package model

import (
	"github.com/pkg/errors"
	"strings"
)

var (
	fromHost      string // 待替换的域名
	isSetFromHost bool   // 是否设置了待替换的域名
	toHost        string // 替换的域名
	isSetToHost   bool   // 是否设置了替换后的域名
)

/*SetOssFromHost 设置代替换域名
参数:
*	from	string	待替换的域名
返回值:
*/
func SetOssFromHost(from string) {
	fromHost = from
	isSetFromHost = true
}

/*SetOssToHost 设置替换后的域名
参数:
*	to	string	替换的域名
返回值:
*/
func SetOssToHost(to string) {
	toHost = to
	isSetToHost = true
}

// OssFilePath 存储到oss的文件路径
type OssFilePath string

/*MarshalText 序列化方法
参数:
返回值:
*	[]byte	[]byte	序列化后的数据
*	error 	error 	错误
*/
func (o OssFilePath) MarshalText() ([]byte, error) {
	if !isSetFromHost {
		return nil, errors.New("请先设置待替换域名")
	}

	if !isSetToHost {
		return nil, errors.New("请先设置替换后的域名")
	}

	if fromHost == "" || toHost == "" { // 任意一个host没有设置都返回原来的路径
		return []byte(o), nil
	}

	return []byte(strings.ReplaceAll(string(o), fromHost, toHost)), nil
}
