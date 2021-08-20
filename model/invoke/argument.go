package invoke

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

/*ProcessArgument 处理gin 句柄的公共逻辑，包括JSON反序列化和验证
参数:
*	ctx     	*gin.Context	gin上下文
*	argument	Query       	参数
返回值:
*	returned	bool        	是否已经返回
*	err     	error       	是否有错误
*/
func ProcessArgument(ctx *gin.Context, argument Argument) (returned bool, err error) {
	if err = ctx.ShouldBindJSON(argument); err != nil {
		err = errors.Wrap(err, `JSON解析错误`)

		ReturnFail(ctx, Fail, err, `JSON解析错误`)

		return true, err
	}

	if err = argument.Validate(); err != nil {
		err = errors.Wrap(err, `参数校验失败`)

		ReturnFail(ctx, Fail, err, `参数校验错误`)

		return true, err
	}

	return false, nil
}

// Argument  参数
type Argument interface {
	Validate() error
}

type Query interface {
	Argument
	Scope(db *gorm.DB) *gorm.DB
}

// ListArgument 表格查询参数
type ListArgument struct {
	Start       int    `json:"start"` // 起点,从0开始
	Limit       int    `json:"limit"` // 数量上限,0表示不限制
	Query       Query  `json:"query"` // 查询条件
	Sorts       Sorts  `json:"sorts"` // 排序字符串
	DefaultSort string `json:"-"`     // 默认排序
}

func (a ListArgument) Scope(db *gorm.DB) *gorm.DB {
	db = db.Offset(a.Start)
	if a.Limit <= 0 {
		db = db.Limit(math.MaxInt32)
	} else {
		db = db.Limit(a.Limit)
	}
	orderStr := a.Sorts.Order()
	if orderStr != "" {
		db = db.Order(orderStr)
	} else if a.DefaultSort != "" {
		db = db.Order(a.DefaultSort)
	}

	if a.Query != nil {
		a.Query.Scope(db)
	}
	return db
}

func (a ListArgument) Validate() error {
	if a.Start < 0 {
		return fmt.Errorf(`起点[%d]必须大于等于0`, a.Start)
	}

	if a.Limit < 0 {
		return fmt.Errorf(`数量上限[%d]必须大于0`, a.Limit)
	}

	if a.Query != nil {
		if argument, ok := a.Query.(Argument); ok {
			if err := argument.Validate(); err != nil {
				return errors.Wrap(err, `查询条件校验失败`)
			}
		}
	}

	return nil
}

func (a ListArgument) OffSetAndLimit(db *gorm.DB) *gorm.DB {
	db = db.Offset(a.Start)

	if a.Limit <= 0 {
		db = db.Limit(math.MaxInt32)
	} else {
		db = db.Limit(a.Limit)
	}

	return db
}

func NewListArgument(query Query) (argument *ListArgument, err error) {
	if err = IsPointer(query, false); err != nil {
		return nil, errors.Wrap(err, `query不满足要求`)
	}

	return &ListArgument{Query: query}, nil
}

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

type Sorts string

func (s Sorts) Order() string {
	split := strings.Split(string(s), ",")
	to := make([]string, len(split))
	for i := range split {
		to[i] = strings.Trim(split[i], `""`)
		to[i] = strings.TrimSpace(to[i])
		if len(to[i]) > 1 {
			if to[i][:1] == "-" {
				to[i] = to[i][1:] + " desc"
			}
		}
	}
	return strings.Join(to, ",")
}
