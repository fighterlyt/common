package metrics

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

/*StructToMetrics 将结构中的字段生成对应的metrics,name为 prefix+前缀+字段名,对于内嵌struct而言,会添加类型名作为中间字段.目前只处理整型和浮点型
参数:
*	data  	    interface{}	参数,只能为结构或者指向结构的指针
*	prefix	    string     	前缀
*   delimiter   string      分隔符
返回值:
*	error 	error      	返回值1
*/
func StructToMetrics(data interface{}, prefix, delimiter string) error {
	if data == nil {
		return errors.New(`参数不能为nil`)
	}

	if delimiter == `` {
		return errors.New(`分隔符不能为空`)
	}

	value := reflect.ValueOf(data)

	return buildStructMetrics(value, prefix, delimiter)
}

func buildStructMetrics(value reflect.Value, prefix, delimiter string) error {
	if value.Kind() == reflect.Ptr && value.IsNil() {
		return nil
	}

	t := value.Type()

	switch t.Kind() {
	case reflect.Struct:
	case reflect.Ptr:
		t = t.Elem()
		value = value.Elem()

		if t.Kind() != reflect.Struct {
			return fmt.Errorf(`指针并非指向struct{},而是[%s]`, t.Kind().String())
		}
	default:
		return fmt.Errorf(`只接受struct 和指向 struct的指针,实际类型是[%s]`, t.Kind().String())
	}

	fieldValue := float64(0)

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		matched := true

		key := t.Field(i).Name

		switch field.Kind() {
		case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
			fieldValue = float64(field.Int())
		case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
			fieldValue = float64(field.Uint())
		case reflect.Float64, reflect.Float32:
			fieldValue = field.Float()
		case reflect.Struct:
			subPrefix := strings.Join([]string{prefix, key}, delimiter)
			if err := buildStructMetrics(field, subPrefix, delimiter); err != nil {
				return errors.Wrapf(err, `构建子struct[%s]`, subPrefix)
			}
		case reflect.Ptr:
			if field.Elem().Kind() == reflect.Struct {
				subPrefix := strings.Join([]string{prefix, key}, delimiter)
				if err := buildStructMetrics(field.Elem(), subPrefix, delimiter); err != nil {
					return errors.Wrapf(err, `构建子struct[%s]`, subPrefix)
				}
			}
		default:
			matched = false
		}

		if matched {
			promauto.NewGauge(prometheus.GaugeOpts{
				Name: fmt.Sprintf(`%s:%s`, prefix, key),
			}).Set(fieldValue)
		}
	}

	return nil
}
