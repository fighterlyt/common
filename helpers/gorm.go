package helpers

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Scope mysql查询时的限制条件
type Scope func(db *gorm.DB) *gorm.DB

/*ClearAll 全部删除
参数:
*	db    	*gorm.DB              	db
*	models	map[string]interface{}  待计数的表,key是描述,value 作为gorm.DB.Model() 参数
返回值:
*	err   	error                 	错误
*/
func ClearAll(db *gorm.DB, models map[string]interface{}) error {
	for desc, model := range models {
		if err := db.Unscoped().Where("1 = 1").Delete(model).Error; err != nil {
			return errors.Wrapf(err, `清理[%s]失败`, desc)
		}
	}

	return nil
}

/*CountAll 全部计数
参数:
*	db    	*gorm.DB              	db
*	models	map[string]interface{}  待计数的表,key是描述,value 作为gorm.DB.Model() 参数
返回值:
*	counts	map[string]int64      	数量,key和models的key相同
*	err   	error                 	错误
*/
func CountAll(db *gorm.DB, models map[string]interface{}) (counts map[string]int64, err error) {
	counts = make(map[string]int64, len(models))

	for desc, model := range models {
		count := counts[desc]
		if err := db.Model(model).Count(&count).Error; err != nil {
			return nil, errors.Wrapf(err, `计数[%s]失败`, desc)
		}

		counts[desc] = count
	}

	return counts, nil
}

const (
	minNotZero = 2 << iota // 最小值不是空
	minIsZero              // 最小值为空

	maxNotZero // 最大值不是空
	maxIsZero  // 最大值为空
)

type interval interface {
	IsZero() bool
}

/*IntervalScope 设置对应db字段在数据库中的范围
参数:
*	dbField	string  	db字段
*	min    	interval	区间左值
*	max    	interval	区间右值
返回值:
*	[]Scope	[]Scope 	数据库条件
*/
func IntervalScope(dbField string, min, max interval) []Scope {
	var scopes []Scope

	b := bitmapper{
		bit: 0,
		min: min,
		max: max,
	}

	switch {
	case b.bitmapIs(minNotZero | maxIsZero):
		scopes = append(scopes, func(db *gorm.DB) *gorm.DB {
			return db.Where(dbField+" >= ?", min)
		})
	case b.bitmapIs(minIsZero | maxNotZero):
		scopes = append(scopes, func(db *gorm.DB) *gorm.DB {
			return db.Where(dbField+" <= ?", max)
		})
	case b.bitmapIs(minNotZero | maxNotZero):
		scopes = append(scopes, func(db *gorm.DB) *gorm.DB {
			return db.Where(dbField+" BETWEEN ? AND ?", min, max)
		})
	default:
	}

	return scopes
}

type bitmapper struct {
	bit int
	min interval
	max interval
}

// bitmap 初始化位图
func (b *bitmapper) bitmap() int {
	if b.bit == 0 {
		if b.min.IsZero() {
			b.bit |= minIsZero
		} else {
			b.bit |= minNotZero
		}

		if b.max.IsZero() {
			b.bit |= maxIsZero
		} else {
			b.bit |= maxNotZero
		}
	}

	return b.bit
}

// bitmapIs 位图判断
func (b *bitmapper) bitmapIs(bit int) bool {
	return b.bitmap()&bit == bit
}

func BuildScope(data interface{}, excludes []string) Scope {
	v := reflect.ValueOf(data)
	t := v.Type()
	tag := `json`

	excludeDict := make(map[string]struct{}, len(excludes))
	for _, exclude := range excludes {
		excludeDict[exclude] = struct{}{}
	}

	return func(db *gorm.DB) *gorm.DB {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			name := t.Field(i).Tag.Get(tag)

			if _, exist := excludeDict[name]; exist { // 排除的字段
				continue
			}

			operator := ``

			switch field.Type().Kind() {
			case reflect.Int64, timeKind:
				if field.Int() > 0 {
					if strings.HasSuffix(name, `min`) {
						operator = `>=`
					} else if strings.HasSuffix(name, `max`) {
						operator = `<=`
					}

					db = db.Where(compose(name, operator), field)
				}
			case reflect.String:
				if field.Len() > 0 {
					db = db.Where(compose(name, operator), field)
				}
			default:
			}
		}

		return db
	}
}

func compose(name, operator string) string {
	if operator == `` {
		operator = `=`
	}

	return fmt.Sprintf(`%s %s ?`, name, operator)
}

var (
	t        = Now()
	timeKind = reflect.TypeOf(t).Kind()
)

// JSONMap 定义的map[string]string,底层使用mysql的json
type JSONMap map[string]string

/*GetWithAlternativeKey 根据key获取值,如果该key对应的值为空且alterKey 不为空字符串(非alterKey的值)，那么返回alterKey的值
参数:
*	key       	string	key
*	alterKey	string	可替换key
返回值:
*	string    	string	值
*/
func (m JSONMap) GetWithAlternativeKey(key, alterKey string) string {
	if m.Get(key) == `` && alterKey != `` {
		return m.Get(alterKey)
	}
	return m.Get(key)
}

/*Get 根据key获取值
参数:
*	key       	string	key值
返回值:
*	string    	string 值
*/
func (m JSONMap) Get(key string) string {
	return m[key]
}

/*Set 设置key/value
参数:
*	key  	string	键值
*	value	string  值
返回值:
*/
func (m JSONMap) Set(key, value string) {
	m[key] = value
}

// Value 返回了mysql存储的真实类型
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}

	ba, err := m.MarshalJSON()

	return string(ba), err
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (m *JSONMap) Scan(val interface{}) error {
	if val == nil {
		*m = make(JSONMap)
		return nil
	}

	var ba []byte

	switch v := val.(type) {
	case []byte:
		ba = v
	case string:
		ba = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", val))
	}

	t := map[string]string{}

	err := json.Unmarshal(ba, &t)

	*m = t

	return err
}

// MarshalJSON to output non base64 encoded []byte
func (m JSONMap) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}

	t := map[string]string(m)

	return json.Marshal(t)
}

// UnmarshalJSON to deserialize []byte
func (m *JSONMap) UnmarshalJSON(b []byte) error {
	t := map[string]string{}
	err := json.Unmarshal(b, &t)
	*m = t

	return err
}

// GormDataType gorm common data type
func (m JSONMap) GormDataType() string {
	return "jsonmap"
}

// GormDBDataType gorm db data type
func (m JSONMap) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	}

	return ""
}

func (m JSONMap) GormValue(_ context.Context, db *gorm.DB) clause.Expr {
	data, _ := m.MarshalJSON()

	switch db.Dialector.Name() {
	case "mysql":
		if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
			return gorm.Expr("CAST(? AS JSON)", string(data))
		}
	}

	return gorm.Expr("?", string(data))
}

func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), `Error 1062: Duplicate entry`)
}

// Amount 专门用于sum(*)的结果
type Amount struct {
	Data decimal.Decimal `gorm:"column:data"`
}
