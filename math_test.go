package gomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAbs_Integers pins that Abs is exact across the full integer range.
// Routing through float64 loses precision above 2^53 and overflows on the way back.
func TestAbs_Integers(t *testing.T) {
	assert.Equal(t, 5, Abs(5))
	assert.Equal(t, 5, Abs(-5))
	assert.Equal(t, 0, Abs(0))

	assert.Equal(t, int64(math.MaxInt64), Abs(int64(math.MaxInt64)))
	assert.Equal(t, int64(math.MaxInt64), Abs(int64(-math.MaxInt64)))
	assert.Equal(t, int64(math.MaxInt64-1), Abs(int64(-(math.MaxInt64 - 1))))

	// Just above 2^53, where float64 can no longer represent every integer.
	const beyondFloat64Mantissa = int64(1)<<53 + 1
	assert.Equal(t, beyondFloat64Mantissa, Abs(-beyondFloat64Mantissa))
}

func TestAbs_Floats(t *testing.T) {
	assert.InDelta(t, 1.5, Abs(1.5), 0.0001)
	assert.InDelta(t, 1.5, Abs(-1.5), 0.0001)
	assert.InDelta(t, 0.0, Abs(0.0), 0.0001)
	assert.True(t, math.IsNaN(float64(Abs(math.NaN()))))
	assert.True(t, math.IsInf(float64(Abs(math.Inf(-1))), 1))
}

// TestMaxMin_Integers pins that Max/Min are exact across the full integer range.
func TestMaxMin_Integers(t *testing.T) {
	assert.Equal(t, 7, Max(3, 7))
	assert.Equal(t, 7, Max(7, 3))
	assert.Equal(t, 3, Min(3, 7))
	assert.Equal(t, 3, Min(7, 3))
	assert.Equal(t, -3, Max(-3, -7))
	assert.Equal(t, -7, Min(-3, -7))

	assert.Equal(t, int64(math.MaxInt64), Max(int64(math.MaxInt64), 1))
	assert.Equal(t, int64(1), Min(int64(math.MaxInt64), 1))
	assert.Equal(t, int64(math.MinInt64), Min(int64(math.MinInt64), -1))
	assert.Equal(t, int64(-1), Max(int64(math.MinInt64), -1))

	// Adjacent integers above 2^53 collapse to the same float64.
	const a = int64(1)<<53 + 1
	const b = int64(1) << 53
	assert.Equal(t, a, Max(a, b))
	assert.Equal(t, b, Min(a, b))
}

func TestMaxMin_Floats(t *testing.T) {
	assert.InDelta(t, 7.5, Max(3.5, 7.5), 0.0001)
	assert.InDelta(t, 3.5, Min(3.5, 7.5), 0.0001)
	assert.InDelta(t, -3.5, Max(-3.5, -7.5), 0.0001)
	assert.InDelta(t, -7.5, Min(-3.5, -7.5), 0.0001)
}

func TestClamp(t *testing.T) {
	assert.Equal(t, 5, Clamp(5, 0, 10))
	assert.Equal(t, 0, Clamp(-5, 0, 10))
	assert.Equal(t, 10, Clamp(15, 0, 10))
	assert.Equal(t, 0, Clamp(0, 0, 10))
	assert.Equal(t, 10, Clamp(10, 0, 10))
}

func TestSqr(t *testing.T) {
	assert.Equal(t, 9, Sqr(3))
	assert.Equal(t, 9, Sqr(-3))
	assert.InDelta(t, 2.25, Sqr(1.5), 0.0001)
}

func TestIsNearlyEqual(t *testing.T) {
	assert.True(t, IsNearlyEqual(1.0, 1.0))
	assert.True(t, IsNearlyEqual(1.0, 1.0005))
	assert.False(t, IsNearlyEqual(1.0, 1.01))
	assert.False(t, IsNearlyEqual(1.0, 2.0))
}

// TestAtan2ArgumentOrder pins the argument order, which used to be documented
// backwards. y comes first, as in the standard library.
func TestAtan2ArgumentOrder(t *testing.T) {
	// Straight up: y = 1, x = 0 is a quarter turn anticlockwise.
	assert.InDelta(t, math.Pi/2, Atan2(1.0, 0.0), 1e-12)
	// Straight right.
	assert.InDelta(t, 0, Atan2(0.0, 1.0), 1e-12)
	// Straight left.
	assert.InDelta(t, math.Pi, Atan2(0.0, -1.0), 1e-12)
	// The 45 degree diagonal, where a swapped order would be indistinguishable —
	// which is why the cases above are the ones that matter.
	assert.InDelta(t, math.Pi/4, Atan2(1.0, 1.0), 1e-12)

	// It must agree with the standard library for the same arguments.
	assert.InDelta(t, math.Atan2(3, -4), Atan2(3.0, -4.0), 1e-12)
}
