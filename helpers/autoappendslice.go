package helpers

import (
	"github.com/pkg/errors"
	"sync"
)

var (
	defaultCapacity = 10 // 一级slice的容量
)

// AutoAppendSlice 自动扩容的slice,为了避免数据量过大导致的扩容问题，在创建的时候，创建一定的容量，一旦超过这个容量，BitSlice会自动创建一个一级slice
type AutoAppendSlice struct {
	mutex           *sync.RWMutex
	elements        []*autoAppendSliceData
	offset          int // 偏移量
	elementCapacity int // 每个二级slice的容量
}

func NewAutoAppendSlice(offset, capacity int) (*AutoAppendSlice, error) {
	if offset < 0 {
		return nil, errors.New("偏移量必须大于0")
	}

	if capacity <= 0 {
		return nil, errors.New("二位数组容量必须大于0")
	}

	bitSlice := &AutoAppendSlice{
		mutex:           &sync.RWMutex{},
		elements:        make([]*autoAppendSliceData, defaultCapacity),
		offset:          offset,
		elementCapacity: capacity,
	}

	for i := 0; i < defaultCapacity; i++ {
		bitSlice.elements[i] = newAutoAppendSliceData(capacity)
	}

	return bitSlice, nil
}

type autoAppendSliceData struct {
	mutex *sync.RWMutex
	data  []interface{}
}

func newAutoAppendSliceData(capacity int) *autoAppendSliceData {
	return &autoAppendSliceData{
		mutex: &sync.RWMutex{},
		data:  make([]interface{}, capacity),
	}
}

func (a *autoAppendSliceData) get(index int64) interface{} {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.data[index]
}

func (a *autoAppendSliceData) set(index int64, data interface{}) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.data[index] = data

	return nil
}

/*GetElement 获取元素
参数:
*	elementID   int        	唯一ID,可以是用户ID,或其他
返回值:
*	interface{}	interface{} 对应的值
*/
func (a *AutoAppendSlice) GetElement(elementID int64) (interface{}, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	firstIndex, secondIndex := a.getIndexAndAutoAppend(elementID, true)

	return a.elements[firstIndex].get(secondIndex), nil
}

func (a *AutoAppendSlice) getIndexAndAutoAppend(elementID int64, isRLock bool) (firstIndex, secondIndex int64) {
	var realIndex = elementID - int64(a.offset-1)
	firstIndex = realIndex / int64(a.elementCapacity)  // 一级slice的下标
	secondIndex = realIndex % int64(a.elementCapacity) // 二级slice的下标

	var distance = int(firstIndex) + 1 - len(a.elements)

	if distance == 0 {
		return firstIndex, secondIndex
	}

	if isRLock {
		a.mutex.RUnlock()

		a.mutex.Lock()
		defer func() {
			a.mutex.Unlock()
			a.mutex.RLock()
		}()
	}

	for i := 0; i < distance; i++ {
		a.elements = append(a.elements, newAutoAppendSliceData(a.elementCapacity))
	}

	return firstIndex, secondIndex
}

/*SetElement 设置数据
参数:
*	elementID	int        	元素ID
*	data     	interface{}	数据
返回值:
*	error    	error      	错误
*/
func (a *AutoAppendSlice) SetElement(elementID int64, data interface{}) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	firstIndex, secondIndex := a.getIndexAndAutoAppend(elementID, false)

	return a.elements[firstIndex].set(secondIndex, data)
}
