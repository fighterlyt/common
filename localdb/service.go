package localdb

// Service 本地存储接口
type Service interface {
	// Read 获取，key是key值,data 是写入的数据，注意：必须是指针
	Read(key []byte, data Item) error
	// Write 写入
	Write(data Item) error
	// Delete 删除
	Delete(key []byte) error
	// IsNotFound 错误是否是未找到，如果err==nil,返回false
	IsNotFound(err error) bool
	// Close 关闭
	Close() error
}

// Item 存储的实体，每个实体的Key()不同，且不能为nil或者空
type Item interface {
	// Key 获取对象的key
	Key() []byte
	// Encode 编码
	Encode() ([]byte, error)
	// Decode 解码
	Decode([]byte) error
}
