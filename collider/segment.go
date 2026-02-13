package collider

import "github.com/wolfgarnet/gomath"

// LineSegmentLineSegment2D tests if two line segments intersects. Supports collinear segments.
// Based on https://github.com/pgkelley4/line-segments-intersect/blob/master/js/line-segments-intersect.js
// Returns true if the segments overlap, but can still return a point where the extended versions will eventually overlap.
func LineSegmentLineSegment2D[T gomath.Number](segment1StartX T, segment1StartY T, segment1EndX T, segment1EndY T, segment2StartX T, segment2StartY T, segment2EndX T, segment2EndY T) (intersectionPoint gomath.Vector2[T], collinear bool, intersecting bool) {
	diff1X, diff1Y := gomath.VectorSubtract(segment1EndX, segment1EndY, segment1StartX, segment1StartY)
	diff2X, diff2Y := gomath.VectorSubtract(segment2EndX, segment2EndY, segment2StartX, segment2StartY)

	diff3X, diff3Y := gomath.VectorSubtract(segment2StartX, segment2StartY, segment1StartX, segment1StartY)

	uNumerator := diff3X*diff1Y - diff3Y*diff1X
	denominator := diff1X*diff2Y - diff1Y*diff2X

	// Lines are collinear.
	if uNumerator == 0 && denominator == 0 {
		// Do they touch? (Are any of the points equal?)
		if (segment1StartX == segment2StartX && segment1StartY == segment2StartY) ||
			(segment1EndX == segment2EndX && segment1EndY == segment2EndY) ||
			(segment1StartX == segment2EndX && segment1StartY == segment2EndY) ||
			(segment1EndX == segment2StartX && segment1EndY == segment2StartY) {
			return intersectionPoint, true, true
		}

		// Do they overlap? (Are all the circle differences in either direction the same sign)
		return intersectionPoint, true, !allEqual(segment2StartX-segment1StartX < 0, segment2StartX-segment1EndX < 0, segment2EndX-segment1StartX < 0, segment2EndX-segment1EndX < 0) ||
			!allEqual(segment2StartY-segment1StartY < 0, segment2StartY-segment1EndY < 0, segment2EndY-segment1StartY < 0, segment2EndY-segment1EndY < 0)
	}

	// lines are parallel
	if denominator == 0 {
		return intersectionPoint, false, false
	}

	u := uNumerator / denominator
	t := (diff3X*diff2Y - diff3Y*diff2X) / denominator

	return gomath.Vector2[T]{
		X: segment1StartX + (t * diff1X), // x position where they cross (might be in infinity)
		Y: segment1StartY + (t * diff1Y), // y position where they cross (might be in infinity)
	}, false, (t >= 0) && (t <= 1) && (u >= 0) && (u <= 1) // if they cross
}

func allEqual(bs ...bool) bool {
	expected := bs[0]
	for i := 1; i < len(bs); i++ {
		if bs[i] != expected {
			return false
		}
	}

	return true
}
