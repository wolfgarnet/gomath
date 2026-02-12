package collider

import "github.com/wolfgarnet/gomath"

func AABBPoint[T gomath.Number](aabbPositionX T, aabbPositionY, aabbHalfWidth T, aabbHalfHeight T, pointx T, pointy T) bool {
	return aabbPositionX-aabbHalfWidth < pointx &&
		aabbPositionX+aabbHalfWidth >= pointx &&
		aabbPositionY-aabbHalfHeight < pointy &&
		aabbPositionY+aabbHalfHeight >= pointy
}

func AABBAABB[T gomath.Number](aabbPosition1x T, aabbPosition1y T, aabbPosition2x T, aabbPosition2y T, aabb1HalfWidth T, aabb1HalfHeight T, aabb2HalfWidth T, aabb2HalfHeight T) bool {
	if gomath.Abs(aabbPosition1x-aabbPosition2x) > (aabb1HalfWidth + aabb2HalfWidth) {
		return false
	}
	if gomath.Abs(aabbPosition1y-aabbPosition2y) > (aabb1HalfHeight + aabb2HalfHeight) {
		return false
	}
	return true
}

func AABBAABBPenetration[T gomath.Number](aabbPosition1x T, aabbPosition1y T, aabbPosition2x T, aabbPosition2y T, aabb1HalfWidth T, aabb1HalfHeight T, aabb2HalfWidth T, aabb2HalfHeight T) (penetration T, normal gomath.Vector2[T], overlapping bool) {
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

	if px < py {
		penetration = px
		if diffx >= 0 {
			normal = gomath.Vector2[T]{X: 1, Y: 0}
		} else {
			normal = gomath.Vector2[T]{X: -1, Y: 0}
		}
	} else {
		penetration = py
		if diffy >= 0 {
			normal = gomath.Vector2[T]{X: 0, Y: 1}
		} else {
			normal = gomath.Vector2[T]{X: 0, Y: -1}
		}
	}

	return penetration, normal, true
}

func AABBLineSegment[T gomath.Number](aabbPositionX T, aabbPositionY, aabbHalfWidth T, aabbHalfHeight T, startx T, starty T, endx T, endy T) bool {
	return AABB2LineSegment(aabbPositionX-aabbHalfWidth, aabbPositionY-aabbHalfHeight, aabbPositionX+aabbHalfWidth, aabbPositionY+aabbHalfHeight, startx, starty, endx, endy)
}

func AABB2LineSegment[T gomath.Number](aabbMinX T, aabbMinY, aabbMaxX T, aabbMaxY T, startx T, starty T, endx T, endy T) bool {
	// Check if both endpoints are inside the AABB
	p1Inside := startx >= aabbMinX && startx <= aabbMaxX &&
		starty >= aabbMinY && starty <= aabbMaxY
	p2Inside := endx >= aabbMinX && endx <= aabbMaxX &&
		endy >= aabbMinY && endy <= aabbMaxY

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

	dx := endx - startx
	dy := endy - starty

	// Check intersection with vertical edges (x = MinX and x = MaxX)
	if dx != 0 {
		// Left edge (x = MinX)
		t := (aabbMinX - startx) / dx
		if t >= 0 && t <= 1 {
			y := starty + t*dy
			if y >= aabbMinY && y <= aabbMaxY {
				return true
			}
		}

		// Right edge (x = MaxX)
		t = (aabbMaxX - startx) / dx
		if t >= 0 && t <= 1 {
			y := starty + t*dy
			if y >= aabbMinY && y <= aabbMaxY {
				return true
			}
		}
	}

	// Check intersection with horizontal edges (y = MinY and y = MaxY)
	if dy != 0 {
		// Bottom edge (y = MinY)
		t := (aabbMinY - starty) / dy
		if t >= 0 && t <= 1 {
			x := startx + t*dx
			if x >= aabbMinX && x <= aabbMaxX {
				return true
			}
		}

		// Top edge (y = MaxY)
		t = (aabbMaxY - starty) / dy
		if t >= 0 && t <= 1 {
			x := startx + t*dx
			if x >= aabbMinX && x <= aabbMaxX {
				return true
			}
		}
	}

	return false
}

func AABBLineSegmentPenetration[T gomath.Number](aabbPositionX T, aabbPositionY, aabbHalfWidth T, aabbHalfHeight T, startx T, starty T, endx T, endy T) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	return penetration, normal, false
}
