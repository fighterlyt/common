package summaryextend

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type Slot string

const (
	// SlotDay 按天汇总
	SlotDay Slot = `天`
	// SlotMonth 按月汇总
	SlotMonth Slot = `月`
	// SlotWhole 总汇总
	SlotWhole Slot = `开天辟地到地老天荒`
)

// Detail 真实数据
type Detail struct {
	ID        int64           `gorm:"column:id;primary_key;column:id;type:bigint(20) unsigned AUTO_INCREMENT;not null;comment:ID" json:"id"`
	Slot      Slot            `gorm:"column:slot;type:varchar(128);comment:槽位类型" json:"slot"`                                                                   //nolint:lll    // 槽位类型
	OwnerID   string          `gorm:"column:ownerID;uniqueIndex:ownerID_slotValue,priority:1;type:varchar(64);comment:所有者ID" json:"ownerID"`                    //nolint:lll    // 所有者ID， 如果是个人统计，那就是用户ID， 如果是组，那就是组ID
	Value     decimal.Decimal `gorm:"column:value;type:decimal(30,8);comment:汇总值" json:"value"`                                                                 // 汇总值
	Value1    decimal.Decimal `gorm:"column:value_1;type:decimal(30,8);comment:汇总值1" json:"value_1"`                                                            //nolint:lll    // 汇总值1
	Value2    decimal.Decimal `gorm:"column:value_2;type:decimal(30,8);comment:汇总值2" json:"value_2"`                                                            //nolint:lll    // 汇总值2
	Value3    decimal.Decimal `gorm:"column:value_3;type:decimal(30,8);comment:汇总值3" json:"value_3"`                                                            //nolint:lll    // 汇总值3
	Value4    decimal.Decimal `gorm:"column:value_4;type:decimal(30,8);comment:汇总值4" json:"value_4"`                                                            //nolint:lll    // 汇总值4
	Value5    decimal.Decimal `gorm:"column:value_5;type:decimal(30,8);comment:汇总值5" json:"value_5"`                                                            //nolint:lll    // 汇总值5
	Value6    decimal.Decimal `gorm:"column:value_6;type:decimal(30,8);comment:汇总值6" json:"value_6"`                                                            //nolint:lll    // 汇总值6
	Value7    decimal.Decimal `gorm:"column:value_7;type:decimal(30,8);comment:汇总值7" json:"value_7"`                                                            //nolint:lll    // 汇总值7
	Value8    decimal.Decimal `gorm:"column:value_8;type:decimal(30,8);comment:汇总值8" json:"value_8"`                                                            //nolint:lll    // 汇总值8
	Value9    decimal.Decimal `gorm:"column:value_9;type:decimal(30,8);comment:汇总值9" json:"value_9"`                                                            //nolint:lll    // 汇总值9
	Value10   decimal.Decimal `gorm:"column:value_10;type:decimal(30,8);comment:汇总值_10" json:"value_10"`                                                        //nolint:lll    // 汇总值10
	SlotValue string          `gorm:"column:slotValue;type:varchar(64);uniqueIndex:ownerID_slotValue,priority:2;index:slotValue;comment:汇总时间" json:"slotValue"` //nolint:lll    // 所属的时间
	tableName string          // 表名
	Times     int64           `gorm:"column:times;comment:次数" json:"次数"`
}

func (s *Detail) SetValue(value decimal.Decimal) {
	s.Value = value
}

/*
newSummary 新建数据
参数:
*	slot     	Slot           	slot种类
*	ownerID  	string          所有者ID
*	value    	decimal.Decimal	值
*	slotValue	string          slot值
返回值:
*	*summary 	*summary       	返回值1
*/
func newSummary(slot Slot, ownerID string, value decimal.Decimal, slotValue string) *Detail {
	return &Detail{
		Slot:      slot,
		OwnerID:   ownerID,
		Value:     value,
		SlotValue: slotValue,
		Times:     1,
	}
}
func (s *Detail) SetExtendValue(index int, value decimal.Decimal) error {
	switch index {
	case 0:
		s.Value1 = value
	case 1:
		s.Value2 = value
	case 2:
		s.Value3 = value
	case 3:
		s.Value4 = value
	case 4:
		s.Value5 = value
	case 5:
		s.Value6 = value
	case 6:
		s.Value7 = value
	case 7:
		s.Value8 = value
	case 8:
		s.Value9 = value
	case 9:
		s.Value10 = value
	default:
		return fmt.Errorf(`最多支持10个扩展数据`)
	}

	return nil
}

