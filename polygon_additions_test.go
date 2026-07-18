package gomath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolygon_SignedArea(t *testing.T) {
	// CCW square -> positive signed area.
	assert.InDelta(t, 4.0, squarePolygon(0, 0, 1).SignedArea(), 1e-9)

	// Reversing the winding negates the signed area but not the area.
	sq := squarePolygon(0, 0, 1)
	rev := Polygon[float64]{Points: []Vector2[float64]{sq.Points[3], sq.Points[2], sq.Points[1], sq.Points[0]}}
	assert.InDelta(t, -4.0, rev.SignedArea(), 1e-9)
	assert.InDelta(t, 4.0, rev.Area(), 1e-9)
}

func TestPolygon_Centroid(t *testing.T) {
	// Square centroid is its centre.
	c := squarePolygon(2, 3, 1).Centroid()
	assert.InDelta(t, 2.0, c.X, 1e-9)
	assert.InDelta(t, 3.0, c.Y, 1e-9)

	// Triangle centroid is the average of the vertices.
	tri := NewPolygon([]Vector2[float64]{{0, 0}, {4, 0}, {0, 3}})
	tc := tri.Centroid()
	assert.InDelta(t, 4.0/3.0, tc.X, 1e-9)
	assert.InDelta(t, 1.0, tc.Y, 1e-9)
}

func TestPolygon_ContainsPoint(t *testing.T) {
	sq := squarePolygon(0, 0, 2) // [-2,2]^2
	assert.True(t, sq.ContainsPoint(Vector2[float64]{0, 0}))
	assert.True(t, sq.ContainsPoint(Vector2[float64]{1.5, -1.5}))
	assert.False(t, sq.ContainsPoint(Vector2[float64]{3, 0}))
	assert.False(t, sq.ContainsPoint(Vector2[float64]{0, 5}))

	// Concave "L" shape: a point in the notch is outside.
	l := NewPolygon([]Vector2[float64]{
		{0, 0}, {4, 0}, {4, 2}, {2, 2}, {2, 4}, {0, 4},
	})
	assert.True(t, l.ContainsPoint(Vector2[float64]{1, 1}))
	assert.False(t, l.ContainsPoint(Vector2[float64]{3, 3})) // in the missing corner
}

func TestPolygon_BoundingBoxPerimeter(t *testing.T) {
	sq := squarePolygon(0, 0, 1) // [-1,1]^2
	bb := sq.BoundingBox()
	assert.Equal(t, Vector2[float64]{-1, -1}, bb.Min())
	assert.Equal(t, Vector2[float64]{1, 1}, bb.Max())

	assert.InDelta(t, 8.0, sq.Perimeter(), 1e-9)
}

func TestPolygon_TranslateCloneEdgeSegment(t *testing.T) {
	sq := squarePolygon(0, 0, 1)

	moved := sq.Translate(Vector2[float64]{10, 20})
	assert.InDelta(t, 10.0, moved.Centroid().X, 1e-9)
	assert.InDelta(t, 20.0, moved.Centroid().Y, 1e-9)
	// Translate does not mutate the original.
	assert.InDelta(t, 0.0, sq.Centroid().X, 1e-9)

	// Clone is an independent copy.
	clone := sq.Clone()
	clone.Points[0] = Vector2[float64]{99, 99}
	assert.NotEqual(t, sq.Points[0], clone.Points[0])

	// EdgeSegment returns edge i as a LineSegment.
	a, b := sq.Edge(0)
	seg := sq.EdgeSegment(0)
	assert.Equal(t, NewLineSegment(a, b), seg)
}
