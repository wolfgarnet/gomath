package gomath

import "fmt"

// Ray3 is a half-line in 3D: an origin and a unit direction, extending forwards
// only.
//
// It carries the 3 for the same reason Box3, OBB3 and Plane3 do — Ray already
// names the 2D one. See Sphere for why that type needs no suffix.
//
// # Forwards only, and why that is the useful choice
//
// Every intersection here rejects a hit behind the origin. That is what makes a
// ray the right primitive for picking: the thing under the cursor is in front of
// the camera, and a line would happily report the object behind the viewer as
// being under the mouse. A caller that genuinely wants both directions can test
// the flipped ray as well.
//
// # The direction is normalised
//
// NewRay3 normalises, so every distance an intersection returns is a true
// distance in world units and results from different shapes are directly
// comparable — which is exactly what picking needs, since "which did I hit" is
// decided by taking the smallest.
//
// The 2D Ray does not normalise, because it admits integer coordinates where a
// unit direction cannot be represented. Ray3 is Float-only and has no such
// excuse.
type Ray3[T Float] struct {
	Origin    Vector3[T]
	Direction Vector3[T]
}

// NewRay3 returns a ray from an origin along a direction, which is normalised.
//
// A zero direction defines no ray; it is returned as-is rather than as NaNs, and
// IsValid reports it. Every intersection treats such a ray as hitting nothing,
// so a degenerate ray fails to pick rather than crashing or picking everything.
func NewRay3[T Float](origin Vector3[T], direction Vector3[T]) Ray3[T] {
	return Ray3[T]{Origin: origin, Direction: direction.Normalize()}
}

// NewRay3Between returns the ray from one point towards another.
//
// This is the form a picking query usually wants: the camera's position and a
// point on the near plane, or one object looking at another.
func NewRay3Between[T Float](from Vector3[T], to Vector3[T]) Ray3[T] {
	return NewRay3(from, to.Sub(from))
}

// IsValid reports whether the ray has a direction. A zero-direction ray is the
// result of asking for the ray between a point and itself.
func (r Ray3[T]) IsValid() bool {
	return !r.Direction.IsZero()
}

// At returns the point at a distance along the ray. With a unit direction the
// parameter is a true distance, so At(0) is the origin.
func (r Ray3[T]) At(distance T) Vector3[T] {
	return r.Origin.Add(r.Direction.MultiplyScalar(distance))
}

// ClosestPoint returns the point on the ray nearest to the given point.
//
// Because a ray starts at its origin rather than extending backwards, a point
// behind the origin has the ORIGIN as its nearest point — not the projection
// onto the infinite line, which would lie off the ray entirely.
func (r Ray3[T]) ClosestPoint(point Vector3[T]) Vector3[T] {
	along := point.Sub(r.Origin).Dot(r.Direction)
	if along <= 0 {
		return r.Origin
	}

	return r.At(along)
}

// DistanceTo returns the distance from a point to the nearest point on the ray.
func (r Ray3[T]) DistanceTo(point Vector3[T]) T {
	return point.Sub(r.ClosestPoint(point)).Length()
}

// DistanceToSq returns the squared distance to a point, avoiding the square root
// when only comparisons are needed.
func (r Ray3[T]) DistanceToSq(point Vector3[T]) T {
	return point.Sub(r.ClosestPoint(point)).LengthSq()
}

// Flipped returns the ray pointing the other way from the same origin, which is
// how a caller tests the half-line this one deliberately excludes.
func (r Ray3[T]) Flipped() Ray3[T] {
	return Ray3[T]{Origin: r.Origin, Direction: r.Direction.MultiplyScalar(-1)}
}

// Transformed returns the ray moved by a matrix: the origin as a position, the
// direction as a direction.
//
// The distinction matters — a direction must not pick up the matrix's
// translation — and it is what lets an oriented-box test move the ray into the
// box's local space rather than rotating the box into the world.
//
// The direction is renormalised, so a matrix with scale still yields a unit ray
// and the distances it returns stay true distances in the TARGET space.
func (r Ray3[T]) Transformed(matrix Matrix4[T]) Ray3[T] {
	origin := matrix.TransformVector4(r.Origin.ToVector4(1))
	direction := matrix.TransformVector4(r.Direction.ToVector4(0))

	return NewRay3(
		Vector3[T]{X: origin.X, Y: origin.Y, Z: origin.Z},
		Vector3[T]{X: direction.X, Y: direction.Y, Z: direction.Z},
	)
}

func (r Ray3[T]) String() string {
	return fmt.Sprintf("Ray3(from %v along %v)", r.Origin, r.Direction)
}
