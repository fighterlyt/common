package summary

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type Slot string

const (
	// SlotDay 按天汇总
	SlotDay Slot = `天`
	// SlotWhole 总汇总
	SlotWhole Slot = `开天辟地到地老天荒`
)

// Detail 真实数据
type Detail struct {
	ID        int64           `gorm:"column:id;primary_key;column:id;type:bigint(20) unsigned AUTO_INCREMENT;not null;comment:ID" json:"id"`
	Slot      Slot            `gorm:"column:slot;type:varchar(128);comment:槽位类型" json:"slot"`                                       // 槽位类型
	OwnerID   int64           `gorm:"column:ownerID;uniqueIndex:ownerID_slotValue;type:bigint(20);comment:所有者ID" json:"ownerID"`    //nolint:lll    // 所有者ID， 如果是个人统计，那就是用户ID， 如果是组，那就是组ID
	Value     decimal.Decimal `gorm:"column:value;type:decimal(30,8);comment:汇总值" json:"value"`                                     // 汇总值
	SlotValue int64           `gorm:"column:slotValue;type:bigint(10);uniqueIndex:ownerID_slotValue;comment:汇总时间" json:"slotValue"` // 所属的时间
	tableName string          // 表名
	Times     int64           `gorm:"column:times;comment:次数" json:"次数"`
}

/*newSummary 新建数据
参数:
*	slot     	Slot           	slot种类
*	ownerID  	int64          	所有者ID
*	value    	decimal.Decimal	值
*	slotValue	int64          	slot值
返回值:
*	*summary 	*summary       	返回值1
*/
func newSummary(slot Slot, ownerID int64, value decimal.Decimal, slotValue int64) *Detail {
	return &Detail{
		Slot:      slot,
		OwnerID:   ownerID,
		Value:     value,
		SlotValue: slotValue,
	}
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

func (s Detail) GetOwnerID() int64 {
	return s.OwnerID
}

func (s Detail) GetValue() decimal.Decimal {
	return s.Value
}

func (s Detail) GetSlotValue() int64 {
	return s.SlotValue
}

func (s Detail) GetTimes() int64 {
	return s.Times
}
