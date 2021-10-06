package helpers

import "flag"

/*IsTest 是否运行测试文件
参数:
返回值:
*	bool	bool	返回值1
*/
func IsTest() bool {
	return flag.Lookup("test.v") != nil
}
