package options

import (
	"fmt"
	"strings"
	"sync"

	"github.com/fighterlyt/common/model"
	"github.com/pkg/errors"
)

type item struct {
	Value int    `json:"value"`
	Text  string `json:"text"`
}

type AllItems struct {
	*sync.RWMutex
	data map[string][]item
}

var (
	allItems = &AllItems{
		RWMutex: &sync.RWMutex{},
		data:    make(map[string][]item, 10),
	}
)

func Register(items model.OptionItems) error {
	if items == nil {
		return errors.New(`参数不能为nil`)
	}

	if len(items.Items()) == 0 {
		return errors.New(`OptionItems() 不能返回空`)
	}

	if strings.TrimSpace(items.Key()) == `` {
		return errors.New(`Key()不能返回空字符串`)
	}

	allItems.Lock()
	defer allItems.Unlock()

	if _, exist := allItems.data[items.Key()]; exist {
		return fmt.Errorf(`%s 已经注册`, items.Key())
	}

	data := make([]item, 0, len(items.Items()))

	for _, elem := range items.Items() {
		data = append(data, item{
			Value: elem.Value(),
			Text:  elem.Text(),
		})
	}

	allItems.data[items.Key()] = data

	return nil
}

func Get(key string) []item {
	allItems.RLock()
	defer allItems.RUnlock()

	items := make([]item, len(allItems.data[key]))

	copy(items, allItems.data[key])

	return items
}
