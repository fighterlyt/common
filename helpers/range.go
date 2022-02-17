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
