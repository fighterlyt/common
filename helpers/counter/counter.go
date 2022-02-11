package summary

import "sync"

// Counter 是一个提供了写操作的计数器
type Counter interface {
	// Set 设置值
	Set(bit int64, value uint64)
	// Clear 无条件清理
	Clear(bit int64)
	// ClearIf  满足条件清理
	ClearIf(bit int64, value uint64)
	// Count 统计所有非默认值，真实写入的数量
	Count() int64
	// ClearAll 无条件清理全部
	ClearAll()
	// ClearAllIfNot 有条件清理全部
	ClearAllIfNot(value uint64)
}

const (
	placeHolder = 0
)

type counter struct {
	lock  *sync.RWMutex
	data  map[int64]uint64
	count int64
}

/*NewCounter 新建计数器，设置的值必须大于0
参数:
*	capacity	int64   	容器
返回值:
*	*counter	*counter	计数器
*/
func NewCounter(capacity int64) *counter {
	result := &counter{
		lock:  &sync.RWMutex{},
		data:  make(map[int64]uint64, capacity),
		count: 0,
	}

	for i := int64(0); i < capacity; i++ {
		result.data[i] = placeHolder
	}

	return result
}

/*Set 设置
参数:
*	bit  	int64 	位数
*	value	uint64	值，必须大于0
返回值:
*/
func (c *counter) Set(bit int64, value uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	old, exist := c.data[bit]

	c.data[bit] = value

	if !exist || old == placeHolder {
		c.count++
	}
}

/*Clear 清理单个
参数:
*	bit	int64	位置
返回值:
*/
func (c *counter) Clear(bit int64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	old, exist := c.data[bit]

	c.data[bit] = placeHolder

	if exist && old != placeHolder {
		c.count--
	}
}

func (c *counter) ClearIf(bit int64, value uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	old, exist := c.data[bit]

	if old != value {
		return
	}

	c.data[bit] = placeHolder

	if exist && old != placeHolder {
		c.count--
	}
}

func (c counter) Count() int64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.count
}

func (c *counter) ClearAll() {
	c.lock.Lock()
	defer c.lock.Unlock()

	for id, value := range c.data {
		if value != placeHolder { // 不是占位符，真实的值
			c.count--
		}

		c.data[id] = placeHolder
	}
}

func (c *counter) ClearAllIfNot(value uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for id, current := range c.data { // 遍历
		if current == value { // 不符合条件，过滤
			continue
		}

		if current != placeHolder { // 不是占位符，真实的值
			c.count--
		}

		c.data[id] = placeHolder
	}
}
