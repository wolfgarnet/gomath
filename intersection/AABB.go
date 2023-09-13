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

func AABB2DPoint[T gomath.Number](a gomath.SquaredAABB2D[T], p gomath.Vector2[T]) bool {
	return a.Center.X-a.HalfSize < p.X &&
		a.Center.X+a.HalfSize >= p.X &&
		a.Center.Y-a.HalfSize < p.Y &&
		a.Center.Y+a.HalfSize >= p.Y
}

func AABB2DAABB2D[T gomath.Number](b1, b2 gomath.SquaredAABB2D[T]) bool {
	if gomath.Abs(b1.Center.X-b2.Center.X) > (b1.HalfSize + b2.HalfSize) {
		return false
	}
	if gomath.Abs(b1.Center.Y-b2.Center.Y) > (b1.HalfSize + b2.HalfSize) {
		return false
	}

	return true
}

func AABB2DAABB2DWithPenetration[T gomath.Number](b1, b2 gomath.SquaredAABB2D[T]) (penetration T, normal *gomath.Vector2[T], overlapping bool) {
	diffx := b1.Center.X - b2.Center.X
	directionx := directionRight
	if diffx < 0 {
		diffx *= -1
		directionx = directionLeft
	}
	sizex := b1.HalfSize + b2.HalfSize
	if diffx > sizex {
		return 0, nil, false
	}
	diffy := b1.Center.Y - b2.Center.Y
	directiony := directionUp
	if diffy < 0 {
		diffy *= -1
		directiony = directionDown
	}
	sizey := b1.HalfSize + b2.HalfSize
	if diffy > sizey {
		return 0, nil, false
	}

	px := sizex - diffx
	py := sizey - diffy
	if px < py {
		if directionx == directionRight {
			return px, &gomath.Vector2[T]{X: -1, Y: 0}, true
		} else {
			return px, &gomath.Vector2[T]{X: 1, Y: 0}, true
		}
	}

	if directiony == directionUp {
		return py, &gomath.Vector2[T]{X: 0, Y: -1}, true
	} else {
		return py, &gomath.Vector2[T]{X: 0, Y: 1}, true
	}
}
