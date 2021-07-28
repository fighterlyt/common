package message

// Record 信息记录
type Record struct {
	ID    int64  `gorm:"column:id;primaryKey;column:id;type:bigint(20) unsigned AUTO_INCREMENT;not null;comment:'ID'" json:"id"`   // ID
	Key   string `gorm:"column:elemKey;index:elemKey_value,unique,priority:1;not null;type:varchar(128);comment:key" json:"key"`   // 分类信息
	Value string `gorm:"column:value;index:elemKey_value,unique,priority:2;not null;type:varchar(255);comment:value" json:"value"` // 值
}

/*NewRecord 新建记录信息
参数:
*	key    	string 	分类key
*	value  	string 	值
返回值:
*	*Record	*Record	返回值1
*/
func NewRecord(key, value string) *Record {
	return &Record{Key: key, Value: value}
}

func (r Record) TableName() string {
	return `generic_message`
}
