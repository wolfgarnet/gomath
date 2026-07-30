package gomath

import "fmt"

// Plane3 is an infinite plane in 3D, stored as a normal and a signed distance
// from the origin along that normal.
//
// The plane is the set of points p satisfying
//
//	Dot(Normal, p) + Distance == 0
//
// so SignedDistance is positive on the side the normal points towards. Frustum
// builds its planes with the normals pointing INWARD, which is what makes
// "inside" mean "positive on all six".
//
// Most methods assume a unit normal. Normalized restores that after any
// construction that might not produce one — notably extraction from a matrix.
type Plane3[T Float] struct {
	Normal   Vector3[T]
	Distance T
}

// NewPlane3 returns the plane with the given normal and signed origin distance.
// The normal is not normalised; call Normalized if it might not be unit length.
func NewPlane3[T Float](normal Vector3[T], distance T) Plane3[T] {
	return Plane3[T]{Normal: normal, Distance: distance}
}

// NewPlane3FromPointNormal returns the plane through a point with the given
// normal, which is normalised.
func NewPlane3FromPointNormal[T Float](point Vector3[T], normal Vector3[T]) Plane3[T] {
	unit := normal.Normalize()

	return Plane3[T]{Normal: unit, Distance: -unit.Dot(point)}
}

// NewPlane3FromPoints returns the plane through three points, with the normal
// following the right-hand rule for counter-clockwise winding — the same
// convention the geometry primitives use.
//
// It reports false when the points are collinear, since they then define no
// unique plane.
func NewPlane3FromPoints[T Float](a Vector3[T], b Vector3[T], c Vector3[T]) (Plane3[T], bool) {
	normal := b.Sub(a).Cross(c.Sub(a))
	if normal.LengthSq() == 0 {
		return Plane3[T]{}, false
	}

	return NewPlane3FromPointNormal(a, normal), true
}

// Normalized returns the same plane with a unit normal, so that SignedDistance
// returns a true distance rather than one scaled by the normal's length.
//
// A zero normal defines no plane, so it is returned unchanged rather than
// producing infinities.
func (p Plane3[T]) Normalized() Plane3[T] {
	lengthSq := p.Normal.LengthSq()
	if lengthSq == 0 {
		return p
	}

	inverse := 1 / Sqrt(lengthSq)

	return Plane3[T]{
		Normal:   p.Normal.MultiplyScalar(inverse),
		Distance: p.Distance * inverse,
	}
}

// SignedDistance returns the distance from the point to the plane, positive on
// the side the normal points towards and negative behind it. It is a true
// distance only for a unit normal.
func (p Plane3[T]) SignedDistance(point Vector3[T]) T {
	return p.Normal.Dot(point) + p.Distance
}

// IsInFront reports whether the point lies on the normal's side, or on the
// plane itself.
func (p Plane3[T]) IsInFront(point Vector3[T]) bool {
	return p.SignedDistance(point) >= 0
}

// ClosestPoint returns the point on the plane nearest the given point, which is
// the point projected along the normal. It assumes a unit normal.
func (p Plane3[T]) ClosestPoint(point Vector3[T]) Vector3[T] {
	return point.Sub(p.Normal.MultiplyScalar(p.SignedDistance(point)))
}

// Flipped returns the same plane facing the other way, so what was in front is
// now behind.
func (p Plane3[T]) Flipped() Plane3[T] {
	return Plane3[T]{Normal: p.Normal.Negate(), Distance: -p.Distance}
}

// Origin returns the point on the plane closest to the world origin.
func (p Plane3[T]) Origin() Vector3[T] {
	return p.Normal.MultiplyScalar(-p.Distance)
}

// ApproxEquals reports whether normal and distance are both within tol.
func (p Plane3[T]) ApproxEquals(other Plane3[T], tol T) bool {
	return Abs(p.Normal.X-other.Normal.X) <= tol &&
		Abs(p.Normal.Y-other.Normal.Y) <= tol &&
		Abs(p.Normal.Z-other.Normal.Z) <= tol &&
		Abs(p.Distance-other.Distance) <= tol
}

func (p Plane3[T]) String() string {
	return fmt.Sprintf("Plane3(normal %v, distance %v)", p.Normal, p.Distance)
}
