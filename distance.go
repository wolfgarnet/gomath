package gomath

import (
	"fmt"
)

func dist2[T Number](v, w Vector2[T]) T { return Sqr(v.X-w.X) + Sqr(v.Y-w.Y) }

// MinimumDistance calculates the minimum distance from a point, p,
// to the line segment defined by a and b.
// https://stackoverflow.com/questions/849211/shortest-distance-between-a-point-and-a-line-segment
func MinimumDistance[T Number](a, b, p Vector2[T]) T {
	l2 := dist2(a, b)
	if l2 == 0 {
		return p.Sub(a).Length()
	}

	t := ((p.X-a.X)*(b.X-a.X) + (p.Y-a.Y)*(b.Y-a.Y)) / l2
	t = Max(0, Min(1, t))

	tmp := NewVector2(
		a.X+t*(b.X-a.X),
		a.Y+t*(b.Y-a.Y),
	)

	return p.Sub(tmp).Length()
}

func DistanceTo[T Number](o1 any, o2 any) T {
	switch v1 := o1.(type) {
	case Vector2[T]:
		return DistanceFromPoint(v1, o2)
	case *Vector2[T]:
		return DistanceFromPoint(*v1, o2)

	case AABB[T]:
		return DistanceFromAABB(v1, o2)
	case *AABB[T]:
		return DistanceFromAABB(*v1, o2)

	case Circle[T]:
		return DistanceFromCircle(v1, o2)
	case *Circle[T]:
		return DistanceFromCircle(*v1, o2)

	default:
		panic(fmt.Sprintf("unknown type: %T", o1))
	}
}

func DistanceFromCircle[T Number](circle Circle[T], o2 any) T {
	switch v2 := o2.(type) {
	case Vector2[T]:
		return DistanceFromPointToCircle(v2, circle)
	case *Vector2[T]:
		return DistanceFromPointToCircle(*v2, circle)

	case AABB[T]:
		return DistanceFromCircleToAABB(circle, v2)
	case *AABB[T]:
		return DistanceFromCircleToAABB(circle, *v2)

	default:
		panic(fmt.Sprintf("unknown type: %T", o2))
	}
}

func DistanceFromAABB[T Number](aabb AABB[T], o2 any) T {
	switch v2 := o2.(type) {
	case Vector2[T]:
		return DistanceFromPointToAABB(v2, aabb)
	case *Vector2[T]:
		return DistanceFromPointToAABB(*v2, aabb)

	case Circle[T]:
		return DistanceFromCircleToAABB(v2, aabb)
	case *Circle[T]:
		return DistanceFromCircleToAABB(*v2, aabb)

	/*case AABB[T]:
		return DistanceFromPointToAABB(point, v2)
	case *AABB[T]:
		return DistanceFromPointToAABB(point, *v2)*/

	default:
		panic(fmt.Sprintf("unknown type: %T", o2))
	}
}

func DistanceFromPoint[T Number](point Vector2[T], o2 any) T {
	switch v2 := o2.(type) {
	case Vector2[T]:
		return point.DistanceTo(v2)
	case *Vector2[T]:
		return point.DistanceTo(*v2)
	case AABB[T]:
		return DistanceFromPointToAABB(point, v2)
	case *AABB[T]:
		return DistanceFromPointToAABB(point, *v2)

	default:
		panic(fmt.Sprintf("unknown type: %T", o2))
	}
}

func DistanceFromPointToCircle[T Number](point Vector2[T], circle Circle[T]) T {
	return point.DistanceTo(circle.Center) - circle.Radius
}

func DistanceFromCircleToAABB[T Number](circle Circle[T], aabb AABB[T]) T {
	return DistanceFromPointToAABB(circle.Center, aabb) - circle.Radius
}

func DistanceFromPointToAABB[T Number](point Vector2[T], aabb AABB[T]) T {
	min := Vector2[T]{
		X: aabb.Center.X - aabb.HalfWidth,
		Y: aabb.Center.Y - aabb.HalfHeight,
	}
	max := Vector2[T]{
		X: aabb.Center.X + aabb.HalfWidth,
		Y: aabb.Center.Y + aabb.HalfHeight,
	}

	clamped := Vector2[T]{
		X: clamp(point.X, min.X, max.X),
		Y: clamp(point.Y, min.Y, max.Y),
	}

	clamped = clamped.Sub(point)
	return clamped.Length()
}

func clamp[T Number](v, min, max T) T {
	return Max(min, Min(max, v))
}
