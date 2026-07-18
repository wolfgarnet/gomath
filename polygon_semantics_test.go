package gomath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// squarePolygon returns an axis-aligned square of the given half-size centred at
// (cx, cy).
func squarePolygon(cx, cy, half float64) Polygon[float64] {
	return NewPolygon([]Vector2[float64]{
		{cx - half, cy - half},
		{cx + half, cy - half},
		{cx + half, cy + half},
		{cx - half, cy + half},
	})
}

// Area must return the true (unsigned) area, not the raw shoelace sum (which is
// twice the signed area). This is RED against the original implementation.
func TestPolygon_AreaIsTrueArea(t *testing.T) {
	// 2x2 square -> area 4 (the raw shoelace sum would be 8).
	assert.InDelta(t, 4.0, squarePolygon(0, 0, 1).Area(), 1e-9)

	// Right triangle with legs 4 and 3 -> area 6 (raw sum would be 12).
	tri := NewPolygon([]Vector2[float64]{{0, 0}, {4, 0}, {0, 3}})
	assert.InDelta(t, 6.0, tri.Area(), 1e-9)
}

// Center2 (perimeter-weighted centroid) must actually accumulate. The original
// discarded the result of center.Add(...), so it returned the origin regardless
// of the polygon's location. RED against the original.
func TestPolygon_Center2Accumulates(t *testing.T) {
	sq := squarePolygon(5, 5, 1)
	got := sq.Center2()
	assert.InDelta(t, 5.0, got.X, 1e-9)
	assert.InDelta(t, 5.0, got.Y, 1e-9)
}

// TranslateToOrigin must actually move the points (the original discarded the
// result of Points[i].Sub(...)). After translating, the centroid sits at the
// origin and the returned vector is the amount removed. RED against the original.
func TestPolygon_TranslateToOriginMovesPoints(t *testing.T) {
	sq := squarePolygon(5, 5, 1)
	v := sq.TranslateToOrigin()

	assert.InDelta(t, 5.0, v.X, 1e-9)
	assert.InDelta(t, 5.0, v.Y, 1e-9)

	// Points are now centred on the origin.
	c := sq.Center2()
	assert.InDelta(t, 0.0, c.X, 1e-9)
	assert.InDelta(t, 0.0, c.Y, 1e-9)
}
