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

func (r Ray[T]) AtDistance(distance T, length T, segment *LineSegment[T]) {
	segment.P1 = r.Origin.Add(r.Direction.MultiplyScalar(distance))
	segment.P2 = segment.P1.Add(r.Direction.MultiplyScalar(length))
}
