package collider

import "github.com/wolfgarnet/gomath"

// LineSegmentsIntersect tests whether two line segments intersect, and returns
// the intersection point (of the extended lines) together with whether the
// segments are collinear. It supports collinear overlap. Ported from the former
// scalar collider package (the pgkelley algorithm); structs are unpacked to the
// original local scalars so the body is behaviour-identical.
//
// Note this differs from LineSegmentLineSegment2D, which returns a pointer to the
// intersection point and no collinear flag.
func LineSegmentsIntersect[T gomath.Number](segment1, segment2 gomath.LineSegment[T]) (intersectionPoint gomath.Vector2[T], collinear bool, intersecting bool) {
	segment1StartX, segment1StartY, segment1EndX, segment1EndY := segment1.P1.X, segment1.P1.Y, segment1.P2.X, segment1.P2.Y
	segment2StartX, segment2StartY, segment2EndX, segment2EndY := segment2.P1.X, segment2.P1.Y, segment2.P2.X, segment2.P2.Y

	diff1X, diff1Y := gomath.VectorSubtract(segment1EndX, segment1EndY, segment1StartX, segment1StartY)
	diff2X, diff2Y := gomath.VectorSubtract(segment2EndX, segment2EndY, segment2StartX, segment2StartY)

	diff3X, diff3Y := gomath.VectorSubtract(segment2StartX, segment2StartY, segment1StartX, segment1StartY)

	uNumerator := diff3X*diff1Y - diff3Y*diff1X
	denominator := diff1X*diff2Y - diff1Y*diff2X

	// Lines are collinear.
	if uNumerator == 0 && denominator == 0 {
		if (segment1StartX == segment2StartX && segment1StartY == segment2StartY) ||
			(segment1EndX == segment2EndX && segment1EndY == segment2EndY) ||
			(segment1StartX == segment2EndX && segment1StartY == segment2EndY) ||
			(segment1EndX == segment2StartX && segment1EndY == segment2StartY) {
			return intersectionPoint, true, true
		}

		return intersectionPoint, true, !allEqual(segment2StartX-segment1StartX < 0, segment2StartX-segment1EndX < 0, segment2EndX-segment1StartX < 0, segment2EndX-segment1EndX < 0) ||
			!allEqual(segment2StartY-segment1StartY < 0, segment2StartY-segment1EndY < 0, segment2EndY-segment1StartY < 0, segment2EndY-segment1EndY < 0)
	}

	// Lines are parallel.
	if denominator == 0 {
		return intersectionPoint, false, false
	}

	u := uNumerator / denominator
	t := (diff3X*diff2Y - diff3Y*diff2X) / denominator

	return gomath.Vector2[T]{
		X: segment1StartX + (t * diff1X),
		Y: segment1StartY + (t * diff1Y),
	}, false, (t >= 0) && (t <= 1) && (u >= 0) && (u <= 1)
}
