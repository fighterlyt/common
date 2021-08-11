package message

import "context"

// Service 服务接口
type Service interface {
	// Get 获取一类数据
	Get(key string) (message []string, err error)
	// Exist 判断是否存在
	Exist(key, message string) (exists bool, err error)
	// Add 添加记录
	Add(ctx context.Context, key, message string) error
	// Delete 删除
	Delete(key string, messages ...string) error
}
