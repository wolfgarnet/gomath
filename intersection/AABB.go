package intersection

import (
	"github.com/wolfgarnet/gomath"
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
