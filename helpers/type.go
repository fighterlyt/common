package helpers

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

/*IsPointer 判断是否为指针
参数:
*	data    	interface{}	数据
*	allowNil	bool       	是否允许为空指针
返回值:
*	error   	error      	错误
*/
func IsPointer(data interface{}, allowNil bool) error {
	if data == nil {
		return errors.New(`data不能为nil`)
	}

	if kind := reflect.TypeOf(data).Kind(); kind != reflect.Ptr {
		return fmt.Errorf(`参数类型不是指针，而是[%s]`, kind.String())
	}

	if !allowNil {
		if reflect.ValueOf(data).IsNil() {
			return errors.New(`data 不能是nil指针`)
		}
	}

	return nil
}

/*IsSlice 判断是否为切片
参数:
*	data    	interface{}	待验证数据
*	allowNil	bool       	是否允许空切片
返回值:
*	error   	error      	错误
*/
func IsSlice(data interface{}, allowNil bool) error {
	if data == nil {
		if !allowNil {
			return errors.New(`data 不能为nil`)
		}

		return nil
	}

	if kind := reflect.TypeOf(data).Kind(); kind != reflect.Slice {
		return fmt.Errorf(`类型不是slice,而是[%s]`, kind.String())
	}

	return nil
}
