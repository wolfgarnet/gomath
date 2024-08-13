package gomath

import (
	"fmt"
	"image"
	"math"
)

type AABB[T Number] struct {
	Center     Vector2[T]
	HalfWidth  T
	HalfHeight T
}

func NewAABBFromRectangle[T Number](rect image.Rectangle) AABB[T] {
	// Make an AABB[T] out of the Rectangle rect
	centerX := T(rect.Min.X+rect.Max.X) / 2
	centerY := T(rect.Min.Y+rect.Max.Y) / 2

	// Compute the half width and half height of the rectangle
	halfWidth := T(rect.Max.X-rect.Min.X) / 2
	halfHeight := T(rect.Max.Y-rect.Min.Y) / 2

	// Create and return the AABB
	return AABB[T]{
		Center:     Vector2[T]{X: centerX, Y: centerY},
		HalfWidth:  halfWidth,
		HalfHeight: halfHeight,
	}
}

func NewAABB[T Number](center Vector2[T], halfWidth T, halfHeight T) AABB[T] {
	return AABB[T]{
		Center:     center,
		HalfWidth:  halfWidth,
		HalfHeight: halfHeight,
	}
}

func (a *AABB[T]) SetCenter(c Vector2[T]) {
	a.Center = c
}

func (a *AABB[T]) GetCenter() Vector2[T] {
	return a.Center
}

func (a *AABB[T]) GetHalfWidth() T {
	return a.HalfWidth
}

func (a *AABB[T]) GetHalfHeight() T {
	return a.HalfHeight
}

func (a *AABB[T]) SetHalfWidth(t T) {
	a.HalfWidth = t
}

func (a *AABB[T]) SetHalfHeight(t T) {
	a.HalfHeight = t
}

func (a AABB[T]) Extend(w T) AABB[T] {
	return AABB[T]{
		Center:     a.Center,
		HalfWidth:  a.HalfWidth + w,
		HalfHeight: a.HalfHeight + w,
	}
}

func (a *AABB[T]) Sides() []LineSegment[T] {
	segments := make([]LineSegment[T], 4)

	p1 := Vector2[T]{
		X: a.Center.X - a.HalfWidth,
		Y: a.Center.Y + a.HalfHeight,
	}
	p2 := Vector2[T]{
		X: a.Center.X + a.HalfWidth,
		Y: a.Center.Y + a.HalfHeight,
	}
	p3 := Vector2[T]{
		X: a.Center.X + a.HalfWidth,
		Y: a.Center.Y - a.HalfHeight,
	}
	p4 := Vector2[T]{
		X: a.Center.X - a.HalfWidth,
		Y: a.Center.Y - a.HalfHeight,
	}

	segments[0] = NewLineSegment(p1, p2)
	segments[1] = NewLineSegment(p2, p3)
	segments[2] = NewLineSegment(p3, p4)
	segments[3] = NewLineSegment(p4, p1)

	return segments
}

func (a *AABB[T]) Polygon() Polygon[T] {
	p1 := Vector2[T]{
		X: a.Center.X - a.HalfWidth,
		Y: a.Center.Y + a.HalfHeight,
	}
	p2 := Vector2[T]{
		X: a.Center.X + a.HalfWidth,
		Y: a.Center.Y + a.HalfHeight,
	}
	p3 := Vector2[T]{
		X: a.Center.X + a.HalfWidth,
		Y: a.Center.Y - a.HalfHeight,
	}
	p4 := Vector2[T]{
		X: a.Center.X - a.HalfWidth,
		Y: a.Center.Y - a.HalfHeight,
	}
	return NewPolygon([]Vector2[T]{p1, p2, p3, p4})
}

func (a *AABB[T]) String() string {
	return fmt.Sprintf("AABB[c=%v w=%v, h=%v]", a.Center, a.HalfWidth, a.HalfHeight)
}

func (a *AABB[T]) Points() []Vector2[T] {
	p1 := Vector2[T]{
		X: a.Center.X - a.HalfWidth,
		Y: a.Center.Y + a.HalfHeight,
	}
	p2 := Vector2[T]{
		X: a.Center.X + a.HalfWidth,
		Y: a.Center.Y + a.HalfHeight,
	}
	p3 := Vector2[T]{
		X: a.Center.X + a.HalfWidth,
		Y: a.Center.Y - a.HalfHeight,
	}
	p4 := Vector2[T]{
		X: a.Center.X - a.HalfWidth,
		Y: a.Center.Y - a.HalfHeight,
	}

	return []Vector2[T]{
		p1, p2, p3, p4,
	}
}

func ShortestDistance[T Float](a AABB[T], b AABB[T]) T {
	var closest = T(math.MaxFloat32)
	for _, a1 := range a.Points() {
		for _, a2 := range b.Points() {
			d := a1.DistanceToSquared(a2)
			if d < closest {
				closest = d
			}
		}
	}

	return Sqrt(closest)
}

type AABB2[T Number] struct {
	MaxX T
	MinX T
	MaxY T
	MinY T
}

func NewAABB2FromRectangle[T Number](rect image.Rectangle) AABB2[T] {
	return AABB2[T]{
		MaxX: T(rect.Max.X),
		MaxY: T(rect.Max.Y),
		MinX: T(rect.Min.X),
		MinY: T(rect.Min.Y),
	}
}

func NewAABB2[T Number](maxx, minx, maxy, miny T) AABB2[T] {
	return AABB2[T]{
		MaxX: maxx,
		MinX: minx,
		MaxY: maxy,
		MinY: miny,
	}
}

type SquaredAABB2D[T Number] struct {
	Center   Vector2[T]
	HalfSize T

	/*calculated bool
	minx, miny T
	maxx, maxy T*/
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

func (a SquaredAABB2D[T]) GetCenter() Vector2[float32] {
	return a.GetCenter()
}

func (a SquaredAABB2D[T]) Sides() []LineSegment[T] {
	segments := make([]LineSegment[T], 4)

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

func (a SquaredAABB2D[T]) Polygon() Polygon[T] {

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
	return NewPolygon([]Vector2[T]{p1, p2, p3, p4})
}

func (a SquaredAABB2D[T]) String() string {
	return fmt.Sprintf("SquaredAABB[c=%v hs=%v]", a.Center, a.HalfSize)
}
