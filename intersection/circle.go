package intersection

import (
	"github.com/wolfgarnet/gomath"
)

// CircleCircle determines if the two circles are intersecting
func CircleCircle[T gomath.Number](a, b gomath.Circle[T]) bool {
	dist := T(b.Center.Sub(a.Center).Length())
	return dist < a.Radius+b.Radius
}

// CircleCircleWithPenetration determines if the two circles are intersecting and how deep the penetration is
func CircleCircleWithPenetration[T gomath.Number](a, b gomath.Circle[T]) (T, gomath.Vector2[T], bool) {
	dist := b.Center.Sub(a.Center).Length()
	penetration := (a.Radius + b.Radius) - dist
	return penetration, b.Center.Sub(a.Center).Normalize(), penetration > 0
}

func CircleSegment[T gomath.Number](c gomath.Circle[T], p1, p2 gomath.Vector2[T]) bool {
	dist := gomath.MinimumDistance(p1, p2, c.Center)

	return dist < c.Radius
}

func CircleLineSegment[T gomath.Number](c gomath.Circle[T], s gomath.LineSegment[T]) bool {
	dist := gomath.MinimumDistance(s.P1, s.P2, c.Center)

	return dist < c.Radius
}

func CircleLineSegmentWithPenetration[T gomath.Number](c gomath.Circle[T], s gomath.LineSegment[T]) (T, bool) {
	dist := gomath.MinimumDistance(s.P1, s.P2, c.Center)

	return c.Radius - dist, dist < c.Radius
}

func CirclePoint[T gomath.Number](c gomath.Circle[T], v gomath.Vector2[T]) bool {
	dist := c.Center.DistanceTo(v)

	return dist < c.Radius
}

func CirclePointWithPenetration[T gomath.Number](c gomath.Circle[T], v gomath.Vector2[T]) (T, gomath.Vector2[T], bool) {
	dist := c.Center.DistanceTo(v)

	return c.Radius - dist, v.Sub(c.Center).Normalize(), dist < c.Radius
}

func CircleAABB[T gomath.Number](c gomath.Circle[T], b gomath.SquaredAABB2D[T]) bool {
	p1 := gomath.Vector2[T]{X: b.Center.X + b.HalfSize, Y: b.Center.Y + b.HalfSize}
	p2 := gomath.Vector2[T]{X: b.Center.X + b.HalfSize, Y: b.Center.Y - b.HalfSize}

	if CircleSegment(c, p1, p2) {
		return true
	}

	p3 := gomath.Vector2[T]{X: b.Center.X - b.HalfSize, Y: b.Center.Y - b.HalfSize}

	if CircleSegment(c, p2, p3) {
		return true
	}

	p4 := gomath.Vector2[T]{X: b.Center.X - b.HalfSize, Y: b.Center.Y + b.HalfSize}

	if CircleSegment(c, p3, p4) {
		return true
	}
	if CircleSegment(c, p4, p1) {
		return true
	}

	if AABB2DPoint(b, c.Center) {
		return true
	}

	return false
}

func CircleAABBWithPenetration[T gomath.Number](c gomath.Circle[T], b gomath.SquaredAABB2D[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	if !CircleAABB(c, b) {
		return 0, normal, false
	}

	var max T
	for _, side := range b.Sides() {
		p, ok := CircleLineSegmentWithPenetration(c, side)
		if !ok {
			continue
		}
		if p > max {
			max = p
			normal = side.Normal()
		}
	}

	return max, normal, true
}
