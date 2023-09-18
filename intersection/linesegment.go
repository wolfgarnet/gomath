package intersection

import (
	"github.com/wolfgarnet/gomath"
	"math"
)

const maxf = 1e300

// LineSegmentLineSegment2D tests if two line segments intersects. Supports collinear segments.
// Based on https://github.com/pgkelley4/line-segments-intersect/blob/master/js/line-segments-intersect.js
// Returns true if the segments overlap, but can still return a point where the extended versions will eventually overlap.
func LineSegmentLineSegment2D[T gomath.Number](p1, p2, q1, q2 gomath.Vector2[T]) (*gomath.Vector2[T], bool) {
	s1 := p2.Sub(p1)
	s2 := q2.Sub(q1)

	qp1 := q1.Sub(p1)

	uNumerator := qp1.X*s1.Y - qp1.Y*s1.X
	denominator := s1.X*s2.Y - s1.Y*s2.X

	// Lines are collinear.
	if uNumerator == 0 && denominator == 0 {
		// Do they touch? (Are any of the points equal?)
		if (p1.X == q1.X && p1.Y == q1.Y) ||
			(p2.X == q2.X && p2.Y == q2.Y) ||
			(p1.X == q2.X && p1.Y == q2.Y) ||
			(p2.X == q1.X && p2.Y == q1.Y) {
			return nil, true
		}

		// Do they overlap? (Are all the circle differences in either direction the same sign)
		return nil, !allEqual((q1.X-p1.Y < 0), (q1.X-p2.X < 0), (q2.X-p1.X < 0), (q2.X-p2.X < 0)) ||
			!allEqual((q1.Y-p1.Y < 0), (q1.Y-p2.Y < 0), (q2.Y-p1.Y < 0), (q2.Y-p2.Y < 0))
	}

	// lines are parallel
	if denominator == 0 {
		return nil, false
	}

	u := uNumerator / denominator
	t := (qp1.X*s2.Y - qp1.Y*s2.X) / denominator

	return &gomath.Vector2[T]{X: p1.X + (t * s1.X), Y: p1.Y + (t * s1.Y)},
		(t >= 0) && (t <= 1) && (u >= 0) && (u <= 1)
}

// LineSegmentLineSegment2D2 finds the intersection point between two line segments.
// This algorithm does not support collinear segments.
// https://stackoverflow.com/questions/563198/how-do-you-detect-where-two-line-segments-intersect
func LineSegmentLineSegment2D2[T gomath.Number](l1P1, l1P2, l2P1, l2P2 gomath.Vector2[T]) (*gomath.Vector2[T], bool) {
	s1x := l1P2.X - l1P1.X
	s1y := l1P2.Y - l1P1.Y
	s2x := l2P2.X - l2P1.X
	s2y := l2P2.Y - l2P1.Y

	s := (-s1y*(l1P1.X-l2P1.X) + s1x*(l1P1.Y-l2P1.Y)) / (-s2x*s1y + s1x*s2y)
	t := (s2x*(l1P1.Y-l2P1.Y) - s2y*(l1P1.X-l2P1.X)) / (-s2x*s1y + s1x*s2y)

	if s >= 0 && s <= 1 && t >= 0 && t <= 1 {
		// Collision detected
		return &gomath.Vector2[T]{X: l1P1.X + (t * s1x), Y: l1P1.Y + (t * s1y)}, true
	}

	return nil, false
}

func LineSquaredAABB[T gomath.Float](ls gomath.LineSegment[T], bb gomath.SquaredAABB2D[T]) bool {
	return LineAABB(ls, gomath.AABB[T]{
		Center:     bb.Center,
		HalfWidth:  bb.HalfSize,
		HalfHeight: bb.HalfSize,
	})
}

