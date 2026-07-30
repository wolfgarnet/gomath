package gomath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func box(cx, cy, hw, hh float64) AABB[float64] {
	return NewAABB(Vector2[float64]{cx, cy}, hw, hh)
}

// --- exact, integer-safe queries ---

func TestAABB_MinMaxSizeArea(t *testing.T) {
	a := NewAABB(Vector2[int]{10, 20}, 2, 3) // spans [8,12] x [17,23]
	assert.Equal(t, Vector2[int]{8, 17}, a.Min())
	assert.Equal(t, Vector2[int]{12, 23}, a.Max())
	assert.Equal(t, 4, a.Width())
	assert.Equal(t, 6, a.Height())
	assert.Equal(t, Vector2[int]{4, 6}, a.Size())
	assert.Equal(t, 24, a.Area())
}

func TestAABB_Contains(t *testing.T) {
	a := NewAABB(Vector2[int]{0, 0}, 2, 2) // [-2,2]^2
	assert.True(t, a.Contains(Vector2[int]{1, 1}))
	assert.True(t, a.Contains(Vector2[int]{2, 2}))  // edge is inclusive
	assert.True(t, a.Contains(Vector2[int]{-2, 0})) // edge is inclusive
	assert.False(t, a.Contains(Vector2[int]{3, 0}))
	assert.False(t, a.Contains(Vector2[int]{0, 3}))
}

func TestAABB_ContainsAABB(t *testing.T) {
	outer := NewAABB(Vector2[int]{0, 0}, 5, 5)
	assert.True(t, outer.ContainsAABB(NewAABB(Vector2[int]{0, 0}, 1, 1)))
	assert.True(t, outer.ContainsAABB(outer)) // contains itself
	assert.False(t, outer.ContainsAABB(NewAABB(Vector2[int]{4, 0}, 2, 2)))
}

func TestAABB_Intersects(t *testing.T) {
	a := NewAABB(Vector2[int]{0, 0}, 2, 2)                           // [-2,2]^2
	assert.True(t, a.Intersects(NewAABB(Vector2[int]{3, 0}, 2, 2)))  // overlap on x
	assert.True(t, a.Intersects(NewAABB(Vector2[int]{4, 0}, 2, 2)))  // edge touch (inclusive)
	assert.False(t, a.Intersects(NewAABB(Vector2[int]{5, 0}, 2, 2))) // gap
	assert.True(t, a.Intersects(a))
}

func TestAABB_Equals(t *testing.T) {
	assert.True(t, NewAABB(Vector2[int]{1, 2}, 3, 4).Equals(NewAABB(Vector2[int]{1, 2}, 3, 4)))
	assert.False(t, NewAABB(Vector2[int]{1, 2}, 3, 4).Equals(NewAABB(Vector2[int]{1, 2}, 3, 5)))
}

// --- float geometry ---

func TestAABB_Intersection(t *testing.T) {
	a := box(0, 0, 2, 2) // [-2,2]^2
	b := box(3, 0, 2, 2) // [1,5] x [-2,2]

	got, ok := a.Intersection(b)
	assert.True(t, ok)
	assert.Equal(t, box(1.5, 0, 0.5, 2), got)

	_, ok = a.Intersection(box(10, 0, 1, 1))
	assert.False(t, ok)
}

func TestAABB_Union(t *testing.T) {
	a := box(0, 0, 2, 2) // [-2,2]^2
	b := box(3, 0, 2, 2) // [1,5] x [-2,2]
	// Union spans [-2,5] x [-2,2].
	assert.Equal(t, box(1.5, 0, 3.5, 2), a.Union(b))
}

func TestAABB_ClosestPoint(t *testing.T) {
	a := box(0, 0, 2, 2)                                                                // [-2,2]^2
	assert.Equal(t, Vector2[float64]{2, 1}, a.ClosestPoint(Vector2[float64]{5, 1}))     // clamped x
	assert.Equal(t, Vector2[float64]{-2, -2}, a.ClosestPoint(Vector2[float64]{-9, -3})) // clamped both
	assert.Equal(t, Vector2[float64]{1, 1}, a.ClosestPoint(Vector2[float64]{1, 1}))     // inside: unchanged
}

func TestAABB_ExpandToInclude(t *testing.T) {
	a := box(0, 0, 2, 2) // [-2,2]^2
	// Growing to include (5,5): new box spans [-2,5] x [-2,5].
	assert.Equal(t, box(1.5, 1.5, 3.5, 3.5), a.ExpandToInclude(Vector2[float64]{5, 5}))
	// A point already inside leaves the box unchanged.
	assert.Equal(t, a, a.ExpandToInclude(Vector2[float64]{0, 1}))
}

// --- constructors ---

func TestAABB_NewFromMinMax(t *testing.T) {
	a := NewAABBFromMinMax(Vector2[float64]{-2, -3}, Vector2[float64]{4, 5})
	assert.Equal(t, Vector2[float64]{1, 1}, a.Center)
	assert.Equal(t, 3.0, a.HalfWidth)
	assert.Equal(t, 4.0, a.HalfHeight)
	// Round-trips through Min/Max.
	assert.Equal(t, Vector2[float64]{-2, -3}, a.Min())
	assert.Equal(t, Vector2[float64]{4, 5}, a.Max())
}

func TestAABB_NewFromPoints(t *testing.T) {
	a := NewAABBFromPoints(
		Vector2[float64]{1, 2},
		Vector2[float64]{-3, 5},
		Vector2[float64]{4, -1},
	)
	// Bounding box spans [-3,4] x [-1,5].
	assert.Equal(t, Vector2[float64]{-3, -1}, a.Min())
	assert.Equal(t, Vector2[float64]{4, 5}, a.Max())
	// Every source point is contained.
	assert.True(t, a.Contains(Vector2[float64]{1, 2}))
	assert.True(t, a.Contains(Vector2[float64]{-3, 5}))
	assert.True(t, a.Contains(Vector2[float64]{4, -1}))
}
