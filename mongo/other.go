package mongo

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func IsInsertDuplicateError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "E11000 duplicate key error collection")
}

/*prettyBsonM 将bson.M转为字符串展示
参数:
*	value	bson.M
返回值:
*	string	string
*/
func prettyBsonM(value bson.M) string {
	builder := &strings.Builder{}
	_, _ = fmt.Fprint(builder, "{")
	i := 0

	for k, v := range value {
		if i != 0 {
			_, _ = fmt.Fprintf(builder, ",")
		}

		if subValue, ok := v.(bson.M); ok {
			_, _ = fmt.Fprintf(builder, "%s:%s", k, prettyBsonM(subValue))
		} else {
			if _, ok = v.([]bson.M); ok {
				subValues := v.([]bson.M)
				subValueStrings := make([]string, 0, len(subValues))
				for _, subValue = range subValues {
					subValueStrings = append(subValueStrings, prettyBsonM(subValue))
				}
				_, _ = fmt.Fprintf(builder, "%s:%s", k, strings.Join(subValueStrings, ","))
			} else {
				_, _ = fmt.Fprintf(builder, "%s:%v", k, v)
			}
		}

		i++
	}

	_, _ = fmt.Fprint(builder, "}")

	return builder.String()
}

/*CombineBsonM 合并多个Bson.M,如果有重复的字段，那么后出现会覆盖前者，返回的值都是经过复制的
参数:
*	documents	...bson.M
返回值:
*	bson.M	bson.M
*/
func CombineBsonM(documents ...bson.M) bson.M {
	count := 0
	for _, document := range documents {
		count += len(document)
	}

	result := make(bson.M, count)

	for _, document := range documents {
		for k, v := range document {
			result[k] = v
		}
	}

	return result
}

/*ZapBsonM 将bson.M转为更加合适的格式，输出到uber.zap
参数:
*	key  	string
*	value	bson.M
返回值:
*	zap.Field	zap.Field
*/
func ZapBsonM(key string, value bson.M) zap.Field {
	return zap.Field{
		Key:    key,
		Type:   zapcore.StringType,
		String: prettyBsonM(value),
	}
}
