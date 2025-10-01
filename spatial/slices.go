package spatial

// RemoveUnordered removes the element at index i from the slice without
// preserving order. It returns the modified slice.
//
//go:inline
func RemoveUnordered[T any](s []T, i int) []T {
	if i < 0 || i >= len(s) {
		return s
	}
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
