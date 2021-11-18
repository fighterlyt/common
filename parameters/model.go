package parameters

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
)

const (
	validateFailFMT = `值[%s]未通过验证[%s]`
)

// Parameter 业务参数
type Parameter struct {
	Key         string `gorm:"column:elemKey;primaryKey;type:varchar(32);comment:唯一识别符" valid:"key,required,stringlength(1|32)" json:"key"` //nolint:lll    // 唯一识别符
	Purpose     string `gorm:"column:purpose;type:varchar(256);comment:用途" valid:"required,stringlength(1|256)" json:"purpose"`             //nolint:lll    // 用途
	Value       string `gorm:"column:value;type:varchar(1024);comment:值" valid:"required,stringlength(1|1024)" json:"value"`                //nolint:lll    // 值
	Description string `gorm:"column:description;type:varchar(256);comment:值描述" valid:"required,stringlength(1|256)" json:"description"`    //nolint:lll    // 值描述
	UpdateTime  int64  `gorm:"column:updateTime;type:bigint;comment:更新时间" json:"update_time"`                                               //nolint:lll    // 最后更新时间
	ValidKey    string `gorm:"column:validKey;type:varchar(32);comment:验证方法key" valid:"required,ascii" json:"validKey"`                     //nolint:lll    // 验证方法key github.com/asaskevich/govalidator
	Hide        bool   `gorm:"column:lock;comment:是否锁定,前端获取不到也不能修改" valid:"isBool" json:"hide"`
	Err         error  `gorm:"-" json:"-"` // 错误信息
}

/*NewParameter 新建业务参数
参数:
*	key        	string    	唯一识别符
*	purpose    	string    	用途
*	value      	string    	值
*	description	string    	值描述
*	validKey   	string    	最后更新时间
返回值:
*	*Parameter 	*Parameter	业务参数
*/
func NewParameter(key, purpose, value, description, validKey string) *Parameter {
	return &Parameter{
		Key:         key,
		Purpose:     purpose,
		Value:       value,
		Description: description,
		ValidKey:    validKey,
		UpdateTime:  time.Now().Unix(),
	}
}

/*UnmarshalBinary 反序列化，方便使用Redis SCAN
参数:
*	data 	[]byte	参数1
返回值:
*	error	error 	返回值1
*/
func (p *Parameter) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

/*MarshalBinary 序列化
参数:
返回值:
*	data	[]byte	返回值1
*	err 	error 	返回值2
*/
func (p Parameter) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

/*Validate 验证 Parameter.Value,Parameter.Err 可被修改
参数:
返回值:
*	error	error	返回值1
*/

func (p *Parameter) Validate() error { // nolint:golint,revive
	if _, err := govalidator.ValidateStruct(p); err != nil {
		return errors.Wrap(err, `字段不满足`)
	}

	var (
		validator          govalidator.Validator
		parameterValidator govalidator.ParamValidator
		interfaceValidator govalidator.InterfaceParamValidator
		customValidator    govalidator.CustomTypeValidator
		exist              bool
	)

	if validator, exist = govalidator.TagMap[p.ValidKey]; exist {
		if !validator(p.Value) {
			return fmt.Errorf(validateFailFMT, p.Value, p.ValidKey)
		}

		return nil
	}

	if parameterValidator, exist = govalidator.ParamTagMap[p.ValidKey]; exist {
		if !parameterValidator(p.Value) {
			return fmt.Errorf(validateFailFMT, p.Value, p.ValidKey)
		}

		return nil
	}

	if interfaceValidator, exist = govalidator.InterfaceParamTagMap[p.ValidKey]; exist {
		if !interfaceValidator(p.Value) {
			return fmt.Errorf(validateFailFMT, p.Value, p.ValidKey)
		}

		return nil
	}

	if customValidator, exist = govalidator.CustomTypeTagMap.Get(p.ValidKey); exist {
		if !customValidator(p.Value, p) {
			if p.Err != nil {
				return p.Err
			}
			return fmt.Errorf(validateFailFMT, p.Value, p.ValidKey)
		}

		return nil
	}

	return fmt.Errorf(`[%s]不是合法的验证方法Key`, p.ValidKey)
}

/*TableName mysql表名
参数:
返回值:
*	string	string	表名
*/
func (p Parameter) TableName() string {
	return `parameters`
}

// History 变更历史
type History struct {
	ID         int64  `gorm:"column:id;primaryKey;comment:id'"`
	Key        string `gorm:"column:elemKey;index;type:varchar(64);comment:识别符" valid:"alpha"`        // key
	Value      string `gorm:"column:value;type:varchar(1024);comment:值" valid:"stringlength(1|1024)"` // 值
	UpdateTime int64  `gorm:"column:updateTime;type:bigint"`                                          // 最后更新时间
	UserID     int64  `gorm:"column:userID;type:bigint;comment:修改用户ID"`                               // 修改用户ID
}

/*NewHistory 新建一个变更记录
参数:
*	key     	string  	参数key
*	value   	string  	最新值
*	userID  	int64   	操作用户ID
返回值:
*	*History	*History	返回值1
*/
func NewHistory(key, value string, userID int64) *History {
	return &History{
		Key:        key,
		Value:      value,
		UserID:     userID,
		UpdateTime: time.Now().Unix(),
	}
}

/*TableName mysql表名
参数:
返回值:
*	string	string	表名
*/
func (History) TableName() string {
	return `parameters_history`
}
