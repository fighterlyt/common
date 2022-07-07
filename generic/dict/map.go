package dict

import (
	"sync"
)

type Map[K comparable, V any] struct {
	entries map[K]V
	lock    *sync.RWMutex
}

func NewMap[K comparable, V any](capacity int) *Map[K, V] {
	return &Map[K, V]{
		entries: make(map[K]V, capacity),
		lock:    &sync.RWMutex{},
	}
}

func (m *Map[K, V]) AddWithCheck(key K, value V) bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, exist := m.entries[key]; exist {
		return false
	}

	m.entries[key] = value

	return true
}

func (m *Map[K, V]) Add(key K, value V) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.entries[key] = value
}

func (m *Map[K, V]) Exist(key K) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	_, exist := m.entries[key]

	return exist
}

func (m *Map[K, V]) Get(key K) V {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.entries[key]
}

func (m *Map[K, V]) GetWithExist(key K) (v V, exist bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	v, exist = m.entries[key]

	return v, exist
}
