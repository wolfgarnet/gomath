package spatial

import (
	"github.com/wolfgarnet/gomath"
	intersection "github.com/wolfgarnet/gomath/intersection"
)

type BoundedItem interface {
	comparable
	GetBounds() any
}

type Containable[T gomath.Float] interface {
	ContainedIn(d gomath.SquaredAABB2D[T]) bool
}

type BoundsCheck interface {
	Check(a, b any) bool
}

type DefaultBoundsCheck struct{}

func (bounds DefaultBoundsCheck) Check(a, b any) bool {
	return intersection.Intersecting2D[float32](a, b)
	//return intersection.IntersectingSquaredAABB[float32](a, b)
}

type IntegerBoundsCheck struct{}

func (bounds IntegerBoundsCheck) Check(a, b any) bool {
	return intersection.IntegerIntersecting(a, b)
}
