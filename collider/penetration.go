package collider

import "github.com/wolfgarnet/gomath"

// This file holds the struct-based collision-penetration (MTV) functions,
// ported byte-for-byte from the former scalar `collider` package so behaviour is
// identical. Each returns (penetration, normal, overlapping); the normal points
// from the first shape toward the second (the obstacle), matching
// TestPenetration_NormalConventionShape1ToShape2. Structs are unpacked to local
// scalars at the top of each function so the body is the exact former collider
// implementation.

// CircleCirclePenetration returns the penetration depth and separation normal for
// two overlapping circles.
func CircleCirclePenetration[T gomath.Number](a, b gomath.Circle[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	radius1, offset1x, offset1y := a.Radius, a.Center.X, a.Center.Y
	radius2, offset2x, offset2y := b.Radius, b.Center.X, b.Center.Y

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

// CircleAABBPenetration returns the penetration depth and separation normal for a
// circle overlapping an axis-aligned box.
func CircleAABBPenetration[T gomath.Number](c gomath.Circle[T], box gomath.AABB[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	circleRadius, circleOffsetX, circleOffsetY := c.Radius, c.Center.X, c.Center.Y
	aabbHalfWidth, aabbHalfHeight, aabbOffsetX, aabbOffsetY := box.HalfWidth, box.HalfHeight, box.Center.X, box.Center.Y

	closestX := gomath.Clamp(circleOffsetX, aabbOffsetX-aabbHalfWidth, aabbOffsetX+aabbHalfWidth)
	closestY := gomath.Clamp(circleOffsetY, aabbOffsetY-aabbHalfHeight, aabbOffsetY+aabbHalfHeight)

	dx := circleOffsetX - closestX
	dy := circleOffsetY - closestY

	distSq := dx*dx + dy*dy
	dist := gomath.Sqrt(distSq)

	// Circle centre inside the box.
	if dist == 0 {
		px := aabbHalfWidth - gomath.Abs(circleOffsetX-aabbOffsetX)
		py := aabbHalfHeight - gomath.Abs(circleOffsetY-aabbOffsetY)

		if px < py {
			penetration = px + circleRadius
			if circleOffsetX > aabbOffsetX {
				return penetration, gomath.Vector2[T]{X: -1, Y: 0}, true
			}
			return penetration, gomath.Vector2[T]{X: 1, Y: 0}, true
		}

		penetration = py + circleRadius
		if circleOffsetY > aabbOffsetY {
			return penetration, gomath.Vector2[T]{X: 0, Y: -1}, true
		}
		return penetration, gomath.Vector2[T]{X: 0, Y: 1}, true
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

// AABBAABBPenetration returns the penetration depth and separation normal for two
// overlapping axis-aligned boxes.
func AABBAABBPenetration[T gomath.Number](a, b gomath.AABB[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	aabbPosition1x, aabbPosition1y, aabb1HalfWidth, aabb1HalfHeight := a.Center.X, a.Center.Y, a.HalfWidth, a.HalfHeight
	aabbPosition2x, aabbPosition2y, aabb2HalfWidth, aabb2HalfHeight := b.Center.X, b.Center.Y, b.HalfWidth, b.HalfHeight

	diffx := aabbPosition1x - aabbPosition2x
	absDiffx := gomath.Abs(diffx)
	sizex := aabb1HalfWidth + aabb2HalfWidth
	if absDiffx > sizex {
		return 0, normal, false
	}

	diffy := aabbPosition1y - aabbPosition2y
	absDiffy := gomath.Abs(diffy)
	sizey := aabb1HalfHeight + aabb2HalfHeight
	if absDiffy > sizey {
		return 0, normal, false
	}

	px := sizex - absDiffx
	py := sizey - absDiffy

	// Normal points from box a toward box b (the obstacle).
	if px < py {
		penetration = px
		if diffx >= 0 {
			normal = gomath.Vector2[T]{X: -1, Y: 0}
		} else {
			normal = gomath.Vector2[T]{X: 1, Y: 0}
		}
	} else {
		penetration = py
		if diffy >= 0 {
			normal = gomath.Vector2[T]{X: 0, Y: -1}
		} else {
			normal = gomath.Vector2[T]{X: 0, Y: 1}
		}
	}

	return penetration, normal, true
}

// CircleLineSegmentPenetration returns the penetration depth and separation normal
// for a circle overlapping a line segment. The normal points from the circle
// toward the segment.
func CircleLineSegmentPenetration[T gomath.Number](c gomath.Circle[T], s gomath.LineSegment[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	circleRadius, circleOffsetX, circleOffsetY := c.Radius, c.Center.X, c.Center.Y
	startSegmentX, startSegmentY, endSegmentX, endSegmentY := s.P1.X, s.P1.Y, s.P2.X, s.P2.Y

	tox, toy, dist := gomath.VectorToMinimumDistance(startSegmentX, startSegmentY, endSegmentX, endSegmentY, circleOffsetX, circleOffsetY)
	normalx, normaly := gomath.NormalizeVector(-tox, -toy)
	return circleRadius - dist, gomath.Vector2[T]{
		X: normalx,
		Y: normaly,
	}, dist < circleRadius
}

// AABBLineSegmentPenetration returns the penetration depth and separation normal
// for a box overlapping a line segment.
func AABBLineSegmentPenetration[T gomath.Number](box gomath.AABB[T], s gomath.LineSegment[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	aabbPositionX, aabbPositionY, aabbHalfWidth, aabbHalfHeight := box.Center.X, box.Center.Y, box.HalfWidth, box.HalfHeight
	startSegmentX, startSegmentY, endSegmentX, endSegmentY := s.P1.X, s.P1.Y, s.P2.X, s.P2.Y

	aabbMinX := aabbPositionX - aabbHalfWidth
	aabbMinY := aabbPositionY - aabbHalfHeight
	aabbMaxX := aabbPositionX + aabbHalfWidth
	aabbMaxY := aabbPositionY + aabbHalfHeight

	overlapping = AABBLineSegmentBounds(aabbMinX, aabbMinY, aabbMaxX, aabbMaxY, startSegmentX, startSegmentY, endSegmentX, endSegmentY)
	if !overlapping {
		return 0, normal, false
	}

	p1Inside := startSegmentX >= aabbMinX && startSegmentX <= aabbMaxX &&
		startSegmentY >= aabbMinY && startSegmentY <= aabbMaxY
	p2Inside := endSegmentX >= aabbMinX && endSegmentX <= aabbMaxX &&
		endSegmentY >= aabbMinY && endSegmentY <= aabbMaxY

	if p1Inside && p2Inside {
		midX := (startSegmentX + endSegmentX) / 2
		midY := (startSegmentY + endSegmentY) / 2

		dx := midX - aabbPositionX
		dy := midY - aabbPositionY

		px := aabbHalfWidth - gomath.Abs(dx)
		py := aabbHalfHeight - gomath.Abs(dy)

		if px < py {
			penetration = px
			if dx >= 0 {
				normal = gomath.Vector2[T]{X: 1, Y: 0}
			} else {
				normal = gomath.Vector2[T]{X: -1, Y: 0}
			}
		} else {
			penetration = py
			if dy >= 0 {
				normal = gomath.Vector2[T]{X: 0, Y: 1}
			} else {
				normal = gomath.Vector2[T]{X: 0, Y: -1}
			}
		}
		return penetration, normal, true
	}

	if p1Inside || p2Inside {
		var px, py T
		var dx, dy T
		if p1Inside {
			dx = startSegmentX - aabbPositionX
			dy = startSegmentY - aabbPositionY
		} else {
			dx = endSegmentX - aabbPositionX
			dy = endSegmentY - aabbPositionY
		}

		px = aabbHalfWidth - gomath.Abs(dx)
		py = aabbHalfHeight - gomath.Abs(dy)

		if px < py {
			penetration = px
			if dx >= 0 {
				normal = gomath.Vector2[T]{X: 1, Y: 0}
			} else {
				normal = gomath.Vector2[T]{X: -1, Y: 0}
			}
		} else {
			penetration = py
			if dy >= 0 {
				normal = gomath.Vector2[T]{X: 0, Y: 1}
			} else {
				normal = gomath.Vector2[T]{X: 0, Y: -1}
			}
		}
		return penetration, normal, true
	}

	dx, dy, _ := gomath.VectorToMinimumDistance(startSegmentX, startSegmentY, endSegmentX, endSegmentY, aabbPositionX, aabbPositionY)
	closestX := aabbPositionX - dx
	closestY := aabbPositionY - dy

	px := aabbHalfWidth - gomath.Abs(closestX-aabbPositionX)
	py := aabbHalfHeight - gomath.Abs(closestY-aabbPositionY)

	if px < py {
		penetration = px
		if closestX-aabbPositionX >= 0 {
			normal = gomath.Vector2[T]{X: 1, Y: 0}
		} else {
			normal = gomath.Vector2[T]{X: -1, Y: 0}
		}
	} else {
		penetration = py
		if closestY-aabbPositionY >= 0 {
			normal = gomath.Vector2[T]{X: 0, Y: 1}
		} else {
			normal = gomath.Vector2[T]{X: 0, Y: -1}
		}
	}

	return penetration, normal, true
}

// AABBLineSegment reports whether the segment intersects or lies inside the box.
func AABBLineSegment[T gomath.Number](box gomath.AABB[T], s gomath.LineSegment[T]) bool {
	return AABBLineSegmentBounds(
		box.Center.X-box.HalfWidth, box.Center.Y-box.HalfHeight,
		box.Center.X+box.HalfWidth, box.Center.Y+box.HalfHeight,
		s.P1.X, s.P1.Y, s.P2.X, s.P2.Y,
	)
}

// AABBLineSegmentBounds reports whether the segment intersects or lies inside the
// box given by its min/max corners.
func AABBLineSegmentBounds[T gomath.Number](aabbMinX, aabbMinY, aabbMaxX, aabbMaxY T, startSegmentX, startSegmentY, endSegmentX, endSegmentY T) bool {
	p1Inside := startSegmentX >= aabbMinX && startSegmentX <= aabbMaxX &&
		startSegmentY >= aabbMinY && startSegmentY <= aabbMaxY
	p2Inside := endSegmentX >= aabbMinX && endSegmentX <= aabbMaxX &&
		endSegmentY >= aabbMinY && endSegmentY <= aabbMaxY

	if p1Inside || p2Inside {
		return true
	}

	dx := endSegmentX - startSegmentX
	dy := endSegmentY - startSegmentY

	if dx != 0 {
		t := (aabbMinX - startSegmentX) / dx
		if t >= 0 && t <= 1 {
			y := startSegmentY + t*dy
			if y >= aabbMinY && y <= aabbMaxY {
				return true
			}
		}
		t = (aabbMaxX - startSegmentX) / dx
		if t >= 0 && t <= 1 {
			y := startSegmentY + t*dy
			if y >= aabbMinY && y <= aabbMaxY {
				return true
			}
		}
	}

	if dy != 0 {
		t := (aabbMinY - startSegmentY) / dy
		if t >= 0 && t <= 1 {
			x := startSegmentX + t*dx
			if x >= aabbMinX && x <= aabbMaxX {
				return true
			}
		}
		t = (aabbMaxY - startSegmentY) / dy
		if t >= 0 && t <= 1 {
			x := startSegmentX + t*dx
			if x >= aabbMinX && x <= aabbMaxX {
				return true
			}
		}
	}

	return false
}
