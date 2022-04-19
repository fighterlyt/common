package cache

import (
	"sync"

	"github.com/pkg/errors"
)

var (
	initCapacity = 10
)

// Cache 泛型缓存，key支持int/int64/string,value 支持任意类型
type Cache[K int | int64 | string, V any] struct {
	data map[K]V
	*sync.RWMutex
}

// NewCache 新建Cache
func NewCache[K int | int64 | string, V any]() *Cache[K, V] {
	result := Cache[K, V]{}
	result.init()

	return &result
}

func (c *Cache[K, V]) init() {
	c.data = make(map[K]V, initCapacity)
	c.RWMutex = &sync.RWMutex{}
}

/*AddBatch 批量添加
参数:
*	keys  	[]K  	键值数组
*	values	[]V  	值数组
返回值:
*	error 	error	错误
*/
func (c *Cache[K, V]) AddBatch(keys []K, values []V) error {
	if len(keys) != len(values) {
		return errors.New(`数量必须相同`)
	}

	c.Lock()
	defer c.Unlock()

	for i := range keys {
		c.add(keys[i], values[i], false)
	}

	return nil
}

/*Add 添加元素
参数:
*	key  	K	键
*	value	V	值
返回值:
*/
func (c *Cache[K, V]) Add(key K, value V) {
	c.add(key, value, true)
}

func (c *Cache[K, V]) add(key K, value V, needLock bool) {
	if needLock {
		c.Lock()
		defer c.Unlock()
	}

	if c.data == nil {
		c.data = make(map[K]V, initCapacity)
	}

	c.data[key] = value
}

/*Remove 根据key删除
参数:
*	key	K	键值
返回值:
*/
func (c *Cache[K, V]) Remove(key K) {
	c.remove(key, true)
}

func (c *Cache[K, V]) remove(key K, needLock bool) {
	if needLock {
		c.Lock()
		defer c.Unlock()
	}

	delete(c.data, key)
}

/*Get 根据兼职获取值
参数:
*	key	K	键值
返回值:
*	V  	V	值
*/
func (c Cache[K, V]) Get(key K) V {
	c.RLock()
	defer c.RUnlock()

	return c.data[key]
}

/*Update 更新
参数:
*	key  	K	键值
*	value	V	值
返回值:
*/
func (c *Cache[K, V]) Update(key K, value V) {
	c.Lock()
	defer c.Unlock()

	c.remove(key, false)
	c.add(key, value, false)
}
