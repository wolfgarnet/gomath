package gomath

import (
	"fmt"
)

type SquaredAABB2D[T Number] struct {
	Center   Vector2[T]
	HalfSize T

	calculated bool
	minx, miny T
	maxx, maxy T
}

func NewSquaredAABB2DXY[T Number](x, y, halfSize T) SquaredAABB2D[T] {
	return SquaredAABB2D[T]{
		Center:   Vector2[T]{X: x, Y: y},
		HalfSize: halfSize,
	}
}

func NewSquaredAABB2D[T Number](center Vector2[T], halfSize T) SquaredAABB2D[T] {
	return SquaredAABB2D[T]{
		Center:   center,
		HalfSize: halfSize,
	}
}

func (a SquaredAABB2D[T]) String() string {
	return fmt.Sprintf("AABB[c=%v s=%v]", a.Center, a.HalfSize)
}
