package intersection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
)

// TestEllipsoidEllipsoid_TranslationInvariant asserts that translating both
// ellipsoids by the same offset does not change whether they overlap.
// EllipsoidEllipsoid previously left e1.Position in world space while
// transforming e2 into e1's space, so the result was only correct with e1 at
// the origin -- which every original test case happened to use.
func TestEllipsoidEllipsoid_TranslationInvariant(t *testing.T) {
	pairs := map[string]struct {
		e1, e2   gomath.Ellipsoid[float64]
		overlaps bool
	}{
		"coincident unit circles": {
			e1:       newEllipsoid64(0, 0, 1, 1),
			e2:       newEllipsoid64(0, 0, 1, 1),
			overlaps: true,
		},
		"touching unit circles": {
			e1:       newEllipsoid64(0, 0, 1, 1),
			e2:       newEllipsoid64(1, 0, 1, 1),
			overlaps: true,
		},
		"separated unit circles": {
			e1:       newEllipsoid64(0, 0, 1, 1),
			e2:       newEllipsoid64(3, 0, 1, 1),
			overlaps: false,
		},
		"wide ellipse vs circle, apart": {
			e1:       newEllipsoid64(0, 0, 3, 1),
			e2:       newEllipsoid64(3, 3, 1, 1),
			overlaps: false,
		},
	}

	offsets := []gomath.Vector2[float64]{
		{X: 0, Y: 0},
		{X: 100, Y: 100},
		{X: -50, Y: 25},
		{X: 0.5, Y: -0.5},
	}

	for name, pair := range pairs {
		t.Run(name, func(t *testing.T) {
			for _, offset := range offsets {
				e1 := translateEllipsoid(pair.e1, offset)
				e2 := translateEllipsoid(pair.e2, offset)

				assert.Equalf(t, pair.overlaps, EllipsoidEllipsoid(e1, e2),
					"offset %v: %v vs %v", offset, e1, e2)
			}
		})
	}
}

// TestPolygonPolygon_AllEdgesAreSeparatingAxes asserts every edge of the second
// polygon is used as a separating axis. The second loop previously stopped one
// edge early, so a shape separated only along that edge reported an overlap.
func TestPolygonPolygon_AllEdgesAreSeparatingAxes(t *testing.T) {
	// Triangle whose hypotenuse runs from (10,0) to (5,5), i.e. the line x+y=10.
	triangle := gomath.NewPolygon([]gomath.Vector2[float64]{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 5}})

	tests := map[string]struct {
		square   gomath.Polygon[float64]
		overlaps bool
	}{
		// min x+y = 11, so it clears the hypotenuse -- but it still overlaps the
		// triangle on the x axis, the y axis and both other edge normals, making
		// the hypotenuse the only axis that separates the two shapes.
		"separated only by the hypotenuse": {
			square: gomath.NewPolygon([]gomath.Vector2[float64]{
				{X: 7.5, Y: 3.5}, {X: 9.5, Y: 3.5}, {X: 9.5, Y: 5.5}, {X: 7.5, Y: 5.5},
			}),
			overlaps: false,
		},
		"well inside the triangle": {
			square: gomath.NewPolygon([]gomath.Vector2[float64]{
				{X: 4, Y: 1}, {X: 6, Y: 1}, {X: 6, Y: 3}, {X: 4, Y: 3},
			}),
			overlaps: true,
		},
		"far away": {
			square: gomath.NewPolygon([]gomath.Vector2[float64]{
				{X: 100, Y: 100}, {X: 102, Y: 100}, {X: 102, Y: 102}, {X: 100, Y: 102},
			}),
			overlaps: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, _, overlapping := PolygonPolygon(test.square, triangle)
			assert.Equal(t, test.overlaps, overlapping)

			// The predicate must not depend on argument order.
			_, _, reversed := PolygonPolygon(triangle, test.square)
			assert.Equal(t, test.overlaps, reversed, "argument order changed the result")
		})
	}
}

