package gomath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Integer T is storage-only for geometry: magnitude methods truncate through
// float64. Pin that as documented, intentional behaviour.
func TestVector3_IntegerTruncationIsDocumented(t *testing.T) {
	// |(1,1,1)| = 1.73..., truncated to 1.
	assert.Equal(t, 1, Vector3[int]{1, 1, 1}.Length())
	// Exact when the magnitude is integral (2-3-6 -> 7).
	assert.Equal(t, 7, Vector3[int]{2, 3, 6}.Length())
	// Normalize collapses because the length truncates before the divide:
	// |(2,2,2)| = 3.46 -> 3, and 2/3 -> 0.
	assert.Equal(t, Vector3[int]{0, 0, 0}, Vector3[int]{2, 2, 2}.Normalize())
}

// SetLength and ScaleToLength intentionally disagree at length 0: SetLength(0)
// collapses a non-zero vector to zero; ScaleToLength(0) leaves it unchanged.
func TestVector3_SetLengthVsScaleToLengthAtZero(t *testing.T) {
	v := Vector3[float64]{3, 4, 0}
	assert.Equal(t, Vector3[float64]{0, 0, 0}, v.SetLength(0))
	assert.Equal(t, Vector3[float64]{3, 4, 0}, v.ScaleToLength(0))
}

// Limit caps magnitude at limit while preserving direction, and leaves vectors
// at or under the limit untouched. Guards the double-sqrt refactor.
func TestVector3_LimitContract(t *testing.T) {
	got := Vector3[float64]{3, 4, 0}.Limit(2.5)
	assert.InDelta(t, 1.5, got.X, 1e-9)
	assert.InDelta(t, 2.0, got.Y, 1e-9)
	assert.InDelta(t, 0, got.Z, 1e-9)
	assert.InDelta(t, 2.5, got.Length(), 1e-9)
	// Parallel to the original (cross product is the zero vector).
	assert.InDelta(t, 0, got.Cross(Vector3[float64]{3, 4, 0}).Length(), 1e-9)

	assert.Equal(t, Vector3[float64]{3, 4, 0}, Vector3[float64]{3, 4, 0}.Limit(5))
	assert.Equal(t, Vector3[float64]{3, 4, 0}, Vector3[float64]{3, 4, 0}.Limit(10))
}

// Set has a value receiver and returns a modified copy; it does NOT mutate the
// receiver.
func TestVector3_SetReturnsCopy(t *testing.T) {
	v := Vector3[float64]{1, 2, 3}
	got := v.Set(9, 9, 9)
	assert.Equal(t, Vector3[float64]{1, 2, 3}, v)
	assert.Equal(t, Vector3[float64]{9, 9, 9}, got)
}
