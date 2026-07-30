package gomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- exact, integer-safe queries ---

func TestCircle_Contains(t *testing.T) {
	c := NewCircle[int](0, 0, 5)
	assert.True(t, c.Contains(Vector2[int]{3, 4}))  // on the boundary (dist 5)
	assert.True(t, c.Contains(Vector2[int]{5, 0}))  // boundary, inclusive
	assert.True(t, c.Contains(Vector2[int]{0, 0}))  // centre
	assert.False(t, c.Contains(Vector2[int]{4, 4})) // dist^2 = 32 > 25
}

func TestCircle_ContainsCircle(t *testing.T) {
	big := NewCircle[int](0, 0, 5)
	assert.True(t, big.ContainsCircle(NewCircle[int](2, 0, 1)))  // fully inside
	assert.True(t, big.ContainsCircle(big))                      // contains itself
	assert.False(t, big.ContainsCircle(NewCircle[int](5, 0, 1))) // pokes out
	assert.False(t, big.ContainsCircle(NewCircle[int](0, 0, 6))) // larger radius
}

func TestCircle_Intersects(t *testing.T) {
	a := NewCircle[int](0, 0, 2)
	assert.True(t, a.Intersects(NewCircle[int](3, 0, 2)))  // overlap
	assert.True(t, a.Intersects(NewCircle[int](4, 0, 2)))  // touch (inclusive)
	assert.False(t, a.Intersects(NewCircle[int](5, 0, 2))) // gap
}

func TestCircle_IntersectsAABB(t *testing.T) {
	c := NewCircle[int](0, 0, 2)
	assert.True(t, c.IntersectsAABB(NewAABB(Vector2[int]{3, 0}, 1, 1)))  // touch at x=2
	assert.False(t, c.IntersectsAABB(NewAABB(Vector2[int]{4, 0}, 1, 1))) // closest at x=3, gap
	assert.True(t, c.IntersectsAABB(NewAABB(Vector2[int]{0, 0}, 5, 5)))  // box contains centre
}

func TestCircle_DiameterBoundingBox(t *testing.T) {
	c := NewCircle[int](1, 2, 3)
	assert.Equal(t, 6, c.Diameter())
	bb := c.BoundingBox()
	assert.Equal(t, NewAABB(Vector2[int]{1, 2}, 3, 3), bb)
	assert.Equal(t, Vector2[int]{-2, -1}, bb.Min())
	assert.Equal(t, Vector2[int]{4, 5}, bb.Max())
}

func TestCircle_Translate(t *testing.T) {
	c := NewCircle[int](1, 1, 2).Translate(Vector2[int]{3, 4})
	assert.Equal(t, NewCircle[int](4, 5, 2), c)
}

func TestCircle_Equals(t *testing.T) {
	assert.True(t, NewCircle[int](1, 2, 3).Equals(NewCircle[int](1, 2, 3)))
	assert.False(t, NewCircle[int](1, 2, 3).Equals(NewCircle[int](1, 2, 4)))
}

// --- float geometry ---

func TestCircle_ClosestPoint(t *testing.T) {
	c := NewCircle[float64](0, 0, 5)
	// Outside: projected onto the boundary.
	got := c.ClosestPoint(Vector2[float64]{10, 0})
	assert.InDelta(t, 5, got.X, 1e-9)
	assert.InDelta(t, 0, got.Y, 1e-9)

	got = c.ClosestPoint(Vector2[float64]{6, 8}) // dist 10, dir (0.6,0.8)
	assert.InDelta(t, 3, got.X, 1e-9)
	assert.InDelta(t, 4, got.Y, 1e-9)

	// Inside: the point itself.
	assert.Equal(t, Vector2[float64]{1, 2}, c.ClosestPoint(Vector2[float64]{1, 2}))
}

func TestCircle_AreaCircumference(t *testing.T) {
	c := NewCircle[float64](0, 0, 2)
	assert.InDelta(t, math.Pi*4, c.Area(), 1e-9)
	assert.InDelta(t, math.Pi*4, c.Circumference(), 1e-9)
}