func LineAABB[T gomath.Float](ls gomath.LineSegment[T], bb gomath.AABB[T]) bool {
	beginToEnd := ls.P2.Sub(ls.P1)
	minToMax := gomath.Vector2[T]{X: bb.HalfWidth, Y: bb.HalfHeight}
	bbCenter := gomath.Vector2[T]{X: bb.Center.X, Y: bb.Center.Y}
	min := bbCenter.Sub(minToMax)
	max := bbCenter.Add(minToMax)
	beginToMin := min.Sub(ls.P1)
	beginToMax := max.Sub(ls.P1)
	tNear := T(-math.MaxFloat32)
	tFar := T(math.MaxFloat32)
	axes := []int{0, 1}

	for _, axis := range axes {
		if beginToEnd.Component(axis) == 0 { // parallel
			if beginToMin.Component(axis) > 0 || beginToMax.Component(axis) < 0 {
				return false
			}
		} else {
			var t1 = beginToMin.Component(axis) / beginToEnd.Component(axis)
			var t2 = beginToMax.Component(axis) / beginToEnd.Component(axis)
			tMin := t1
			tMax := t2
			if t2 < tMin {
				tMin = t2
			}
			if t2 > tMax {
				tMax = t2
			}
			if tMin > tNear {
				tNear = tMin
			}
			if tMax < tFar {
				tFar = tMax
			}
			if tNear > tFar || tFar < 0 {
				return false
			}
		}
	}

	if tNear >= 0 && tNear <= 1 {
		return true
	}
	if tFar >= 0 && tFar <= 1 {
		return true
	}

	return false
}

func LineInsideSquaredAABB[T gomath.Number](ls gomath.LineSegment[T], bb gomath.SquaredAABB2D[T]) bool {
	if ls.P1.X > bb.Center.X-bb.HalfSize &&
		ls.P2.X > bb.Center.X-bb.HalfSize &&
		ls.P1.X < bb.Center.X+bb.HalfSize &&
		ls.P2.X < bb.Center.X+bb.HalfSize &&
		ls.P1.Y > bb.Center.Y-bb.HalfSize &&
		ls.P2.Y > bb.Center.Y-bb.HalfSize &&
		ls.P1.Y < bb.Center.Y+bb.HalfSize &&
		ls.P2.Y < bb.Center.Y+bb.HalfSize {
		return true
	}

	return false
}

func LineInsideAABB[T gomath.Number](ls gomath.LineSegment[T], bb gomath.AABB[T]) bool {
	if ls.P1.X > bb.Center.X-bb.HalfWidth &&
		ls.P2.X > bb.Center.X-bb.HalfWidth &&
		ls.P1.X < bb.Center.X+bb.HalfWidth &&
		ls.P2.X < bb.Center.X+bb.HalfWidth &&
		ls.P1.Y > bb.Center.Y-bb.HalfHeight &&
		ls.P2.Y > bb.Center.Y-bb.HalfHeight &&
		ls.P1.Y < bb.Center.Y+bb.HalfHeight &&
		ls.P2.Y < bb.Center.Y+bb.HalfHeight {
		return true
	}

	return false
}

func LineSegmentInsideOrIntersectsSquaredAABB[T gomath.Float](ls gomath.LineSegment[T], bb gomath.SquaredAABB2D[T]) bool {
	if ls.P1.X < bb.Center.X-bb.HalfSize && ls.P2.X < bb.Center.X-bb.HalfSize {
		return false
	}
	if ls.P1.X > bb.Center.X+bb.HalfSize && ls.P2.X > bb.Center.X+bb.HalfSize {
		return false
	}
	if ls.P1.Y < bb.Center.Y-bb.HalfSize && ls.P2.Y < bb.Center.Y-bb.HalfSize {
		return false
	}
	if ls.P1.Y > bb.Center.Y+bb.HalfSize && ls.P2.Y > bb.Center.Y+bb.HalfSize {
		return false
	}

	//if ls.Intersects(bb) {
	if LineSquaredAABB(ls, bb) {
		return true
	}

	// Check if inside bbox
	//return ls.Inside(bb)
	return LineInsideSquaredAABB(ls, bb)
}

func LineSegmentInsideOrIntersectsAABB[T gomath.Float](ls gomath.LineSegment[T], bb gomath.AABB[T]) bool {
	if ls.P1.X < bb.Center.X-bb.HalfWidth && ls.P2.X < bb.Center.X-bb.HalfWidth {
		return false
	}
	if ls.P1.X > bb.Center.X+bb.HalfWidth && ls.P2.X > bb.Center.X+bb.HalfWidth {
		return false
	}
	if ls.P1.Y < bb.Center.Y-bb.HalfHeight && ls.P2.Y < bb.Center.Y-bb.HalfHeight {
		return false
	}
	if ls.P1.Y > bb.Center.Y+bb.HalfHeight && ls.P2.Y > bb.Center.Y+bb.HalfHeight {
		return false
	}

	//if ls.Intersects(bb) {
	if LineAABB(ls, bb) {
		return true
	}

	// Check if inside bbox
	//return ls.Inside(bb)
	return LineInsideAABB(ls, bb)
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
