package helpers

import (
	"sync"
)

var (
	defaultCapacity       = 10    // 一级slice的容量
	defaultSecondCapacity = 10000 // 二级slice的容量
)

// BitSlice 二级slice,为了避免数据量过大导致的扩容问题，在创建的时候，创建一定的容量，一旦超过这个容量，BitSlice会自动创建一个一级slice
type BitSlice struct {
	mutex  *sync.Mutex
	data   []*bitSliceData
	offset int // 偏移量
}

func NewBitSlice(offset int) *BitSlice {
	bitSlice := &BitSlice{
		mutex:  &sync.Mutex{},
		data:   make([]*bitSliceData, defaultCapacity),
		offset: offset,
	}

	for i := 0; i < defaultCapacity; i++ {
		bitSlice.data[i] = newBitSliceData()
	}

	return bitSlice
}

type bitSliceData struct {
	mutex *sync.Mutex
	data  []interface{}
}

func newBitSliceData() *bitSliceData {
	return &bitSliceData{
		mutex: &sync.Mutex{},
		data:  make([]interface{}, defaultSecondCapacity),
	}
}
