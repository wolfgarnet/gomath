package gomath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- exact, integer-safe ---

func TestLineSegment_MidpointLengthSq(t *testing.T) {
	s := NewLineSegmentXY[int](0, 0, 10, 4)
	assert.Equal(t, Vector2[int]{5, 2}, s.Midpoint())
	assert.Equal(t, 25, NewLineSegmentXY[int](0, 0, 3, 4).LengthSquared())
}

func TestLineSegment_ReversedTranslate(t *testing.T) {
	s := NewLineSegmentXY[int](1, 2, 3, 4)
	assert.Equal(t, NewLineSegmentXY[int](3, 4, 1, 2), s.Reversed())
	assert.Equal(t, NewLineSegmentXY[int](3, 5, 5, 7), s.Translate(Vector2[int]{2, 3}))
}

func TestLineSegment_BoundingBoxEquals(t *testing.T) {
	// float64 so the odd-span box is represented exactly (integer AABB centres
	// truncate on odd spans - the documented AABB caveat).
	s := NewLineSegmentXY[float64](-1, 3, 4, -2)
	bb := s.BoundingBox()
	assert.Equal(t, Vector2[float64]{-1, -2}, bb.Min())
	assert.Equal(t, Vector2[float64]{4, 3}, bb.Max())

	si := NewLineSegmentXY[int](1, 2, 3, 4)
	assert.True(t, si.Equals(NewLineSegmentXY[int](1, 2, 3, 4)))
	assert.False(t, si.Equals(si.Reversed()))
}

// --- float geometry ---

func TestLineSegment_PointAt(t *testing.T) {
	s := NewLineSegmentXY[float64](0, 0, 10, 0)
	assert.Equal(t, Vector2[float64]{0, 0}, s.PointAt(0))
	assert.Equal(t, Vector2[float64]{5, 0}, s.PointAt(0.5))
	assert.Equal(t, Vector2[float64]{10, 0}, s.PointAt(1))
}

func TestLineSegment_UnitDirection(t *testing.T) {
	got := NewLineSegmentXY[float64](0, 0, 0, 5).UnitDirection()
	assert.InDelta(t, 0, got.X, 1e-9)
	assert.InDelta(t, 1, got.Y, 1e-9)
}

func TestLineSegment_ClosestPointAndDistance(t *testing.T) {
	s := NewLineSegmentXY[float64](0, 0, 2, 0)
	assert.Equal(t, Vector2[float64]{1, 0}, s.ClosestPoint(Vector2[float64]{1, 1})) // above the middle
	assert.Equal(t, Vector2[float64]{2, 0}, s.ClosestPoint(Vector2[float64]{3, 0})) // past the end
	assert.Equal(t, Vector2[float64]{0, 0}, s.ClosestPoint(Vector2[float64]{-1, 0}))

	assert.InDelta(t, 1, s.DistanceToPoint(Vector2[float64]{1, 1}), 1e-9)
	assert.InDelta(t, 1, s.DistanceToPointSquared(Vector2[float64]{1, 1}), 1e-9)
}
