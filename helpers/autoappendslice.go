package helpers

import (
	"sync"
)

var (
	defaultCapacity = 10 // 一级slice的容量
)

// AutoAppendSlice 自动扩容的slice,为了避免数据量过大导致的扩容问题，在创建的时候，创建一定的容量，一旦超过这个容量，BitSlice会自动创建一个一级slice
type AutoAppendSlice struct {
	mutex           *sync.Mutex
	elements        []*autoAppendSliceData
	offset          int // 偏移量
	elementCapacity int // 每个二级slice的容量
}

func NewAutoAppendSlice(offset int, capacity int) *AutoAppendSlice {
	bitSlice := &AutoAppendSlice{
		mutex:           &sync.Mutex{},
		elements:        make([]*autoAppendSliceData, defaultCapacity),
		offset:          offset,
		elementCapacity: capacity,
	}

	for i := 0; i < defaultCapacity; i++ {
		bitSlice.elements[i] = newAutoAppendSliceData(capacity)
	}

	return bitSlice
}

type autoAppendSliceData struct {
	mutex *sync.Mutex
	data  []interface{}
}

func newAutoAppendSliceData(capacity int) *autoAppendSliceData {
	return &autoAppendSliceData{
		mutex: &sync.Mutex{},
		data:  make([]interface{}, capacity),
	}
}

func (a *autoAppendSliceData) get(index int) interface{} {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.data[index]
}

func (a *autoAppendSliceData) set(index int, data interface{}) error {
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
func (a *AutoAppendSlice) GetElement(elementID int) (interface{}, error) {
	firstIndex, secondIndex := a.getIndex(elementID)

	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.elements[firstIndex].get(secondIndex), nil
}

func (a *AutoAppendSlice) getIndex(elementID int) (firstIndex, secondIndex int) {
	firstIndex = (elementID - a.offset - 1) / a.elementCapacity  // 一级slice的下标
	secondIndex = (elementID - a.offset - 1) % a.elementCapacity // 二级slice的下标

	var distance = firstIndex + 1 - len(a.elements)

	if distance > 0 {
		for i := 0; i < distance; i++ {
			a.elements = append(a.elements, newAutoAppendSliceData(a.elementCapacity))
		}
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
func (a *AutoAppendSlice) SetElement(elementID int, data interface{}) error {
	firstIndex, secondIndex := a.getIndex(elementID)

	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.elements[firstIndex].set(secondIndex, data)
}
