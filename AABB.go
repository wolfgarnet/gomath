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

func (a SquaredAABB2D[T]) Sides() []LineSegment[T] {
	segments := make([]LineSegment[T], 0, 4)

	p1 := Vector2[T]{
		X: a.Center.X - a.HalfSize,
		Y: a.Center.Y + a.HalfSize,
	}
	p2 := Vector2[T]{
		X: a.Center.X + a.HalfSize,
		Y: a.Center.Y + a.HalfSize,
	}
	p3 := Vector2[T]{
		X: a.Center.X + a.HalfSize,
		Y: a.Center.Y - a.HalfSize,
	}
	p4 := Vector2[T]{
		X: a.Center.X - a.HalfSize,
		Y: a.Center.Y - a.HalfSize,
	}

	segments[0] = NewLineSegment(p1, p2)
	segments[1] = NewLineSegment(p2, p3)
	segments[2] = NewLineSegment(p3, p4)
	segments[3] = NewLineSegment(p4, p1)

	return segments
}

func (a SquaredAABB2D[T]) String() string {
	return fmt.Sprintf("AABB[c=%v s=%v]", a.Center, a.HalfSize)
}
