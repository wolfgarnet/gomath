package intersection

import (
	"github.com/wolfgarnet/gomath"
)

type Direction int

const (
	directionUp Direction = iota
	directionRight
	directionDown
	directionLeft
)

func SquaredAABBPoint[T gomath.Number](a gomath.SquaredAABB2D[T], p gomath.Vector2[T]) bool {
	return a.Center.X-a.HalfSize < p.X &&
		a.Center.X+a.HalfSize >= p.X &&
		a.Center.Y-a.HalfSize < p.Y &&
		a.Center.Y+a.HalfSize >= p.Y
}

func AABBPoint[T gomath.Number](a gomath.AABB[T], p gomath.Vector2[T]) bool {
	return a.Center.X-a.HalfWidth < p.X &&
		a.Center.X+a.HalfWidth >= p.X &&
		a.Center.Y-a.HalfHeight < p.Y &&
		a.Center.Y+a.HalfHeight >= p.Y
}

func AABBPointCollider[T gomath.Number](aabb gomath.AABBCollider[T], aabbOffset gomath.Vector2[T], p gomath.Vector2[T]) bool {
	return AABBPoint(gomath.AABB[T]{Center: aabbOffset, HalfWidth: aabb.HalfWidth, HalfHeight: aabb.HalfHeight}, p)
}

func SquaredAABBSquaredAABB[T gomath.Number](b1, b2 gomath.SquaredAABB2D[T]) bool {
	if gomath.Abs(b1.Center.X-b2.Center.X) > (b1.HalfSize + b2.HalfSize) {
		return false
	}
	if gomath.Abs(b1.Center.Y-b2.Center.Y) > (b1.HalfSize + b2.HalfSize) {
		return false
	}

	return true
}

func AABBAABBColliders[T gomath.Number](aabb1 gomath.AABBCollider[T], aabb2 gomath.AABBCollider[T], aabb1Offset gomath.Vector2[T], aabb2Offset gomath.Vector2[T]) bool {
	if gomath.Abs(aabb1Offset.X-aabb2Offset.X) > (aabb1.HalfWidth + aabb2.HalfWidth) {
		return false
	}
	if gomath.Abs(aabb1Offset.Y-aabb2Offset.Y) > (aabb1.HalfHeight + aabb2.HalfHeight) {
		return false
	}
	return true
}

func AABBAABB[T gomath.Number](b1, b2 gomath.AABB[T]) bool {
	if gomath.Abs(b1.Center.X-b2.Center.X) > (b1.HalfWidth + b2.HalfWidth) {
		return false
	}
	if gomath.Abs(b1.Center.Y-b2.Center.Y) > (b1.HalfHeight + b2.HalfHeight) {
		return false
	}

	return true
}

func AABB2LineSegment[T gomath.Number](aabb gomath.AABB2[float32], segment gomath.LineSegment[float32]) bool {
	// Check if both endpoints are inside the AABB
	p1Inside := segment.P1.X >= aabb.MinX && segment.P1.X <= aabb.MaxX &&
		segment.P1.Y >= aabb.MinY && segment.P1.Y <= aabb.MaxY
	p2Inside := segment.P2.X >= aabb.MinX && segment.P2.X <= aabb.MaxX &&
		segment.P2.Y >= aabb.MinY && segment.P2.Y <= aabb.MaxY

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

	dx := segment.P2.X - segment.P1.X
	dy := segment.P2.Y - segment.P1.Y

	// Check intersection with vertical edges (x = MinX and x = MaxX)
	if dx != 0 {
		// Left edge (x = MinX)
		t := (aabb.MinX - segment.P1.X) / dx
		if t >= 0 && t <= 1 {
			y := segment.P1.Y + t*dy
			if y >= aabb.MinY && y <= aabb.MaxY {
				return true
			}
		}

		// Right edge (x = MaxX)
		t = (aabb.MaxX - segment.P1.X) / dx
		if t >= 0 && t <= 1 {
			y := segment.P1.Y + t*dy
			if y >= aabb.MinY && y <= aabb.MaxY {
				return true
			}
		}
	}

	// Check intersection with horizontal edges (y = MinY and y = MaxY)
	if dy != 0 {
		// Bottom edge (y = MinY)
		t := (aabb.MinY - segment.P1.Y) / dy
		if t >= 0 && t <= 1 {
			x := segment.P1.X + t*dx
			if x >= aabb.MinX && x <= aabb.MaxX {
				return true
			}
		}

		// Top edge (y = MaxY)
		t = (aabb.MaxY - segment.P1.Y) / dy
		if t >= 0 && t <= 1 {
			x := segment.P1.X + t*dx
			if x >= aabb.MinX && x <= aabb.MaxX {
				return true
			}
		}
	}

	return false
}

// AABB2AABB2 returns true if the two AABBs intersect
// Only works if positioned at the origin
func AABB2AABB2[T gomath.Number](a, b gomath.AABB2[T]) bool {
	if a.MinX > b.MaxX || a.MaxX < b.MinX {
		return false
	}

	if a.MinY > b.MaxY || a.MaxY < b.MinY {
		return false
	}

	return true
}

