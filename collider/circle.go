package collider

import "github.com/wolfgarnet/gomath"

func CircleCircle[T gomath.Number](radius1, radius2 T, offset1x, offset1y, offset2x, offset2y T) bool {
	dx := offset2x - offset1x
	dy := offset2y - offset1y
	distSq := dx*dx + dy*dy
	radiiSum := radius1 + radius2
	return distSq < radiiSum*radiiSum
}

func CircleCirclePenetration[T gomath.Number](radius1, radius2 T, offset1x, offset1y, offset2x, offset2y T) (penetration T, normalX T, normalY T, overlapping bool) {
	dx := offset2x - offset1x
	dy := offset2y - offset1y
	distSq := dx*dx + dy*dy
	radiiSum := radius1 + radius2
	if distSq >= radiiSum*radiiSum {
		return 0, 0, 0, false
	}

	dist := gomath.Sqrt(distSq)
	penetration = radiiSum - dist

	if dist == 0 {
		return penetration, 0, 0, true
	}

	invDist := T(1) / dist
	return penetration, dx * invDist, dy * invDist, true
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

func CircleAABBPenetration[T gomath.Number](circleRadius T, circleOffsetX, circleOffsetY T, aabbHalfWidth, aabbHalfHeight T, aabbOffsetX, aabbOffsetY T) (penetration T, normalX T, normalY T, overlapping bool) {
	// Find the closest point on the AABB to the circle center
	closestX := clamp(circleOffsetX, aabbOffsetX-aabbHalfWidth, aabbOffsetX+aabbHalfWidth)
	closestY := clamp(circleOffsetY, aabbOffsetY-aabbHalfHeight, aabbOffsetY+aabbHalfHeight)

	dx := circleOffsetX - closestX
	dy := circleOffsetY - closestY

	distSq := dx*dx + dy*dy
	if distSq >= circleRadius*circleRadius {
		// Even if not intersecting via closest point, if the circle center is INSIDE the AABB,
		// we still might have an intersection that CircleAABB would catch if we used a more robust method.
		// However, for AABB, the closest point method IS robust, except when the circle center is INSIDE.
		// If circle center is inside, closestX == circleOffsetX and closestY == circleOffsetY, distSq == 0.
	}

	dist := gomath.Sqrt(distSq)

	// If the circle center is inside the AABB
	if dist == 0 {
		// Find the distance to the closest edge
		px := aabbHalfWidth - gomath.Abs(circleOffsetX-aabbOffsetX)
		py := aabbHalfHeight - gomath.Abs(circleOffsetY-aabbOffsetY)

		if px < py {
			penetration = px + circleRadius
			if circleOffsetX > aabbOffsetX {
				return penetration, 1, 0, true
			}
			return penetration, -1, 0, true
		}

		penetration = py + circleRadius
		if circleOffsetY > aabbOffsetY {
			return penetration, 0, 1, true
		}
		return penetration, 0, -1, true
	}

	penetration = circleRadius - dist
	if penetration <= 0 {
		return 0, 0, 0, false
	}

	invDist := T(1) / dist
	return penetration, dx * invDist, dy * invDist, true
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
