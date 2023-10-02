package gomath

import "fmt"

type Ray2D[T Number] struct {
	Origin    Vector2[T]
	Direction Vector2[T]
}

func NewRay2D[T Number](origin, direction Vector2[T]) Ray2D[T] {
	return Ray2D[T]{
		Origin:    origin,
		Direction: direction,
	}
}

func (r Ray2D[T]) String() string {
	return fmt.Sprintf("%v->%v", r.Origin, r.Direction)
}
