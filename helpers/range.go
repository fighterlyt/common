package helpers

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type DecimalRangeArgument struct {
	Min   decimal.Decimal `json:"min"`
	Max   decimal.Decimal `json:"max"`
	Filed string          `json:"-"` // 属性名称,用于返回错误信息
}

func (d DecimalRangeArgument) Validate() error {
	if d.Filed == "" {
		return errors.New("请先设置属性名称")
	}

	if d.Min.GreaterThan(decimal.Zero) && d.Max.GreaterThan(decimal.Zero) && d.Min.GreaterThan(d.Max) {
		return fmt.Errorf("%s金额上限不能低于%s金额下限", d.Filed, d.Filed)
	}

	return nil
}

func (d DecimalRangeArgument) Scope(field string) Scope {
	return func(db *gorm.DB) *gorm.DB {
		if d.Min.GreaterThan(decimal.Zero) {
			db = db.Where(fmt.Sprintf(`%s >= ?`, field), d.Min)
		}

		if d.Max.GreaterThan(decimal.Zero) {
			db = db.Where(fmt.Sprintf(`%s <= ?`, field), d.Max)
		}

		return db
	}
}

// Int64RangeArgument 整数间隔
type Int64RangeArgument struct {
	Min   int64  `json:"min"`
	Max   int64  `json:"max"`
	Field string `json:"-"` // 属性名称,用于返回错误信息
}

/*NewInt64RangeArgument 新建Int64整数间隔参数
参数:
*	field              	string             	参数1
返回值:
*	*Int64RangeArgument	*Int64RangeArgument	返回值1
*/
func NewInt64RangeArgument(field string) *Int64RangeArgument {
	return &Int64RangeArgument{Field: field}
}

func (d Int64RangeArgument) Validate() error {
	if d.Field == "" {
		return errors.New("请先设置属性名称")
	}

	if d.Min > 0 && d.Max > 0 && d.Min > d.Max {
		return fmt.Errorf("%s上限不能低于%s下限", d.Field, d.Field)
	}

	return nil
}

func (d Int64RangeArgument) Scope() Scope {
	return func(db *gorm.DB) *gorm.DB {
		if d.Min > 0 {
			db = db.Where(fmt.Sprintf(`%s >= ?`, d.Field), d.Min)
		}

		if d.Max > 0 {
			db = db.Where(fmt.Sprintf(`%s <= ?`, d.Field), d.Max)
		}

		return db
	}
}