func MinMaxVector[T gomath.Number](amin gomath.Vector2[T], bmin gomath.Vector2[T], amax gomath.Vector2[T], bmax gomath.Vector2[T]) bool {
	if amin.X > bmax.X || amax.X < bmin.X {
		return false
	}

	if amin.Y > bmax.Y || amax.Y < bmin.Y {
		return false
	}

	return true
}

func AABBSquaredAABB[T gomath.Number](b1 gomath.AABB[T], b2 gomath.SquaredAABB2D[T]) bool {
	if gomath.Abs(b1.Center.X-b2.Center.X) > (b1.HalfWidth + b2.HalfSize) {
		return false
	}
	if gomath.Abs(b1.Center.Y-b2.Center.Y) > (b1.HalfHeight + b2.HalfSize) {
		return false
	}

	return true
}

func SquaredAABB2DSquaredAABB2DWithPenetration[T gomath.Number](b1, b2 gomath.SquaredAABB2D[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	return AABB2DAABB2DWithPenetration(
		gomath.AABB[T]{
			Center:     b1.Center,
			HalfWidth:  b1.HalfSize,
			HalfHeight: b1.HalfSize,
		},
		gomath.AABB[T]{
			Center:     b2.Center,
			HalfWidth:  b2.HalfSize,
			HalfHeight: b2.HalfSize,
		},
	)
}

func AABBSquaredAABB2DWithPenetration[T gomath.Number](b1 gomath.AABB[T], b2 gomath.SquaredAABB2D[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	return AABB2DAABB2DWithPenetration(
		b1,
		gomath.AABB[T]{
			Center:     b2.Center,
			HalfWidth:  b2.HalfSize,
			HalfHeight: b2.HalfSize,
		},
	)
}

func AABBCollidersPenetration[T gomath.Number](aabb1 gomath.AABBCollider[T], aabb1Offset gomath.Vector2[T], aabb2 gomath.AABBCollider[T], aabb2Offset gomath.Vector2[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	diffx := aabb1Offset.X - aabb2Offset.X
	directionx := directionRight
	if diffx < 0 {
		diffx *= -1
		directionx = directionLeft
	}
	sizex := aabb1.HalfWidth + aabb2.HalfWidth
	if diffx > sizex {
		return 0, normal, false
	}
	diffy := aabb1Offset.Y - aabb2Offset.Y
	directiony := directionUp
	if diffy < 0 {
		diffy *= -1
		directiony = directionDown
	}
	sizey := aabb1.HalfHeight + aabb2.HalfHeight
	if diffy > sizey {
		return 0, normal, false
	}

	px := sizex - diffx
	py := sizey - diffy
	if px < py {
		if directionx == directionRight {
			return px, gomath.Vector2[T]{X: -1, Y: 0}, true
		}
		return px, gomath.Vector2[T]{X: 1, Y: 0}, true
	}

	if directiony == directionUp {
		return py, gomath.Vector2[T]{X: 0, Y: -1}, true
	}
	return py, gomath.Vector2[T]{X: 0, Y: 1}, true
}

func AABB2DAABB2DWithPenetration[T gomath.Number](b1, b2 gomath.AABB[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	diffx := b1.Center.X - b2.Center.X
	directionx := directionRight
	if diffx < 0 {
		diffx *= -1
		directionx = directionLeft
	}
	sizex := b1.HalfWidth + b2.HalfWidth
	if diffx > sizex {
		return 0, normal, false
	}
	diffy := b1.Center.Y - b2.Center.Y
	directiony := directionUp
	if diffy < 0 {
		diffy *= -1
		directiony = directionDown
	}
	sizey := b1.HalfHeight + b2.HalfHeight
	if diffy > sizey {
		return 0, normal, false
	}

	px := sizex - diffx
	py := sizey - diffy
	if px < py {
		if directionx == directionRight {
			return px, gomath.Vector2[T]{X: -1, Y: 0}, true
		}
		return px, gomath.Vector2[T]{X: 1, Y: 0}, true
	}

	if directiony == directionUp {
		return py, gomath.Vector2[T]{X: 0, Y: -1}, true
	}
	return py, gomath.Vector2[T]{X: 0, Y: 1}, true
}

func PointAABB2DPenetration[T gomath.Number](p gomath.Vector2[T], b gomath.AABB[T]) (penetration T, normal gomath.Vector2[float32], inside bool) {
	dx := p.X - b.Center.X
	dy := p.Y - b.Center.Y

	px := b.HalfWidth - gomath.Abs(dx)
	py := b.HalfHeight - gomath.Abs(dy)

	if px < 0 || py < 0 {
		return 0, normal, false
	}

	if px < py {
		if dx < 0 {
			return px, gomath.Vector2[float32]{X: -1, Y: 0}, true
		}
		return px, gomath.Vector2[float32]{X: 1, Y: 0}, true
	}
	if dy < 0 {
		return py, gomath.Vector2[float32]{X: 0, Y: -1}, true
	}
	return py, gomath.Vector2[float32]{X: 0, Y: 1}, true
}

// PointSideOfBox returns true if the point is inside the box defined by the two line segments
func PointSideOfBox[T gomath.Number](p gomath.Vector2[T], ls1 gomath.LineSegment[T], ls2 gomath.LineSegment[T]) bool {
	left := gomath.IsLeft(p, ls1.P1, ls2.P1)
	right := gomath.IsLeft(p, ls1.P2, ls2.P2)

	return left > 0 && right < 0
}
