package collider

import "github.com/wolfgarnet/gomath"

func CircleCircle[T gomath.Number](radius1, offset1x, offset1y, radius2, offset2x, offset2y T) bool {
	dx := offset2x - offset1x
	dy := offset2y - offset1y
	distSq := dx*dx + dy*dy
	radiiSum := radius1 + radius2
	return distSq < radiiSum*radiiSum
}

func CircleCirclePenetration[T gomath.Number](radius1, offset1x, offset1y, radius2, offset2x, offset2y T) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	dx := offset2x - offset1x
	dy := offset2y - offset1y
	distSq := dx*dx + dy*dy
	radiiSum := radius1 + radius2
	if distSq >= radiiSum*radiiSum {
		return 0, normal, false
	}

	dist := gomath.Sqrt(distSq)
	penetration = radiiSum - dist

	if dist == 0 {
		return penetration, normal, true
	}

	invDist := T(1) / dist
	return penetration, gomath.Vector2[T]{
		X: dx * invDist,
		Y: dy * invDist,
	}, true
}

func CircleAABB[T gomath.Number](circleRadius T, circleOffsetX, circleOffsetY T, aabbHalfWidth, aabbHalfHeight T, aabbOffsetX, aabbOffsetY T) bool {
	// Find the closest point on the AABB to the circle center
	closestX := clamp(circleOffsetX, aabbOffsetX-aabbHalfWidth, aabbOffsetX+aabbHalfWidth)
	closestY := clamp(circleOffsetY, aabbOffsetY-aabbHalfHeight, aabbOffsetY+aabbHalfHeight)

	dx := circleOffsetX - closestX
	dy := circleOffsetY - closestY

	distSq := dx*dx + dy*dy
	return distSq < circleRadius*circleRadius
}

func CircleAABBPenetration[T gomath.Number](circleRadius T, circleOffsetX, circleOffsetY T, aabbHalfWidth, aabbHalfHeight T, aabbOffsetX, aabbOffsetY T) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	// Find the closest point on the AABB to the circle center
	closestX := clamp(circleOffsetX, aabbOffsetX-aabbHalfWidth, aabbOffsetX+aabbHalfWidth)
	closestY := clamp(circleOffsetY, aabbOffsetY-aabbHalfHeight, aabbOffsetY+aabbHalfHeight)

	dx := circleOffsetX - closestX
	dy := circleOffsetY - closestY

	distSq := dx*dx + dy*dy
	dist := gomath.Sqrt(distSq)

	// If the circle center is inside the AABB
	if dist == 0 {
		// Find the distance to the closest edge
		px := aabbHalfWidth - gomath.Abs(circleOffsetX-aabbOffsetX)
		py := aabbHalfHeight - gomath.Abs(circleOffsetY-aabbOffsetY)

		if px < py {
			penetration = px + circleRadius
			if circleOffsetX > aabbOffsetX {
				return penetration, gomath.Vector2[T]{
					X: -1,
					Y: 0,
				}, true
			}
			return penetration, gomath.Vector2[T]{
				X: 1,
				Y: 0,
			}, true
		}

		penetration = py + circleRadius
		if circleOffsetY > aabbOffsetY {
			return penetration, gomath.Vector2[T]{
				X: 0,
				Y: -1,
			}, true
		}
		return penetration, gomath.Vector2[T]{
			X: 0,
			Y: 1,
		}, true
	}

	penetration = circleRadius - dist
	if penetration <= 0 {
		return 0, normal, false
	}

	invDist := T(1) / dist
	return penetration, gomath.Vector2[T]{
		X: -dx * invDist,
		Y: -dy * invDist,
	}, true
}

func CircleLineSegmentPenetration[T gomath.Number](circleRadius T, circleOffsetX, circleOffsetY T, startSegmentX T, startSegmentY T, endSegmentX T, endSegmentY T) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	// VectorToMinimumDistance returns (circle - closestPointOnSegment), i.e. a
	// vector pointing from the segment toward the circle. Negate it so the normal
	// points from the circle toward the segment (the obstacle), matching the
	// convention of CircleCircle/CircleAABB.
	tox, toy, dist := gomath.VectorToMinimumDistance(startSegmentX, startSegmentY, endSegmentX, endSegmentY, circleOffsetX, circleOffsetY)
	normalx, normaly := gomath.NormalizeVector(-tox, -toy)
	return circleRadius - dist, gomath.Vector2[T]{
		X: normalx,
		Y: normaly,
	}, dist < circleRadius
}

func CircleSegment[T gomath.Number](circleRadius T, circleOffsetX, circleOffsetY T, startSegmentX T, startSegmentY T, endSegmentX T, endSegmentY T) bool {
	dist := gomath.MinimumDistanceValues(startSegmentX, startSegmentY, endSegmentX, endSegmentY, circleOffsetX, circleOffsetY)

	return dist < circleRadius
}

func clamp[T gomath.Number](val, min, max T) T {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
