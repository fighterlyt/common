package header

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"sync"
)

type AllItems struct {
	*sync.RWMutex
	data map[string][]Resp
}

var (
	allItems = &AllItems{
		RWMutex: &sync.RWMutex{},
		data:    make(map[string][]Resp, 10),
	}
)

func Register(items Items) error {
	if items == nil {
		return errors.New(`参数不能为nil`)
	}

	if len(items.Items()) == 0 {
		return errors.New(`Items() 不能返回空`)
	}

	if strings.TrimSpace(items.Key()) == `` {
		return errors.New(`Key()不能返回空字符串`)
	}

	allItems.Lock()
	defer allItems.Unlock()

	if _, exist := allItems.data[items.Key()]; exist {
		return fmt.Errorf(`%s 已经注册`, items.Key())
	}

	data := make([]Resp, 0, len(items.Items()))

	for _, elem := range items.Items() {
		data = append(data, elem)
	}

	allItems.data[items.Key()] = data

	return nil
}

func Get(key string) []Resp {
	allItems.RLock()
	defer allItems.RUnlock()

	items := make([]Resp, len(allItems.data[key]))

	copy(items, allItems.data[key])

	return items
}
