package parameters

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fighterlyt/common/helpers"
	"github.com/shopspring/decimal"

	"github.com/fighterlyt/log"
	"go.uber.org/zap"
)

const (
	// FieldDelimiter 字段分隔符
	FieldDelimiter = `:`
)

var (
	isDuration = func(i interface{}, o interface{}) bool {
		data, ok := i.(string)
		if !ok {
			return false
		}

		_, err := time.ParseDuration(data)

		return err == nil
	}
	keyValid      = Regex(Delimiter(`\w+`, FieldDelimiter, 2))
	isTronAddress = func(i interface{}, o interface{}) bool {
		data, ok := i.(string)
		if !ok {
			return false
		}
		return helpers.ValidateAddress(data)
	}
	usdtPositiveValue = Regex(regexp.MustCompile(`^\d+(\.\d{1,6})?$`))
	usdtValue         = Regex(regexp.MustCompile(`^-?\d+(\.\d{1,6})?$`))
	tronAddresses     = func(i interface{}, o interface{}) bool {
		data, ok := i.(string)
		if !ok {
			return false
		}

		fields := strings.Split(data, `,`)

		for _, field := range fields {
			if !helpers.ValidateAddress(field) {
				println(`tronAddress`, field)
				return false
			}
		}

		return true
	}

	rate = func(i interface{}, o interface{}) bool {
		rateInfo, err := decimal.NewFromString(i.(string))
		if err != nil {
			moduleLogger.Error("验证参数失败", zap.String(`错误`, err.Error()))

			return false
		}

		if rateInfo.LessThan(decimal.Zero) || rateInfo.GreaterThan(decimal.NewFromInt(100)) {
			moduleLogger.Error("比率必须在0-100之间", zap.String(`实际`, rateInfo.String()))

			return false
		}

		return true
	}

	positiveInteger = func(i interface{}, o interface{}) bool {
		data, ok := i.(string)
		if !ok {
			return false
		}

		result, err := strconv.ParseInt(data, 10, 64)
		if err != nil {
			println(`false`, result)
			return false
		}

		if result <= 0 {
			println(`false`, result)
			return false
		}

		return true
	}

	isBool = func(i interface{}, o interface{}) bool {
		data, ok := i.(string)
		if !ok {
			return false
		}

		if data == "0" || data == "1" {
			return true
		}

		return false
	}

	isString = func(i interface{}, o interface{}) bool {
		_, ok := i.(string)

		return ok
	}
	isAttr = func(i interface{}, o interface{}) bool {
		switch v := i.(type) {
		case int, int8, int16, int32, int64:
			if v == 0 || v == 1 {
				return true
			}
		default:
			return false
		}
		return false
	}

	notifyExpressionRate = func(i interface{}, o interface{}) bool {
		if helpers.IsTest() {
			return true
		}
		data, ok := i.(string)
		if !ok {
			return false
		}
		var temp time.Duration
		for index, field := range strings.Split(data, ",") {
			duration, err := time.ParseDuration(field)
			if err != nil {
				return false
			}
			if index == 0 {
				temp = duration
				continue
			}
			if temp > duration {
				return false
			}
		}
		// 字符串为空是不通知
		return true
	}
)

/*
Regex 使用正则表达式验证
参数:
*	regexExpr       *regexp.Regexp                              正则表达式
返回值:
*	validator       func(i interface{}, o interface{}) bool		govalidator的自定义验证器
*/
func Regex(regexExpr *regexp.Regexp) func(i interface{}, o interface{}) bool {
	return func(i interface{}, _ interface{}) bool {
		data, ok := i.(string)
		if !ok {
			return false
		}

		return regexExpr.MatchString(data)
	}
}

/*
Delimiter 生成一个由delimiter分隔，每个部分内容为content,共有count个content
参数:
*	content       	string        	内容
*	Delimiter     	string        	分隔符
*	count         	int           	内容数量
返回值:
*	*regexp.Regexp	*regexp.Regexp	正则
*/
func Delimiter(content, delimiter string, count int) *regexp.Regexp {
	result := content

	for i := 0; i < count-1; i++ {
		result += delimiter + content
	}

	result = `\b` + result + `\b` // \b表示单词的开始和结尾

	return regexp.MustCompile(result)
}

var (
	moduleLogger log.Logger
)
