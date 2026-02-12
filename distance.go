package gomath

import (
	"fmt"
)

func dist2[T Number](v, w Vector2[T]) T   { return Sqr(v.X-w.X) + Sqr(v.Y-w.Y) }
func dist22[T Number](ax, ay, bx, by T) T { return Sqr(ax-bx) + Sqr(ay-by) }

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

func MinimumDistanceValues[T Number](ax, ay, bx, by, px, py T) T {
	l2 := dist22(ax, ay, bx, by)
	if l2 == 0 {
		return LengthVectors(px, py, ax, ay)
	}

	t := ((px-ax)*(bx-ax) + (py-ay)*(by-ay)) / l2
	t = Max(0, Min(1, t))

	tmpx, tmpy := ax+t*(bx-ax), ay+t*(by-ay)

	return LengthVectors(px, py, tmpx, tmpy)
}

func VectorToMinimumDistance[T Number](vector1X, vector1Y, vector2X, vector2Y, pointX, pointY T) (dx T, dy T, l T) {
	length := dist22(vector1X, vector1Y, vector2X, vector2Y)
	if length == 0 {
		// If there's no distance between the two vectors
		dx, dy = VectorSubtract(pointX, pointY, vector1X, vector1Y)
		return dx, dy, LengthVectors(pointX, pointY, vector1X, vector1Y)
	}

	t := ((pointX-vector1X)*(vector2X-vector1X) + (pointY-vector1Y)*(vector2Y-vector1Y)) / length
	t = Max(0, Min(1, t))

	tmpx, tmpy := vector1X+t*(vector2X-vector1X), vector1Y+t*(vector2Y-vector1Y)
	dx, dy = VectorSubtract(pointX, pointY, tmpx, tmpy)

	return dx, dy, LengthVectors(pointX, pointY, tmpx, tmpy)
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

	case Circle[T]:
		dist := circle.Center.DistanceTo(v2.Center)
		return dist - circle.Radius - v2.Radius
	case *Circle[T]:
		dist := circle.Center.DistanceTo(v2.Center)
		return dist - circle.Radius - v2.Radius

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
