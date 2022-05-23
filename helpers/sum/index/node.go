package index

import "github.com/shopspring/decimal"

type ValueKind string

type Node[T int64 | string] interface {
	// Add 累加某一项值
	Add(value decimal.Decimal, kind ValueKind)
	// Get  获取某一项值
	Get(kind ValueKind) decimal.Decimal
	ID() T
}

type node struct {
	lines []lineInfo
}

type lineInfo struct {
	tree  *MultiTree
	index int
}

type MultiTree[T int64 | string] struct {
	values []Node
	index  map[T]Node
}

func NewMultiTree[T int64 | string](length int) *MultiTree[T] {
	return &MultiTree[T]{
		values: make([]Node, length+1),
		index:  make(map[T]Node, length),
	}
}

func (b MultiTree[T]) Init(nodes []Node, kinds []ValueKind) {
	for _, kind := range kinds {
		b.UpdateNode(nodes, kind)
	}
}

func (b MultiTree[T]) UpdateNode(nodes []Node, kind ValueKind) {
	for i, node := range nodes {
		b.Update(i+1, node.Get(kind), kind)
		b.index[node.ID()] = nodes[i]
	}
}
func (b MultiTree[T]) Update(index int, delta decimal.Decimal, kind ValueKind) {
	index++

	for index < len(b.values) {
		b.values[index].Add(delta, kind)
		index += index & -index
	}
}

func (b MultiTree[T]) Sum(index int, kind ValueKind) decimal.Decimal {
	index++

	result := decimal.Zero

	for index > 0 {
		result = result.Add(b.values[index].Get(kind))
		index -= index & -index
	}

	return result
}

func (b MultiTree[T]) RangeSum(from, to int, kind ValueKind) decimal.Decimal {
	return b.Sum(to, kind).Sub(b.Sum(from-1, kind))
}
