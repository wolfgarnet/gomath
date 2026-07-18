package spatial

import (
	"github.com/wolfgarnet/gomath"
	"github.com/wolfgarnet/gomath/collider"
)

type PositionedItem[T gomath.Number, ID comparable] interface {
	ID() ID
	Position() gomath.Vector2[T]
}

type BoundedItem[ID comparable] interface {
	GetBounds() any
	ID() ID
}

type Containable[T gomath.Float] interface {
	ContainedIn(d gomath.SquaredAABB2D[T]) bool
}

type BoundsCheck interface {
	Check(a, b any) bool
}

type DefaultBoundsCheck struct{}

func (bounds DefaultBoundsCheck) Check(a, b any) bool {
	return collider.Intersecting2D[float32](a, b)
	//return collider.IntersectingSquaredAABB[float32](a, b)
}

type IntegerBoundsCheck struct{}

func (bounds IntegerBoundsCheck) Check(a, b any) bool {
	return collider.IntegerIntersecting(a, b)
}
