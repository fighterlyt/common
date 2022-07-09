package slice

func RemoveDuplicate[T int | int64 | string](data []T) []T {
	m := make(map[T]struct{}, 10)

	for i := range data {
		m[data[i]] = struct{}{}
	}

	result := make([]T, 0, len(m))

	for k := range m {
		result = append(result, k)
	}

	return result
}