func (s *Detail) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *Detail) MarshalBinary() (data []byte, err error) {
	return json.Marshal(s)
}

func (s Detail) TableName() string {
	return s.tableName
}

func (s Detail) GetID() int64 {
	return s.ID
}

func (s Detail) GetSlot() Slot {
	return s.Slot
}

func (s Detail) GetOwnerID() string {
	return s.OwnerID
}

func (s Detail) GetValue() decimal.Decimal {
	return s.Value
}
func (s Detail) GetExtendValue() []decimal.Decimal {
	return []decimal.Decimal{s.Value1, s.Value2, s.Value3, s.Value4, s.Value5, s.Value6, s.Value7, s.Value8, s.Value9, s.Value10}
}

func (s Detail) GetSlotValue() string {
	return s.SlotValue
}

func (s Detail) GetTimes() int64 {
	return s.Times
}

type Decimal8 decimal.Decimal

func (d Decimal8) GobEncode() ([]byte, error) {
	return decimal.Decimal(d).GobEncode()
}

func (d *Decimal8) GobDecode(bytes []byte) error {
	return (*decimal.Decimal)(d).GobDecode(bytes)
}

func (d Decimal8) Value() (driver.Value, error) {
	return decimal.Decimal(d).Shift(8).Value()
}

func (d *Decimal8) Scan(value any) error {
	// first try to see if the data is stored in database as a Numeric datatype
	origin := decimal.New(0, 0)
	if err := origin.Scan(value); err != nil {
		return err
	}

	*d = (Decimal8)(origin.Shift(-8))

	return nil
}

func (d *Decimal8) UnmarshalText(text []byte) error {
	return (*decimal.Decimal)(d).UnmarshalText(text)
}

func (d Decimal8) MarshalText() (text []byte, err error) {
	return decimal.Decimal(d).MarshalText()
}

func (d *Decimal8) UnmarshalJSON(bytes []byte) error {
	return (*decimal.Decimal)(d).UnmarshalJSON(bytes)
}

func (d Decimal8) MarshalJSON() ([]byte, error) {
	return decimal.Decimal(d).MarshalJSON()
}

func (d *Decimal8) UnmarshalBinary(data []byte) error {
	return (*decimal.Decimal)(d).UnmarshalBinary(data)
}

func (d Decimal8) MarshalBinary() (data []byte, err error) {
	return decimal.Decimal(d).MarshalBinary()
}

