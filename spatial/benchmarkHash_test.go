package spatial

import "testing"

var (
	testSpecification = NewGridSpecification(64, 64, 16, 32, 32)
)

func BenchmarkGridSpecification_HashPositionIndexCapped(b *testing.B) {
	for n := 0; n < b.N; n++ {
		hash := testSpecification.HashPositionIndexCapped(0, 0)
		_ = hash
	}
}

func BenchmarkGridSpecification_HashPositionToroidalIndex(b *testing.B) {
	for n := 0; n < b.N; n++ {
		hash := testSpecification.HashPositionToroidalIndex(0, 0)
		_ = hash
	}
}
