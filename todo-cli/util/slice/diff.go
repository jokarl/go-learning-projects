package slice

// Diff returns a slice containing the
// elements of `a` that are not present in `b`.
func Diff[T comparable](a, b []T) []T {
	m := make(map[T]struct{}, len(b))
	for _, v := range b {
		m[v] = struct{}{}
	}

	var diff []T
	for _, v := range a {
		if _, found := m[v]; !found {
			diff = append(diff, v)
		}
	}

	return diff
}
