package public

import "strings"

/*StringChang 将阿里云oss图片地址替换为cdn地址
参数:
*	originStr	string	参数1
返回值:
*	newStr  	string  返回值1
*/

func StringChange(originStr string) string {
	aliUrl := "https://dubai-real.oss-accelerate-overseas.aliyuncs.com/"
	cdnUrl := "https://d.khols8.com/"
	newStr := strings.Replace(originStr, aliUrl, cdnUrl, 1)
	return newStr
}

