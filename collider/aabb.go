package collider

import "github.com/wolfgarnet/gomath"

func AABBPoint[T gomath.Number](aabbPositionX T, aabbPositionY, aabbHalfWidth T, aabbHalfHeight T, pointx T, pointy T) bool {
	return aabbPositionX-aabbHalfWidth < pointx &&
		aabbPositionX+aabbHalfWidth >= pointx &&
		aabbPositionY-aabbHalfHeight < pointy &&
		aabbPositionY+aabbHalfHeight >= pointy
}

func AABBAABB[T gomath.Number](aabbPosition1x T, aabbPosition1y T, aabb1HalfWidth T, aabb1HalfHeight T, aabbPosition2x T, aabbPosition2y T, aabb2HalfWidth T, aabb2HalfHeight T) bool {
	if gomath.Abs(aabbPosition1x-aabbPosition2x) > (aabb1HalfWidth + aabb2HalfWidth) {
		return false
	}
	if gomath.Abs(aabbPosition1y-aabbPosition2y) > (aabb1HalfHeight + aabb2HalfHeight) {
		return false
	}
	return true
}

func AABBAABBPenetration[T gomath.Number](aabbPosition1x T, aabbPosition1y T, aabb1HalfWidth T, aabb1HalfHeight T, aabbPosition2x T, aabbPosition2y T, aabb2HalfWidth T, aabb2HalfHeight T) (penetration T, normal gomath.Vector2[T], overlapping bool) {
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

	// diff = position1 - position2, so the direction from box1 toward box2 is
	// -diff. The normal points from box1 toward box2 (the obstacle), matching the
	// convention of CircleCircle/CircleAABB.
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

func AABBLineSegment[T gomath.Number](aabbPositionX T, aabbPositionY, aabbHalfWidth T, aabbHalfHeight T, startSegmentX T, startSegmentY T, endSegmentX T, endSegmentY T) bool {
	return AABB2LineSegment(aabbPositionX-aabbHalfWidth, aabbPositionY-aabbHalfHeight, aabbPositionX+aabbHalfWidth, aabbPositionY+aabbHalfHeight, startSegmentX, startSegmentY, endSegmentX, endSegmentY)
}

func AABB2LineSegment[T gomath.Number](aabbMinX T, aabbMinY, aabbMaxX T, aabbMaxY T, startSegmentX T, startSegmentY T, endSegmentX T, endSegmentY T) bool {
	// Check if both endpoints are inside the AABB
	p1Inside := startSegmentX >= aabbMinX && startSegmentX <= aabbMaxX &&
		startSegmentY >= aabbMinY && startSegmentY <= aabbMaxY
	p2Inside := endSegmentX >= aabbMinX && endSegmentX <= aabbMaxX &&
		endSegmentY >= aabbMinY && endSegmentY <= aabbMaxY

	// If both points are inside, the segment is contained
	if p1Inside && p2Inside {
		return true
	}

	// If at least one point is inside, there's an intersection
	if p1Inside || p2Inside {
		return true
	}

	// Check if the line segment intersects any of the AABB edges
	// Using parametric line equation: P = P1 + t * (P2 - P1), where 0 <= t <= 1

	dx := endSegmentX - startSegmentX
	dy := endSegmentY - startSegmentY

	// Check intersection with vertical edges (x = MinX and x = MaxX)
	if dx != 0 {
		// Left edge (x = MinX)
		t := (aabbMinX - startSegmentX) / dx
		if t >= 0 && t <= 1 {
			y := startSegmentY + t*dy
			if y >= aabbMinY && y <= aabbMaxY {
				return true
			}
		}

		// Right edge (x = MaxX)
		t = (aabbMaxX - startSegmentX) / dx
		if t >= 0 && t <= 1 {
			y := startSegmentY + t*dy
			if y >= aabbMinY && y <= aabbMaxY {
				return true
			}
		}
	}

	// Check intersection with horizontal edges (y = MinY and y = MaxY)
	if dy != 0 {
		// Bottom edge (y = MinY)
		t := (aabbMinY - startSegmentY) / dy
		if t >= 0 && t <= 1 {
			x := startSegmentX + t*dx
			if x >= aabbMinX && x <= aabbMaxX {
				return true
			}
		}

		// Top edge (y = MaxY)
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

func AABBLineSegmentPenetration[T gomath.Number](aabbPositionX T, aabbPositionY, aabbHalfWidth T, aabbHalfHeight T, startSegmentX T, startSegmentY T, endSegmentX T, endSegmentY T) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	aabbMinX := aabbPositionX - aabbHalfWidth
	aabbMinY := aabbPositionY - aabbHalfHeight
	aabbMaxX := aabbPositionX + aabbHalfWidth
	aabbMaxY := aabbPositionY + aabbHalfHeight

	overlapping = AABB2LineSegment(aabbMinX, aabbMinY, aabbMaxX, aabbMaxY, startSegmentX, startSegmentY, endSegmentX, endSegmentY)
	if !overlapping {
		return 0, normal, false
	}

	// If it overlaps, we find the "penetration" as the vector from the segment's closest point
	// to the AABB center, or similar.
	// Actually, a more useful penetration for a segment vs AABB is often the distance
	// required to push the segment out of the AABB along the shortest path.

	// We can check the segment's points.
	p1Inside := startSegmentX >= aabbMinX && startSegmentX <= aabbMaxX &&
		startSegmentY >= aabbMinY && startSegmentY <= aabbMaxY
	p2Inside := endSegmentX >= aabbMinX && endSegmentX <= aabbMaxX &&
		endSegmentY >= aabbMinY && endSegmentY <= aabbMaxY

	if p1Inside && p2Inside {
		// Both inside, penetration is tricky to define. Let's find which point is closer to an edge.
		// Or use the center of the segment.
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
		// Since it's COMPLETELY inside, we might want to add the distance from mid point to the furthest point of segment?
		// No, usually penetration is just "how much to move to stop overlapping".
		return penetration, normal, true
	}

	// If one point is inside, find its penetration.
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

	// Both points are outside, but they intersect (the segment crosses the AABB).
	// Find the closest point on the segment to the AABB center.
	dx, dy, _ := gomath.VectorToMinimumDistance(startSegmentX, startSegmentY, endSegmentX, endSegmentY, aabbPositionX, aabbPositionY)
	// dx, dy is from closest point on segment to aabbPosition
	// So closest point on segment is (aabbPositionX + dx, aabbPositionY + dy)
	// Wait, VectorToMinimumDistance(v1, v2, p) returns (p - closestPointOnSegment).
	// So closestPointOnSegment = (pointX - dx, pointY - dy)

	closestX := aabbPositionX - dx
	closestY := aabbPositionY - dy

	// Penetration for this point:
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
