package helpers

import "strings"

func ContainsSub(data string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(candidate, data) {
			return true
		}
	}

	return false
}

/*Contains 是否保存，判断数组中是否存在某个字符串
参数:
*	value     	string   	是否存在的值
*	candidates	...string	数组
返回值:
*	bool      	bool     	返回值1
*/
func Contains(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if value == candidate {
			return true
		}
	}

	return false
}

/*IsStringEmpty 判断字符串是否为空
参数:
*	str 	string	参数1
返回值:
*	bool	bool  	返回值1
*/
func IsStringEmpty(str string) bool {
	return strings.TrimSpace(str) == ``
}
