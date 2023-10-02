package gomath

import "fmt"

type Ray[T Number] struct {
	Origin    Vector2[T]
	Direction Vector2[T]
}

func NewRay[T Number](origin, direction Vector2[T]) Ray[T] {
	return Ray[T]{
		Origin:    origin,
		Direction: direction,
	}
}

func (r Ray[T]) String() string {
	return fmt.Sprintf("%v->%v", r.Origin, r.Direction)
}
