package index

import "github.com/shopspring/decimal"

type DecimalTree struct {
	values []decimal.Decimal
}

func NewDecimalTree(length int) *DecimalTree {
	return &DecimalTree{
		values: make([]decimal.Decimal, length+1),
	}
}

func (b DecimalTree) Init(values []decimal.Decimal) {
	for i, value := range values {
		b.Update(i, value)
	}
}

func (b DecimalTree) Update(index int, delta decimal.Decimal) {
	index++

	for index < len(b.values) {
		b.values[index] = b.values[index].Add(delta)
		index += index & -index
	}
}

func (b DecimalTree) Sum(index int) decimal.Decimal {
	index++

	result := decimal.Zero

	for index > 0 {
		result = result.Add(b.values[index])
		index -= index & -index
	}

	return result
}

func (b DecimalTree) RangeSum(from, to int) decimal.Decimal {
	return b.Sum(to).Sub(b.Sum(from - 1))
}
