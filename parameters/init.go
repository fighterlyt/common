package parameters

import "github.com/asaskevich/govalidator"

/*Init 初始化
参数:
*	path      	string                                    	配置文件路径
*	validators	map[string]govalidator.CustomTypeValidator	扩展的验证方法
返回值:
*/
func Init(path string, validators map[string]govalidator.CustomTypeValidator) {
	dataPath = path

	for key, validator := range validators {
		govalidator.CustomTypeTagMap.Set(key, validator)
	}

	govalidator.CustomTypeTagMap.Set(`duration`, isDuration)
	govalidator.CustomTypeTagMap.Set(`key`, keyValid)
	govalidator.CustomTypeTagMap.Set(`tronAddress`, isTronAddress)
	govalidator.CustomTypeTagMap.Set(`usdtPositiveValue`, usdtPositiveValue)
	govalidator.CustomTypeTagMap.Set(`usdtValue`, usdtValue)
	govalidator.CustomTypeTagMap.Set(`tronAddresses`, tronAddresses)
	govalidator.CustomTypeTagMap.Set(`rate`, rate)
	govalidator.CustomTypeTagMap.Set(`positiveInteger`, positiveInteger)
	govalidator.CustomTypeTagMap.Set(`isBool`, isBool)
	govalidator.CustomTypeTagMap.Set(`isString`, isString)
	govalidator.CustomTypeTagMap.Set(`isAttr`, isAttr)
}