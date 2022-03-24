package header

// Service 下拉框服务
type Service interface {
	// Register 注册
	Register(items Items) error
}

type Items interface {
	// Key 组的key
	Key() string
	// Items 所有记录
	Items() []Resp
}

type headerItems struct {
	key   string
	items []Resp
}

func (h *headerItems) Key() string {
	return h.key
}

func (h *headerItems) Items() []Resp {
	return h.items
}

func NewHeaderItems(key string, items []Resp) *headerItems {
	return &headerItems{
		key:   key,
		items: items,
	}
}
