package intersection

import (
	"github.com/davecgh/go-spew/spew"
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

func SquaredAABBSquaredAABB[T gomath.Number](b1, b2 gomath.SquaredAABB2D[T]) bool {
	if gomath.Abs(b1.Center.X-b2.Center.X) > (b1.HalfSize + b2.HalfSize) {
		return false
	}
	if gomath.Abs(b1.Center.Y-b2.Center.Y) > (b1.HalfSize + b2.HalfSize) {
		return false
	}

	return true
}

func AABBAABB[T gomath.Number](b1, b2 gomath.AABB[T]) bool {
	spew.Dump("COMPARING", b1, b2)
	if gomath.Abs(b1.Center.X-b2.Center.X) > (b1.HalfWidth + b2.HalfWidth) {
		return false
	}
	if gomath.Abs(b1.Center.Y-b2.Center.Y) > (b1.HalfHeight + b2.HalfHeight) {
		return false
	}

	return true
}

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
		} else {
			return px, gomath.Vector2[T]{X: 1, Y: 0}, true
		}
	}

	if directiony == directionUp {
		return py, gomath.Vector2[T]{X: 0, Y: -1}, true
	} else {
		return py, gomath.Vector2[T]{X: 0, Y: 1}, true
	}
}

func PointSideOfBox[T gomath.Number](p gomath.Vector2[T], ls1 gomath.LineSegment[T], ls2 gomath.LineSegment[T]) bool {
	left := gomath.IsLeft(p, ls1.P1, ls2.P1)
	right := gomath.IsLeft(p, ls1.P2, ls2.P2)

	return left > 0 && right < 0
}