type DetailWithBigInt struct {
	ID        int64    `gorm:"column:id;primary_key;column:id;type:bigint(20) unsigned AUTO_INCREMENT;not null;comment:ID" json:"id"`
	Slot      Slot     `gorm:"column:slot;type:varchar(128);comment:槽位类型" json:"slot"`                                                                   //nolint:lll    // 槽位类型
	OwnerID   string   `gorm:"column:ownerID;uniqueIndex:ownerID_slotValue,priority:1;type:varchar(64);comment:所有者ID" json:"ownerID"`                    //nolint:lll    // 所有者ID， 如果是个人统计，那就是用户ID， 如果是组，那就是组ID
	Value     Decimal8 `gorm:"column:value;type:decimal(30,8);comment:汇总值" json:"value"`                                                                 // 汇总值
	Value1    Decimal8 `gorm:"column:value_1;type:decimal(30,8);comment:汇总值1" json:"value_1"`                                                            //nolint:lll    // 汇总值1
	Value2    Decimal8 `gorm:"column:value_2;type:decimal(30,8);comment:汇总值2" json:"value_2"`                                                            //nolint:lll    // 汇总值2
	Value3    Decimal8 `gorm:"column:value_3;type:decimal(30,8);comment:汇总值3" json:"value_3"`                                                            //nolint:lll    // 汇总值3
	Value4    Decimal8 `gorm:"column:value_4;type:decimal(30,8);comment:汇总值4" json:"value_4"`                                                            //nolint:lll    // 汇总值4
	Value5    Decimal8 `gorm:"column:value_5;type:decimal(30,8);comment:汇总值5" json:"value_5"`                                                            //nolint:lll    // 汇总值5
	Value6    Decimal8 `gorm:"column:value_6;type:decimal(30,8);comment:汇总值6" json:"value_6"`                                                            //nolint:lll    // 汇总值6
	Value7    Decimal8 `gorm:"column:value_7;type:decimal(30,8);comment:汇总值7" json:"value_7"`                                                            //nolint:lll    // 汇总值7
	Value8    Decimal8 `gorm:"column:value_8;type:decimal(30,8);comment:汇总值8" json:"value_8"`                                                            //nolint:lll    // 汇总值8
	Value9    Decimal8 `gorm:"column:value_9;type:decimal(30,8);comment:汇总值9" json:"value_9"`                                                            //nolint:lll    // 汇总值9
	Value10   Decimal8 `gorm:"column:value_10;type:decimal(30,8);comment:汇总值_10" json:"value_10"`                                                        //nolint:lll    // 汇总值10
	SlotValue string   `gorm:"column:slotValue;type:varchar(64);uniqueIndex:ownerID_slotValue,priority:2;index:slotValue;comment:汇总时间" json:"slotValue"` //nolint:lll    // 所属的时间
	tableName string   // 表名
	Times     int64    `gorm:"column:times;comment:次数" json:"次数"`
}

func (s DetailWithBigInt) SetValue(value decimal.Decimal) {
	s.Value = Decimal8(value)
}

/*
NewDetailWithBigInt 新建数据
参数:
*	slot     	Slot           	slot种类
*	ownerID  	string          所有者ID
*	value    	decimal.Decimal	值
*	slotValue	string          slot值
返回值:
*	*summary 	*summary       	返回值1
*/
func NewDetailWithBigInt(slot Slot, ownerID string, value decimal.Decimal, slotValue string) *DetailWithBigInt {
	return &DetailWithBigInt{
		Slot:      slot,
		OwnerID:   ownerID,
		Value:     Decimal8(value),
		SlotValue: slotValue,
		Times:     1,
	}
}
func (s *DetailWithBigInt) SetExtendValue(index int, originValue decimal.Decimal) error {
	value := Decimal8(originValue)

	switch index {
	case 0:
		s.Value1 = value
	case 1:
		s.Value2 = value
	case 2:
		s.Value3 = value
	case 3:
		s.Value4 = value
	case 4:
		s.Value5 = value
	case 5:
		s.Value6 = value
	case 6:
		s.Value7 = value
	case 7:
		s.Value8 = value
	case 8:
		s.Value9 = value
	case 9:
		s.Value10 = value
	default:
		return fmt.Errorf(`最多支持10个扩展数据`)
	}

	return nil
}

func (s *DetailWithBigInt) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *DetailWithBigInt) MarshalBinary() (data []byte, err error) {
	return json.Marshal(s)
}

func (s DetailWithBigInt) TableName() string {
	return s.tableName
}

func (s DetailWithBigInt) GetID() int64 {
	return s.ID
}

func (s DetailWithBigInt) GetSlot() Slot {
	return s.Slot
}

func (s DetailWithBigInt) GetOwnerID() string {
	return s.OwnerID
}

func (s DetailWithBigInt) GetValue() decimal.Decimal {
	return decimal.Decimal(s.Value)
}
func (s DetailWithBigInt) GetExtendValue() []decimal.Decimal {
	return []decimal.Decimal{decimal.Decimal(s.Value1), decimal.Decimal(s.Value2), decimal.Decimal(s.Value3), decimal.Decimal(s.Value4), decimal.Decimal(s.Value5), decimal.Decimal(s.Value6),
		decimal.Decimal(s.Value7), decimal.Decimal(s.Value8), decimal.Decimal(s.Value9), decimal.Decimal(s.Value10)}
}

func (s DetailWithBigInt) GetSlotValue() string {
	return s.SlotValue
}

func (s DetailWithBigInt) GetTimes() int64 {
	return s.Times
}
