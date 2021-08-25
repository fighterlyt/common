package helpers

import "go.uber.org/zap"

/*ZapError 通过zap返回错误
参数:
*	err      	error    	参数1
返回值:
*	zap.Field	zap.Field	返回值1
*/
func ZapError(err error) zap.Field {
	if err == nil {
		return zap.String(`错误`, `无`)
	}

	return zap.String(`错误`, err.Error())
}
