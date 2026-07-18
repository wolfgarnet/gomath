package gomath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Item 1: integer T is storage-only. Magnitude/geometry methods round-trip
// through float64 and truncate back to T, so they are meaningless for ints.
// These tests pin that as documented, intentional behaviour (use a float T for
// geometry).
func TestVector2_IntegerTruncationIsDocumented(t *testing.T) {
	// |(1,1)| = 1.414..., truncated to 1.
	assert.Equal(t, 1, Vector2[int]{1, 1}.Length())
	// Exact when the magnitude is integral.
	assert.Equal(t, 5, Vector2[int]{3, 4}.Length())
	// Normalize collapses because the length truncates before the divide:
	// |(3,3)| = 4.24 -> 4, and 3/4 -> 0.
	assert.Equal(t, Vector2[int]{0, 0}, Vector2[int]{3, 3}.Normalize())
}

// Item 2: SetLength and ScaleToLength intentionally disagree at length 0.
// SetLength(0) collapses to the zero vector; ScaleToLength(0) leaves the vector
// unchanged. Pin the divergence so it stays a conscious choice.
func TestVector2_SetLengthVsScaleToLengthAtZero(t *testing.T) {
	v := Vector2[float64]{3, 4}
	assert.Equal(t, Vector2[float64]{0, 0}, v.SetLength(0))
	assert.Equal(t, Vector2[float64]{3, 4}, v.ScaleToLength(0))
}

// Item 3: Limit caps magnitude at limit while preserving direction, and leaves
// vectors at or under the limit untouched. Guards the double-sqrt refactor.
func TestVector2_LimitContract(t *testing.T) {
	// Over the limit: result length is exactly the limit, same direction.
	got := Vector2[float64]{3, 4}.Limit(2.5)
	assert.InDelta(t, 1.5, got.X, 1e-9)
	assert.InDelta(t, 2.0, got.Y, 1e-9)
	assert.InDelta(t, 2.5, got.Length(), 1e-9)
	assert.InDelta(t, 0, got.Cross(Vector2[float64]{3, 4}), 1e-9) // parallel to original

	// At or under the limit: unchanged.
	assert.Equal(t, Vector2[float64]{3, 4}, Vector2[float64]{3, 4}.Limit(5))
	assert.Equal(t, Vector2[float64]{3, 4}, Vector2[float64]{3, 4}.Limit(10))
}

// Item 4: Set has a value receiver and returns a modified copy; it does NOT
// mutate the receiver. Pin that so callers who write `v.Set(...)` and discard
// the result have their mistake caught by the contract.
func TestVector2_SetReturnsCopy(t *testing.T) {
	v := Vector2[float64]{1, 2}
	got := v.Set(9, 9)
	assert.Equal(t, Vector2[float64]{1, 2}, v)   // original untouched
	assert.Equal(t, Vector2[float64]{9, 9}, got) // copy carries the new values
}

// Item 4: NormalClockwise/NormalCounterClockwise return perpendiculars (not
// normalized) of the same length as the source, pointing in opposite senses.
func TestVector2_NormalClockwiseCounterClockwise(t *testing.T) {
	v := Vector2[float64]{3, 4}
	cw := v.NormalClockwise()
	ccw := v.NormalCounterClockwise()

	assert.Equal(t, Vector2[float64]{4, -3}, cw)
	assert.Equal(t, Vector2[float64]{-4, 3}, ccw)

	// Perpendicular to the source, and opposite to each other.
	assert.InDelta(t, 0, v.Dot(cw), 1e-9)
	assert.InDelta(t, 0, v.Dot(ccw), 1e-9)
	assert.Equal(t, cw, ccw.Negate())

	// Not normalized: same magnitude as the source.
	assert.InDelta(t, v.Length(), cw.Length(), 1e-9)
}
