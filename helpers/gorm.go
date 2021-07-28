package helpers

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
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