// TestLineBox_SegmentInside asserts a segment lying entirely within the box counts
// as intersecting it, matching the SquaredAABB path (LineSquaredBoxContains).
func TestLineBox_SegmentInside(t *testing.T) {
	box := gomath.AABB[float64]{Center: gomath.Vector2[float64]{X: 0, Y: 0}, HalfWidth: 10, HalfHeight: 10}

	tests := map[string]struct {
		segment  gomath.LineSegment[float64]
		expected bool
	}{
		"fully inside":   {segment: gomath.NewLineSegmentXY[float64](-1, -1, 1, 1), expected: true},
		"crossing":       {segment: gomath.NewLineSegmentXY[float64](-20, 0, 20, 0), expected: true},
		"one end inside": {segment: gomath.NewLineSegmentXY[float64](0, 0, 20, 0), expected: true},
		"fully outside":  {segment: gomath.NewLineSegmentXY[float64](50, 50, 60, 60), expected: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, LineBox(test.segment, box))

			// The generic dispatch must agree with the concrete function.
			assert.Equal(t, test.expected, IntersectingAABB[float64](box, test.segment))
		})
	}
}

// TestLineBoxSide_NoIntersection asserts LineBoxSide reports "no side" as nil
// rather than as a pointer to a zero-value segment, which a caller cannot
// distinguish from a real segment lying at the origin.
func TestLineBoxSide_NoIntersection(t *testing.T) {
	box := gomath.AABB[float64]{Center: gomath.Vector2[float64]{X: 0, Y: 0}, HalfWidth: 1, HalfHeight: 1}

	t.Run("no intersection", func(t *testing.T) {
		side := LineBoxSide(gomath.NewLineSegmentXY[float64](100, 100, 200, 200), box)
		assert.Nil(t, side)
	})

	t.Run("intersects", func(t *testing.T) {
		// Comes in from the right, so it must hit the x=+1 side.
		side := LineBoxSide(gomath.NewLineSegmentXY[float64](5, 0, 0, 0), box)
		if assert.NotNil(t, side) {
			assert.InDelta(t, 1.0, side.P1.X, 0.0001)
			assert.InDelta(t, 1.0, side.P2.X, 0.0001)
		}
	})
}

// TestIntersectingRay2_NearestHit asserts the reported distance is to the nearest
// intersection with a circle, not the farthest.
func TestIntersectingRay2_NearestHit(t *testing.T) {
	ray := gomath.NewRay(gomath.Vector2[float64]{X: 0, Y: 0}, gomath.Vector2[float64]{X: 1, Y: 0})

	tests := map[string]struct {
		circle   gomath.Circle[float64]
		hits     bool
		distance float64
	}{
		// Spans x in [8,12] along the ray.
		"ahead":    {circle: gomath.NewCircle[float64](10, 0, 2), hits: true, distance: 8},
		"tangent":  {circle: gomath.NewCircle[float64](10, 2, 2), hits: true, distance: 10},
		"off axis": {circle: gomath.NewCircle[float64](10, 5, 2), hits: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			hits, distance := IntersectingRay2[float64](ray, test.circle)
			assert.Equal(t, test.hits, hits)
			if test.hits {
				assert.InDelta(t, test.distance, distance, 0.001)
			}
		})
	}
}

// TestRayCircle_HalfLine asserts a ray is treated as a half-line: geometry behind
// the origin must not count as an intersection.
func TestRayCircle_HalfLine(t *testing.T) {
	ray := gomath.NewRay(gomath.Vector2[float64]{X: 0, Y: 0}, gomath.Vector2[float64]{X: 1, Y: 0})

	tests := map[string]struct {
		circle       gomath.Circle[float64]
		intersations int
	}{
		"entirely behind":       {circle: gomath.NewCircle[float64](-10, 0, 2), intersations: 0},
		"entirely ahead":        {circle: gomath.NewCircle[float64](10, 0, 2), intersations: 2},
		"straddling the origin": {circle: gomath.NewCircle[float64](0, 0, 2), intersations: 1},
		"beside the ray":        {circle: gomath.NewCircle[float64](10, 50, 2), intersations: 0},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			n, p1, p2 := RayCircle(ray, test.circle)
			assert.Equal(t, test.intersations, n)

			// Any reported hit must lie ahead of the origin.
			for i, p := range []gomath.Vector2[float64]{p1, p2} {
				if i >= n {
					break
				}
				assert.GreaterOrEqualf(t, p.Sub(ray.Origin).Dot(ray.Direction), -0.0001,
					"intersection %v is behind the ray origin", p)
			}
		})
	}
}

