package helpers

import "net/url"

/*IsURL 判断是否是合法的绝对URL
参数:
*	str 	string	URL
返回值:
*	bool	bool  	非法
*/
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
