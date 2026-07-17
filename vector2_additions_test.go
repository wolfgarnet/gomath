package gomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- exact integer-safe operations ---

func TestVector2_Cross(t *testing.T) {
	assert.Equal(t, 1, Vector2[int]{1, 0}.Cross(Vector2[int]{0, 1}))
	assert.Equal(t, -1, Vector2[int]{0, 1}.Cross(Vector2[int]{1, 0}))
	assert.Equal(t, -2, Vector2[int]{2, 3}.Cross(Vector2[int]{4, 5}))
	assert.Equal(t, 0, Vector2[int]{2, 2}.Cross(Vector2[int]{3, 3})) // parallel
	// Cross is anti-symmetric.
	a, b := Vector2[float64]{2, 5}, Vector2[float64]{-3, 7}
	assert.InDelta(t, a.Cross(b), -b.Cross(a), 1e-9)
}

func TestVector2_Perp(t *testing.T) {
	// The CCW (left) perpendicular: Right -> Up.
	assert.Equal(t, Vector2[int]{0, 1}, Vector2[int]{1, 0}.Perp())
	assert.Equal(t, Vector2[int]{-4, 3}, Vector2[int]{3, 4}.Perp())
	// Perp is orthogonal to the original.
	v := Vector2[float64]{3, 4}
	assert.InDelta(t, 0, v.Dot(v.Perp()), 1e-9)
	// Matches the existing NormalCounterClockwise.
	assert.Equal(t, v.NormalCounterClockwise(), v.Perp())
}

func TestVector2_Negate(t *testing.T) {
	assert.Equal(t, Vector2[int]{-3, 4}, Vector2[int]{3, -4}.Negate())
	assert.Equal(t, Vector2[int]{0, 0}, Vector2[int]{0, 0}.Negate())
}

func TestVector2_Abs(t *testing.T) {
	assert.Equal(t, Vector2[int]{3, 4}, Vector2[int]{-3, -4}.Abs())
	assert.Equal(t, Vector2[float64]{1.5, 2.5}, Vector2[float64]{-1.5, 2.5}.Abs())
}

func TestVector2_MinMax(t *testing.T) {
	a, b := Vector2[int]{1, 5}, Vector2[int]{3, 2}
	assert.Equal(t, Vector2[int]{1, 2}, a.Min(b))
	assert.Equal(t, Vector2[int]{3, 5}, a.Max(b))
}

func TestVector2_Clamp(t *testing.T) {
	lo, hi := Vector2[int]{0, 0}, Vector2[int]{4, 4}
	assert.Equal(t, Vector2[int]{4, 0}, Vector2[int]{5, -3}.Clamp(lo, hi))
	assert.Equal(t, Vector2[int]{2, 3}, Vector2[int]{2, 3}.Clamp(lo, hi)) // inside
}

func TestVector2_ManhattanChebyshev(t *testing.T) {
	a, b := Vector2[int]{0, 0}, Vector2[int]{3, -4}
	assert.Equal(t, 7, a.ManhattanDistance(b))
	assert.Equal(t, 4, a.ChebyshevDistance(b))
	assert.Equal(t, a.ManhattanDistance(b), b.ManhattanDistance(a)) // symmetric
}

func TestVector2_FloorCeilRound(t *testing.T) {
	assert.Equal(t, Vector2[float64]{1, -2}, Vector2[float64]{1.7, -1.2}.Floor())
	assert.Equal(t, Vector2[float64]{2, -1}, Vector2[float64]{1.2, -1.7}.Ceil())
	assert.Equal(t, Vector2[float64]{2, -2}, Vector2[float64]{1.5, -1.5}.Round())
}

// --- float / geometric operations ---

func TestVector2_ClampLength(t *testing.T) {
	tests := map[string]struct {
		v        Vector2[float64]
		min, max float64
		want     Vector2[float64]
	}{
		"too long":    {Vector2[float64]{3, 4}, 0, 2, Vector2[float64]{1.2, 1.6}},
		"too short":   {Vector2[float64]{3, 4}, 10, 20, Vector2[float64]{6, 8}},
		"within":      {Vector2[float64]{3, 4}, 1, 10, Vector2[float64]{3, 4}},
		"zero vector": {Vector2[float64]{0, 0}, 1, 10, Vector2[float64]{0, 0}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.v.ClampLength(tt.min, tt.max)
			assert.InDelta(t, tt.want.X, got.X, 1e-9)
			assert.InDelta(t, tt.want.Y, got.Y, 1e-9)
		})
	}
}

func TestVector2_MoveTowards(t *testing.T) {
	// Partial step, capped by maxDelta.
	got := Vector2[float64]{0, 0}.MoveTowards(Vector2[float64]{10, 0}, 3)
	assert.InDelta(t, 3, got.X, 1e-9)
	assert.InDelta(t, 0, got.Y, 1e-9)

	// maxDelta exceeds the remaining distance: snap exactly to target.
	got = Vector2[float64]{0, 0}.MoveTowards(Vector2[float64]{10, 0}, 100)
	assert.Equal(t, Vector2[float64]{10, 0}, got)

	// Already at target.
	got = Vector2[float64]{5, 5}.MoveTowards(Vector2[float64]{5, 5}, 3)
	assert.Equal(t, Vector2[float64]{5, 5}, got)
}