// TestAABBInsideCircleColliders asserts the function tests what its name says.
// It was a verbatim copy of CircleInsideAABBColliders.
func TestAABBInsideCircleColliders(t *testing.T) {
	origin := gomath.Vector2[float64]{X: 0, Y: 0}

	t.Run("small box inside big circle", func(t *testing.T) {
		circle := gomath.CircleCollider[float64]{Radius: 100}
		box := gomath.AABBCollider[float64]{HalfWidth: 1, HalfHeight: 1}

		assert.True(t, AABBInsideCircleColliders(circle, origin, box, origin))
		assert.False(t, CircleInsideAABBColliders(circle, origin, box, origin))
	})

	t.Run("big box outside small circle", func(t *testing.T) {
		circle := gomath.CircleCollider[float64]{Radius: 1}
		box := gomath.AABBCollider[float64]{HalfWidth: 100, HalfHeight: 100}

		assert.False(t, AABBInsideCircleColliders(circle, origin, box, origin))
		assert.True(t, CircleInsideAABBColliders(circle, origin, box, origin))
	})

	t.Run("box corners poke out of circle", func(t *testing.T) {
		// Half-diagonal is sqrt(2)*10 = 14.1, so the corners fall outside r=11.
		circle := gomath.CircleCollider[float64]{Radius: 11}
		box := gomath.AABBCollider[float64]{HalfWidth: 10, HalfHeight: 10}

		assert.False(t, AABBInsideCircleColliders(circle, origin, box, origin))
	})
}

// TestCircleAABBWithPenetration2_ContainedNormal asserts that an overlapping
// result always carries a usable unit normal. A circle further from every wall
// than its own radius previously returned a zero normal.
func TestCircleAABBWithPenetration2_ContainedNormal(t *testing.T) {
	// Box spans x in [0,1024], y in [0,96]. The circle sits well inside it,
	// 32 from the nearest wall (y=96) but only 6 in radius.
	circle := gomath.NewCircle[float32](64, 64, 6)
	box := gomath.AABB[float32]{Center: gomath.NewVector2[float32](512, 48), HalfWidth: 512, HalfHeight: 48}

	penetration, normal, overlapping := CircleAABBWithPenetration2(circle, box)

	assert.True(t, overlapping, "a circle inside the box overlaps it")
	assert.Greater(t, penetration, float32(0))
	assert.InDelta(t, 1.0, normal.Length(), 0.001, "normal must be a unit vector, got %v", normal)
	// The nearest wall is horizontal, so the normal must point along y.
	assert.InDelta(t, 0.0, normal.X, 0.001, "expected the nearest (horizontal) wall, got %v", normal)
}

// TestLineSegmentLineSegment2D_Integers asserts the integer instantiation, which
// IntegerIntersectionLineSegment relies on, reports intersections correctly.
// The ratios were computed in T, so integer division truncated them to zero.
func TestLineSegmentLineSegment2D_Integers(t *testing.T) {
	tests := map[string]struct {
		a, b       gomath.LineSegment[int]
		intersects bool
		at         gomath.Vector2[int]
	}{
		"crossing at 5,5": {
			a:          gomath.NewLineSegmentXY[int](0, 0, 10, 10),
			b:          gomath.NewLineSegmentXY[int](0, 10, 10, 0),
			intersects: true,
			at:         gomath.Vector2[int]{X: 5, Y: 5},
		},
		"crossing at 2,3": {
			a:          gomath.NewLineSegmentXY[int](0, 3, 10, 3),
			b:          gomath.NewLineSegmentXY[int](2, 0, 2, 10),
			intersects: true,
			at:         gomath.Vector2[int]{X: 2, Y: 3},
		},
		"apart": {
			a:          gomath.NewLineSegmentXY[int](0, 0, 1, 1),
			b:          gomath.NewLineSegmentXY[int](50, 50, 60, 60),
			intersects: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			point, intersects := LineSegmentLineSegment2D(test.a.P1, test.a.P2, test.b.P1, test.b.P2)
			assert.Equal(t, test.intersects, intersects)
			if !test.intersects {
				return
			}
			if assert.NotNil(t, point) {
				assert.Equal(t, test.at, *point)
			}

			// The public integer entry point must agree.
			assert.True(t, IntegerIntersectionLineSegment(test.a, test.b))
		})
	}
}

func newEllipsoid64(x, y, a, b float64) gomath.Ellipsoid[float64] {
	return gomath.Ellipsoid[float64]{
		Position: gomath.Vector2[float64]{X: x, Y: y},
		Radii:    gomath.Vector2[float64]{X: a, Y: b},
	}
}

func translateEllipsoid(e gomath.Ellipsoid[float64], offset gomath.Vector2[float64]) gomath.Ellipsoid[float64] {
	return gomath.Ellipsoid[float64]{
		Position: e.Position.Add(offset),
		Radii:    e.Radii,
	}
}
