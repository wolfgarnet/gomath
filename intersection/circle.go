package intersection

import (
	"github.com/wolfgarnet/gomath"
	"github.com/wolfgarnet/gomath/collider"
)

// CircleCircle determines if the two circles are intersecting
func CircleCircle[T gomath.Number](a, b gomath.Circle[T]) bool {
	dist := b.Center.Sub(a.Center).Length()
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

func CircleSegmentCollider[t gomath.Number](circle gomath.CircleCollider[t], offset gomath.Vector2[t], p1, p2 gomath.Vector2[t]) bool {
	dist := gomath.MinimumDistance(p1, p2, offset)
	return dist < circle.Radius
}

func CircleLineSegmentCollider[T gomath.Number](circle gomath.CircleCollider[T], offset gomath.Vector2[T], segment gomath.LineSegment[T]) bool {
	dist := gomath.MinimumDistance(segment.P1, segment.P2, offset)

	return dist < circle.Radius
}

func CircleLineSegmentColliderPenetration[T gomath.Number](circle gomath.CircleCollider[T], offset gomath.Vector2[T], segment gomath.LineSegment[T]) (T, bool) {
	dist := gomath.MinimumDistance(segment.P1, segment.P2, offset)
	return circle.Radius - dist, dist < circle.Radius
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

func CircleSquaredAABB[T gomath.Number](c gomath.Circle[T], b gomath.SquaredAABB2D[T]) bool {
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

	if SquaredAABBPoint(b, c.Center) {
		return true
	}

	return false
}

func CircleAABBColliders[T gomath.Number](circle gomath.CircleCollider[T], circleOffset gomath.Vector2[T], aabb gomath.AABBCollider[T], aabbOffset gomath.Vector2[T]) bool {
	return collider.CircleAABB(circle.Radius, circleOffset.X, circleOffset.Y, aabb.HalfWidth, aabb.HalfHeight, aabbOffset.X, aabbOffset.Y)
}

func CircleAABB[T gomath.Number](c gomath.Circle[T], b gomath.AABB[T]) bool {
	p1 := gomath.Vector2[T]{X: b.Center.X + b.HalfWidth, Y: b.Center.Y + b.HalfHeight}
	p2 := gomath.Vector2[T]{X: b.Center.X + b.HalfWidth, Y: b.Center.Y - b.HalfHeight}

	if CircleSegment(c, p1, p2) {
		return true
	}

	p3 := gomath.Vector2[T]{X: b.Center.X - b.HalfWidth, Y: b.Center.Y - b.HalfHeight}

	if CircleSegment(c, p2, p3) {
		return true
	}

	p4 := gomath.Vector2[T]{X: b.Center.X - b.HalfWidth, Y: b.Center.Y + b.HalfHeight}

	if CircleSegment(c, p3, p4) {
		return true
	}
	if CircleSegment(c, p4, p1) {
		return true
	}

	if AABBPoint(b, c.Center) {
		return true
	}

	return false
}

func CircleSquaredAABBWithPenetration[T gomath.Number](c gomath.Circle[T], b gomath.SquaredAABB2D[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	if !CircleSquaredAABB(c, b) {
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

func CircleAABBWithPenetration[T gomath.Number](c gomath.Circle[T], b gomath.AABB[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
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

func CircleInsideAABBColliders[T gomath.Number](circle gomath.CircleCollider[T], circleOffset gomath.Vector2[T], aabb gomath.AABBCollider[T], aabbOffset gomath.Vector2[T]) bool {
	return circleOffset.X-circle.Radius >= aabbOffset.X-aabb.HalfWidth &&
		circleOffset.X+circle.Radius <= aabbOffset.X+aabb.HalfWidth &&
		circleOffset.Y-circle.Radius >= aabbOffset.Y-aabb.HalfHeight &&
		circleOffset.Y+circle.Radius <= aabbOffset.Y+aabb.HalfHeight
}

func isCircleInsideAABB[T gomath.Number](c gomath.Circle[T], b gomath.AABB[T]) bool {
	return c.Center.X-c.Radius >= b.Center.X-b.HalfWidth &&
		c.Center.X+c.Radius <= b.Center.X+b.HalfWidth &&
		c.Center.Y-c.Radius >= b.Center.Y-b.HalfHeight &&
		c.Center.Y+c.Radius <= b.Center.Y+b.HalfHeight
}

func AABBInsideCircleColliders[T gomath.Number](circle gomath.CircleCollider[T], circleOffset gomath.Vector2[T], aabb gomath.AABBCollider[T], aabbOffset gomath.Vector2[T]) bool {
	return circleOffset.X-circle.Radius >= aabbOffset.X-aabb.HalfWidth &&
		circleOffset.X+circle.Radius <= aabbOffset.X+aabb.HalfWidth &&
		circleOffset.Y-circle.Radius >= aabbOffset.Y-aabb.HalfHeight &&
		circleOffset.Y+circle.Radius <= aabbOffset.Y+aabb.HalfHeight
}

func isAABBInsideCircle[T gomath.Number](c gomath.Circle[T], b gomath.AABB[T]) bool {
	corners := []gomath.Vector2[T]{
		{X: b.Center.X - b.HalfWidth, Y: b.Center.Y - b.HalfHeight},
		{X: b.Center.X + b.HalfWidth, Y: b.Center.Y - b.HalfHeight},
		{X: b.Center.X + b.HalfWidth, Y: b.Center.Y + b.HalfHeight},
		{X: b.Center.X - b.HalfWidth, Y: b.Center.Y + b.HalfHeight},
	}

	for _, corner := range corners {
		if c.Center.DistanceTo(corner) > c.Radius {
			return false
		}
	}
	return true
}

func CircleAABBCollidersPenetration[T gomath.Number](circle gomath.CircleCollider[T], circleOffset gomath.Vector2[T], aabb gomath.AABBCollider[T], aabbOffset gomath.Vector2[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	return collider.CircleAABBPenetration(circle.Radius, circleOffset.X, circleOffset.Y, aabb.HalfWidth, aabb.HalfHeight, aabbOffset.X, aabbOffset.Y)
}

func CircleAABBWithPenetration2[T gomath.Number](c gomath.Circle[T], b gomath.AABB[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	if !CircleAABB(c, b) {
		return 0, normal, false
	}

	if isCircleInsideAABB(c, b) {
		minDist := c.Radius
		sides := b.Sides()
		for _, side := range sides {
			dist := gomath.MinimumDistance(side.P1, side.P2, c.Center)
			if dist < minDist {
				minDist = dist
				normal = side.Normal()
			}
		}
		return minDist, normal, true
	}

	if isAABBInsideCircle(c, b) {
		maxCornerDist := T(0)
		corners := []gomath.Vector2[T]{
			{X: b.Center.X - b.HalfWidth, Y: b.Center.Y - b.HalfHeight},
			{X: b.Center.X + b.HalfWidth, Y: b.Center.Y - b.HalfHeight},
			{X: b.Center.X + b.HalfWidth, Y: b.Center.Y + b.HalfHeight},
			{X: b.Center.X - b.HalfWidth, Y: b.Center.Y + b.HalfHeight},
		}
		for _, corner := range corners {
			dist := c.Radius - c.Center.DistanceTo(corner)
			if dist > maxCornerDist {
				maxCornerDist = dist
				normal = corner.Sub(c.Center).Normalize()
			}
		}
		return maxCornerDist, normal, true
	}

	var max T
	sides := b.Sides()
	for i, side := range sides {
		j := (i + 2) % 4
		if !PointSideOfBox(c.Center, side, sides[j]) {
			continue
		}
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
