package gomath

// NextPowerOfTwo returns the next power of two >= n.
// If n is already a power of two, it returns n.
func NextPowerOfTwo(n uint64) uint64 {
	if n == 0 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32

	return n + 1
}
