package gomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRay_PointAt(t *testing.T) {
	r := NewRay(Vector2[float64]{1, 1}, Vector2[float64]{1, 0})
	assert.Equal(t, Vector2[float64]{4, 1}, r.PointAt(3))
	assert.Equal(t, Vector2[float64]{1, 1}, r.PointAt(0))
}

func TestRay_Normalized(t *testing.T) {
	got := NewRay(Vector2[float64]{2, 2}, Vector2[float64]{0, 5}).Normalized()
	assert.Equal(t, Vector2[float64]{2, 2}, got.Origin)
	assert.InDelta(t, 0, got.Direction.X, 1e-9)
	assert.InDelta(t, 1, got.Direction.Y, 1e-9)
}

func TestRay_ClosestPointAndDistance(t *testing.T) {
	r := NewRay(Vector2[float64]{0, 0}, Vector2[float64]{1, 0})

	// Ahead of the origin: projected onto the ray.
	assert.Equal(t, Vector2[float64]{3, 0}, r.ClosestPoint(Vector2[float64]{3, 5}))
	assert.InDelta(t, 5, r.DistanceToPoint(Vector2[float64]{3, 5}), 1e-9)

	// Behind the origin: clamped to the origin.
	assert.Equal(t, Vector2[float64]{0, 0}, r.ClosestPoint(Vector2[float64]{-2, 3}))
	assert.InDelta(t, math.Hypot(2, 3), r.DistanceToPoint(Vector2[float64]{-2, 3}), 1e-9)
}

func TestRay_Equals(t *testing.T) {
	r := NewRay(Vector2[float64]{1, 2}, Vector2[float64]{1, 0})
	assert.True(t, r.Equals(NewRay(Vector2[float64]{1, 2}, Vector2[float64]{1, 0})))
	assert.False(t, r.Equals(NewRay(Vector2[float64]{1, 2}, Vector2[float64]{0, 1})))
}