func TestVector2_ProjectReject(t *testing.T) {
	v := Vector2[float64]{3, 4}
	onto := Vector2[float64]{1, 0}
	assert.Equal(t, Vector2[float64]{3, 0}, v.Project(onto))
	assert.Equal(t, Vector2[float64]{0, 4}, v.Reject(onto))

	// Non-unit onto must give the same projection.
	assert.InDelta(t, 3, v.Project(Vector2[float64]{2, 0}).X, 1e-9)

	// Projecting onto the zero vector yields the zero vector (no division blow-up).
	assert.Equal(t, Vector2[float64]{0, 0}, v.Project(Vector2[float64]{0, 0}))

	// Project + Reject reconstruct the original.
	p := v.Project(Vector2[float64]{1, 2})
	r := v.Reject(Vector2[float64]{1, 2})
	assert.InDelta(t, v.X, p.X+r.X, 1e-9)
	assert.InDelta(t, v.Y, p.Y+r.Y, 1e-9)
}

func TestVector2_Angle(t *testing.T) {
	assert.InDelta(t, 0, Vector2[float64]{1, 0}.Angle(), 1e-9)
	assert.InDelta(t, math.Pi/2, Vector2[float64]{0, 1}.Angle(), 1e-9)
	assert.InDelta(t, math.Pi, Vector2[float64]{-1, 0}.Angle(), 1e-9)
	assert.InDelta(t, -math.Pi/2, Vector2[float64]{0, -1}.Angle(), 1e-9)
	assert.InDelta(t, math.Pi/4, Vector2[float64]{1, 1}.Angle(), 1e-9)
}

func TestVector2_RotateAround(t *testing.T) {
	// Rotating (2,1) a quarter turn CCW about (1,1) gives (1,2).
	got := Vector2[float64]{2, 1}.RotateAround(Vector2[float64]{1, 1}, math.Pi/2)
	assert.InDelta(t, 1, got.X, 1e-9)
	assert.InDelta(t, 2, got.Y, 1e-9)

	// Rotating about the point itself is a no-op.
	got = Vector2[float64]{5, 7}.RotateAround(Vector2[float64]{5, 7}, 1.234)
	assert.InDelta(t, 5, got.X, 1e-9)
	assert.InDelta(t, 7, got.Y, 1e-9)
}

func TestVector2_IsNaNIsFinite(t *testing.T) {
	assert.True(t, Vector2[float64]{math.NaN(), 0}.IsNaN())
	assert.False(t, Vector2[float64]{1, 2}.IsNaN())

	assert.True(t, Vector2[float64]{1, 2}.IsFinite())
	assert.False(t, Vector2[float64]{math.Inf(1), 0}.IsFinite())
	assert.False(t, Vector2[float64]{math.NaN(), 0}.IsFinite())

	// Integer vectors are always finite and never NaN.
	assert.True(t, Vector2[int]{3, 4}.IsFinite())
	assert.False(t, Vector2[int]{3, 4}.IsNaN())
}

// --- free constructors ---

func TestVector2_FromAngle(t *testing.T) {
	got := FromAngle[float64](0)
	assert.InDelta(t, 1, got.X, 1e-9)
	assert.InDelta(t, 0, got.Y, 1e-9)

	got = FromAngle[float64](math.Pi / 2)
	assert.InDelta(t, 0, got.X, 1e-9)
	assert.InDelta(t, 1, got.Y, 1e-9)

	// FromAngle always produces a unit vector.
	assert.InDelta(t, 1, FromAngle[float64](1.234).Length(), 1e-9)
}

func TestVector2_FromPolar(t *testing.T) {
	got := FromPolar[float64](0, 5)
	assert.InDelta(t, 5, got.X, 1e-9)
	assert.InDelta(t, 0, got.Y, 1e-9)

	got = FromPolar[float64](math.Pi/2, 3)
	assert.InDelta(t, 0, got.X, 1e-9)
	assert.InDelta(t, 3, got.Y, 1e-9)

	// Angle round-trips.
	v := FromPolar[float64](0.7, 4)
	assert.InDelta(t, 0.7, v.Angle(), 1e-9)
	assert.InDelta(t, 4, v.Length(), 1e-9)
}

func TestVector2_Constructors(t *testing.T) {
	assert.Equal(t, Vector2[float64]{0, 0}, Zero[float64]())
	assert.Equal(t, Vector2[float64]{1, 1}, One[float64]())
	assert.Equal(t, Vector2[int]{0, 1}, Up[int]())
	assert.Equal(t, Vector2[int]{0, -1}, Down[int]())
	assert.Equal(t, Vector2[int]{-1, 0}, Left[int]())
	assert.Equal(t, Vector2[int]{1, 0}, Right[int]())
	assert.Equal(t, Vector2[int]{1, 0}, UnitX[int]())
	assert.Equal(t, Vector2[int]{0, 1}, UnitY[int]())

	// Up/Right agree with the existing float32 package globals.
	assert.Equal(t, VectorUp32, Up[float32]())
	assert.Equal(t, VectorRight32, Right[float32]())
}
