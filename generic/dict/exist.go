package dict

// Exists 只是用来判断是否有存在
type Exists[K comparable] struct {
	*Map[K, struct{}]
}

func NewExists[K comparable](capacity int) *Exists[K] {
	return &Exists[K]{
		Map: NewMap[K, struct{}](capacity),
	}
}

func (e *Exists[K]) Add(key K) {
	e.Map.Add(key, struct{}{})
}
