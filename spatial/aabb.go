package spatial

import (
	"github.com/wolfgarnet/gomath"
	"golang.org/x/exp/constraints"
)

type AABB32 struct {
	CenterX, CenterY float32
	HalfWidth        float32
	HalfHeight       float32
}

func AABBAABB32(b1, b2 AABB32) bool {
	dx := b1.CenterX - b2.CenterX
	if dx < 0 {
		dx = -dx
	}
	if dx > b1.HalfWidth+b2.HalfWidth {
		return false
	}

	dy := b1.CenterY - b2.CenterY
	if dy < 0 {
		dy = -dy
	}
	if dy > b1.HalfHeight+b2.HalfHeight {
		return false
	}
	return true
}

type AABB[T constraints.Float] struct {
	CenterX, CenterY T
	HalfWidth        T
	HalfHeight       T
}

func AABBAABB[T constraints.Float](b1, b2 AABB[T]) bool {
	dx := b1.CenterX - b2.CenterX
	if dx < 0 {
		dx = -dx
	}
	if dx > b1.HalfWidth+b2.HalfWidth {
		return false
	}

	dy := b1.CenterY - b2.CenterY
	if dy < 0 {
		dy = -dy
	}
	if dy > b1.HalfHeight+b2.HalfHeight {
		return false
	}
	return true
}

type AABB2[T constraints.Float] struct {
	Center     [2]T
	HalfWidth  T
	HalfHeight T
}

func AABBAABB2[T constraints.Float](b1, b2 AABB2[T]) bool {
	dx := b1.Center[0] - b2.Center[0]
	if dx < 0 {
		dx = -dx
	}
	if dx > b1.HalfWidth+b2.HalfWidth {
		return false
	}

	dy := b1.Center[1] - b2.Center[1]
	if dy < 0 {
		dy = -dy
	}
	if dy > b1.HalfHeight+b2.HalfHeight {
		return false
	}
	return true
}

func AABBAABB3[T constraints.Float](b1, b2 gomath.AABB[T]) bool {
	dx := b1.Center.X - b2.Center.X
	if dx < 0 {
		dx = -dx
	}
	if dx > b1.HalfWidth+b2.HalfWidth {
		return false
	}

	dy := b1.Center.Y - b2.Center.Y
	if dy < 0 {
		dy = -dy
	}
	if dy > b1.HalfHeight+b2.HalfHeight {
		return false
	}
	return true
}

func SquaredAABBAABB[T constraints.Float](b1 gomath.SquaredAABB2D[T], b2 gomath.AABB[T]) bool {
	dx := b1.Center.X - b2.Center.X
	if dx < 0 {
		dx = -dx
	}
	if dx > b1.HalfSize+b2.HalfWidth {
		return false
	}

	dy := b1.Center.Y - b2.Center.Y
	if dy < 0 {
		dy = -dy
	}
	if dy > b1.HalfSize+b2.HalfHeight {
		return false
	}
	return true
}
