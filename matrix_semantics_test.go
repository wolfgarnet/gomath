package gomath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Set takes its arguments in row-major order (n11, n12, n13, n21, ...) even
// though the backing array is column-major. Pin that mapping.
func TestMatrix_SetIsRowMajor(t *testing.T) {
	var m Matrix[float64]
	m.Set(
		1, 2, 3,
		4, 5, 6,
		7, 8, 9,
	)
	// Column-major storage: element(row, col) lives at m[col*3+row].
	assert.Equal(t, Matrix[float64]{1, 4, 7, 2, 5, 8, 3, 6, 9}, m)
	// The translation column (col 2, rows 0 and 1) is the position.
	assert.Equal(t, Vector2[float64]{3, 6}, m.GetPosition())
}

// Translation is additive: it adds to the existing translation column rather
// than replacing it. SetPosition replaces. Pin the distinction.
func TestMatrix_TranslationIsAdditive(t *testing.T) {
	m := NewMatrix[float64]()
	m.Translation(1, 2)
	m.Translation(3, 4)
	assert.Equal(t, Vector2[float64]{4, 6}, m.GetPosition())

	m.SetPosition(10, 20)
	assert.Equal(t, Vector2[float64]{10, 20}, m.GetPosition())
}

// Integer T is storage-only for the transform operations that go through
// float64 (Rotation/Angle) or division (Inverse). Element-wise ops stay exact.
func TestMatrix_IntegerIsStorageOnly(t *testing.T) {
	// A non-orthonormal integer matrix has |det| != 1, so the inverse divides
	// and truncates to garbage: report the truncation as intentional.
	m := Matrix[int]{2, 0, 0, 0, 2, 0, 0, 0, 1} // scale-by-2, det 4
	inv, ok := m.Inverse()
	assert.True(t, ok)
	// 1/4 truncates to 0 in integer arithmetic, so the inverse is all zeros.
	assert.Equal(t, Matrix[int]{}, inv)

	// Element-wise scalar multiply is exact for integers.
	assert.Equal(t, Matrix[int]{2, 4, 6, 8, 10, 12, 14, 16, 18}, Matrix[int]{1, 2, 3, 4, 5, 6, 7, 8, 9}.MultiplyScalar(2))
}
