package slices

func GroupBy[T any](input []T, eq func(*T, *T) bool) (groups [][]T) {
	if len(input) < 1 {
		return
	}

	start := 0
	for i := range input[1:] {
		if !eq(&input[i], &input[i+1]) {
			groups = append(groups, input[start:i+1])
			start = i + 1
		}
	}
	groups = append(groups, input[start:])
	return
}
