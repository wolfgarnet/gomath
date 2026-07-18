package gomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- exact integer-safe operations ---

func TestVector3_Cross(t *testing.T) {
	// Right-handed basis: X x Y = Z, Y x Z = X, Z x X = Y.
	assert.Equal(t, Vector3[int]{0, 0, 1}, Vector3[int]{1, 0, 0}.Cross(Vector3[int]{0, 1, 0}))
	assert.Equal(t, Vector3[int]{1, 0, 0}, Vector3[int]{0, 1, 0}.Cross(Vector3[int]{0, 0, 1}))
	assert.Equal(t, Vector3[int]{0, 1, 0}, Vector3[int]{0, 0, 1}.Cross(Vector3[int]{1, 0, 0}))

	// Anti-commutative, and orthogonal to both operands.
	a, b := Vector3[float64]{2, 5, -1}, Vector3[float64]{-3, 7, 4}
	c := a.Cross(b)
	assert.Equal(t, b.Cross(a).Negate(), c)
	assert.InDelta(t, 0, a.Dot(c), 1e-9)
	assert.InDelta(t, 0, b.Dot(c), 1e-9)

	// Parallel vectors cross to zero.
	assert.Equal(t, Vector3[int]{0, 0, 0}, Vector3[int]{2, 2, 2}.Cross(Vector3[int]{3, 3, 3}))
}

func TestVector3_Negate(t *testing.T) {
	assert.Equal(t, Vector3[int]{-1, 2, -3}, Vector3[int]{1, -2, 3}.Negate())
}

func TestVector3_Abs(t *testing.T) {
	assert.Equal(t, Vector3[int]{1, 2, 3}, Vector3[int]{-1, -2, -3}.Abs())
}

func TestVector3_MinMax(t *testing.T) {
	a, b := Vector3[int]{1, 5, 3}, Vector3[int]{3, 2, 4}
	assert.Equal(t, Vector3[int]{1, 2, 3}, a.Min(b))
	assert.Equal(t, Vector3[int]{3, 5, 4}, a.Max(b))
}

func TestVector3_Clamp(t *testing.T) {
	lo, hi := Vector3[int]{0, 0, 0}, Vector3[int]{4, 4, 4}
	assert.Equal(t, Vector3[int]{4, 0, 2}, Vector3[int]{5, -3, 2}.Clamp(lo, hi))
}

func TestVector3_ManhattanChebyshev(t *testing.T) {
	a, b := Vector3[int]{0, 0, 0}, Vector3[int]{3, -4, 5}
	assert.Equal(t, 12, a.ManhattanDistance(b))
	assert.Equal(t, 5, a.ChebyshevDistance(b))
	assert.Equal(t, a.ManhattanDistance(b), b.ManhattanDistance(a))
}

func TestVector3_FloorCeilRound(t *testing.T) {
	assert.Equal(t, Vector3[float64]{1, -2, 2}, Vector3[float64]{1.7, -1.2, 2.9}.Floor())
	assert.Equal(t, Vector3[float64]{2, -1, 3}, Vector3[float64]{1.2, -1.7, 2.1}.Ceil())
	assert.Equal(t, Vector3[float64]{2, -2, 2}, Vector3[float64]{1.5, -1.5, 2.4}.Round())
}

func TestVector3_IsZero(t *testing.T) {
	assert.True(t, Vector3[int]{0, 0, 0}.IsZero())
	assert.False(t, Vector3[int]{0, 0, 1}.IsZero())
}

// --- float / geometric operations ---

func TestVector3_ClampLength(t *testing.T) {
	tests := map[string]struct {
		v        Vector3[float64]
		min, max float64
		want     Vector3[float64]
	}{
		"too long":    {Vector3[float64]{3, 4, 0}, 0, 2, Vector3[float64]{1.2, 1.6, 0}},
		"too short":   {Vector3[float64]{3, 4, 0}, 10, 20, Vector3[float64]{6, 8, 0}},
		"within":      {Vector3[float64]{3, 4, 0}, 1, 10, Vector3[float64]{3, 4, 0}},
		"zero vector": {Vector3[float64]{0, 0, 0}, 1, 10, Vector3[float64]{0, 0, 0}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.v.ClampLength(tt.min, tt.max)
			assert.InDelta(t, tt.want.X, got.X, 1e-9)
			assert.InDelta(t, tt.want.Y, got.Y, 1e-9)
			assert.InDelta(t, tt.want.Z, got.Z, 1e-9)
		})
	}
}

func TestVector3_MoveTowards(t *testing.T) {
	got := Vector3[float64]{0, 0, 0}.MoveTowards(Vector3[float64]{10, 0, 0}, 3)
	assert.InDelta(t, 3, got.X, 1e-9)

	got = Vector3[float64]{0, 0, 0}.MoveTowards(Vector3[float64]{10, 0, 0}, 100)
	assert.Equal(t, Vector3[float64]{10, 0, 0}, got)

	got = Vector3[float64]{5, 5, 5}.MoveTowards(Vector3[float64]{5, 5, 5}, 3)
	assert.Equal(t, Vector3[float64]{5, 5, 5}, got)
}

