package helpers

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type DecimalRangeArgument struct {
	Min decimal.Decimal `json:"min"`
	Max decimal.Decimal `json:"max"`
}

func (d DecimalRangeArgument) Validate() error {
	if d.Min.GreaterThan(decimal.Zero) && d.Max.GreaterThan(decimal.Zero) && d.Min.GreaterThan(d.Max) {
		return errors.New("开始时间必须早于结束时间")
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
