package index

import (
	"testing"

	"github.com/shopspring/decimal"
)

var (
	originValues []decimal.Decimal
	length       = 8
	tree         *DecimalTree
)

func TestBinaryIndexedTree_Init(t *testing.T) {
	originValues = make([]decimal.Decimal, 0, length)
	for i := 0; i < length; i++ {
		originValues = append(originValues, decimal.NewFromInt(int64(i+1)))
	}

	tree = NewDecimalTree(length)

	tree.Init(originValues)

	for i, value := range tree.values {
		t.Log(`value`, i, value.String())
	}
}

func TestBinaryIndexedTree_Sum(t *testing.T) {
	TestBinaryIndexedTree_Init(t)

	for i := 0; i < length; i++ {
		t.Log(`sum`, i+1, tree.Sum(i).String())
	}
}

func TestBinaryIndexedTree_RangeSum(t *testing.T) {
	TestBinaryIndexedTree_Init(t)

	for i := 0; i < length; i++ {
		for j := i; j < length; j++ {
			t.Logf(`rangeSum [%d-%d] %s`, i, j, tree.RangeSum(i, j))
		}
	}
}
