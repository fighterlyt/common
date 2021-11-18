package model

import "encoding/json"

// OptionService 下拉框服务
type OptionService interface {
	// Register 注册
	Register(items OptionItems) error
}

// OptionItem 下拉框选项
type OptionItem interface {
	Value() int
	Text() string
	MarshalJSON() ([]byte, error)
}

// OptionItems 下拉框选项组
type OptionItems interface {
	// Key 组的key
	Key() string
	// Items 所有记录
	Items() []OptionItem
}

func MarshalJSON(item OptionItem) ([]byte, error) {
	return json.Marshal(temp{
		Value: item.Value(),
		Text:  item.Text(),
	})
}

type temp struct {
	Value int    `json:"value"`
	Text  string `json:"text"`
}

type optionItems struct {
	key   string
	items []OptionItem
}

func (o optionItems) Key() string {
	return o.key
}

func (o optionItems) Items() []OptionItem {
	return o.items
}

func NewOptionItems(key string, items ...OptionItem) OptionItems {
	return &optionItems{key: key, items: items}
}
