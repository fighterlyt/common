package invoke

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.com/nova_dubai/common/helpers"
)

var (
	ErrFail = errors.New(`操作失败`)
)

// StatCode 状态码
type StatCode int

const (
	// Success 成功
	Success StatCode = 1
	// Fail 失败
	Fail StatCode = 0
	// Unauthorized 未授权
	Unauthorized StatCode = -1
)

// ListResult HTTP表格返回参数
type ListResult struct {
	Total int64       `json:"total"` // 总数量
	Rows  interface{} `json:"rows"`  // 返回值
}

/*NewListResult 新建表格返回值
参数:
*	count  	int64      	总数量
*	records	interface{}	记录
返回值:
*	result 	*ListResult	结果
*	err    	error      	错误
*/
func NewListResult(count int64, records interface{}) (result *ListResult, err error) {
	if err = IsSlice(records, true); err != nil {
		return nil, errors.Wrap(err, `参数必须是slice或者nil`)
	}
	if records == nil {
		records = make([]interface{}, 0)
	} else if reflect.ValueOf(records).IsValid() {
		switch reflect.TypeOf(records).Kind() {
		case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			if reflect.ValueOf(records).IsNil() {
				records = make([]interface{}, 0)
			}

		}
	}
	return &ListResult{Total: count, Rows: records}, nil
}

// Result http返回格式
type Result struct {
	Code   StatCode    `json:"code"`
	Msg    string      `json:"msg"`
	Detail string      `json:"detail"`
	Data   interface{} `json:"data"` // 业务数据
}

func NewSuccessResult(msg string, data interface{}) *Result {
	if msg == `` {
		msg = `操作成功`
	}
	return NewResult(Success, msg, data, ``)
}

/*NewResult 构造返回结构
参数:
*	code       	StatCode   	错误码
*	msg        	string     	描述
*	data       	interface{}	数据
*   details     string      错误详情
返回值:
*	*HTTPResult	*HTTPResult	返回结构
*/
func NewResult(code StatCode, msg string, data interface{}, detail string) *Result {
	return &Result{
		Code:   code,
		Msg:    msg,
		Data:   data,
		Detail: detail,
	}
}

func ReturnFail(ctx *gin.Context, code StatCode, err error, detail string) {
	if err == nil {
		ctx.JSON(http.StatusOK, NewSuccessResult(`操作成功`, nil))
		return
	}

	var msg = err.Error()

	if code == Unauthorized {
		msg = "登陆失效"
		detail = "需要重新登录"
	}

	helpers.CtxError(ctx, err)

	result := NewResult(code, msg, nil, detail)
	ctx.JSON(http.StatusOK, result)
}

func ReturnSuccess(ctx *gin.Context, data interface{}) {
	JSON(ctx, data)
	// ctx.translateJSON(http.StatusOK, NewSuccessResult(``, data))
}

func JSON(ctx *gin.Context, data interface{}) {
	ctx.Render(http.StatusAccepted, translateJSON{Data: NewSuccessResult(``, data), lang: ctx.GetHeader(`lang`)})

}

type Dictionary struct {
	Key   interface{} `json:"key"`   // 键
	Value string      `json:"value"` // 值
}

type DictionarySwagger struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Detail string `json:"detail"`
	Data   struct {
		Count   int          `json:"total"`
		Records []Dictionary `json:"rows"`
	} `json:"data"`
}

type HttpListSwagger struct {
	Code   int        `json:"code"`
	Msg    string     `json:"msg"`
	Detail string     `json:"detail"`
	Data   ListResult `json:"data"`
}

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