func TestVector3_ProjectReject(t *testing.T) {
	v := Vector3[float64]{3, 4, 5}
	onto := Vector3[float64]{1, 0, 0}
	assert.Equal(t, Vector3[float64]{3, 0, 0}, v.Project(onto))
	assert.Equal(t, Vector3[float64]{0, 4, 5}, v.Reject(onto))

	// Projecting onto the zero vector yields the zero vector.
	assert.Equal(t, Vector3[float64]{0, 0, 0}, v.Project(Vector3[float64]{0, 0, 0}))

	// Project + Reject reconstruct the original.
	p := v.Project(Vector3[float64]{1, 2, 3})
	r := v.Reject(Vector3[float64]{1, 2, 3})
	assert.InDelta(t, v.X, p.X+r.X, 1e-9)
	assert.InDelta(t, v.Y, p.Y+r.Y, 1e-9)
	assert.InDelta(t, v.Z, p.Z+r.Z, 1e-9)
}

func TestVector3_ScaleToLength(t *testing.T) {
	assert.Equal(t, Vector3[float64]{6, 8, 0}, Vector3[float64]{3, 4, 0}.ScaleToLength(10))
	// Length 0 leaves the vector unchanged (mirrors Vector2.ScaleToLength).
	assert.Equal(t, Vector3[float64]{3, 4, 0}, Vector3[float64]{3, 4, 0}.ScaleToLength(0))
	assert.Equal(t, Vector3[float64]{0, 0, 0}, Vector3[float64]{0, 0, 0}.ScaleToLength(5))
}

func TestVector3_DirectionTo(t *testing.T) {
	got := Vector3[float64]{0, 0, 0}.DirectionTo(Vector3[float64]{0, 0, 5})
	assert.InDelta(t, 0, got.X, 1e-9)
	assert.InDelta(t, 0, got.Y, 1e-9)
	assert.InDelta(t, 1, got.Z, 1e-9)
}

func TestVector3_IsNaNIsFinite(t *testing.T) {
	assert.True(t, Vector3[float64]{math.NaN(), 0, 0}.IsNaN())
	assert.False(t, Vector3[float64]{1, 2, 3}.IsNaN())

	assert.True(t, Vector3[float64]{1, 2, 3}.IsFinite())
	assert.False(t, Vector3[float64]{math.Inf(1), 0, 0}.IsFinite())
	assert.False(t, Vector3[float64]{0, math.NaN(), 0}.IsFinite())

	assert.True(t, Vector3[int]{1, 2, 3}.IsFinite())
	assert.False(t, Vector3[int]{1, 2, 3}.IsNaN())
}

func TestVector3_RotateAxis(t *testing.T) {
	// Rotating X a quarter turn CCW about Z gives Y.
	got := Vector3[float64]{1, 0, 0}.RotateAxis(Vector3[float64]{0, 0, 1}, math.Pi/2)
	assert.InDelta(t, 0, got.X, 1e-9)
	assert.InDelta(t, 1, got.Y, 1e-9)
	assert.InDelta(t, 0, got.Z, 1e-9)

	// A vector parallel to the axis is unchanged.
	got = Vector3[float64]{0, 0, 3}.RotateAxis(Vector3[float64]{0, 0, 1}, 1.234)
	assert.InDelta(t, 0, got.X, 1e-9)
	assert.InDelta(t, 0, got.Y, 1e-9)
	assert.InDelta(t, 3, got.Z, 1e-9)

	// Rotation preserves length.
	v := Vector3[float64]{1, 2, 2} // length 3
	r := v.RotateAxis(Vector3[float64]{0, 1, 0}, 0.7)
	assert.InDelta(t, 3, r.Length(), 1e-9)
}

// --- free constructors ---

func TestVector3_Constructors(t *testing.T) {
	assert.Equal(t, Vector3[float64]{0, 0, 0}, Zero3[float64]())
	assert.Equal(t, Vector3[float64]{1, 1, 1}, One3[float64]())
	assert.Equal(t, Vector3[int]{0, 1, 0}, Up3[int]())
	assert.Equal(t, Vector3[int]{0, -1, 0}, Down3[int]())
	assert.Equal(t, Vector3[int]{-1, 0, 0}, Left3[int]())
	assert.Equal(t, Vector3[int]{1, 0, 0}, Right3[int]())
	assert.Equal(t, Vector3[int]{0, 0, 1}, Forward3[int]())
	assert.Equal(t, Vector3[int]{0, 0, -1}, Back3[int]())
	assert.Equal(t, Vector3[int]{1, 0, 0}, UnitX3[int]())
	assert.Equal(t, Vector3[int]{0, 1, 0}, UnitY3[int]())
	assert.Equal(t, Vector3[int]{0, 0, 1}, UnitZ3[int]())
}
