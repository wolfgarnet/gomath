package gomath

import (
	"fmt"
	"math"
)

// Sphere is a ball in 3D: a centre and a radius.
//
// It is named Sphere rather than Sphere3 because there is nothing to distinguish
// it from — Circle is the 2D counterpart, so no suffix is needed. Box3, OBB3 and
// Plane3 carry theirs because AABB and Matrix already claimed the unsuffixed
// names for 2D.
//
// A sphere is the cheapest useful bounding volume: overlap is one distance
// comparison, and it is invariant under rotation, so a rotating object's bounding
// sphere never has to be recomputed. That is what makes it the right first
// rejection test in a cull or broadphase, ahead of anything box-shaped.
type Sphere[T Number] struct {
	Center Vector3[T]
	Radius T
}

// radiusSlack is the relative amount by which a radius derived from a square
// root is grown.
//
// ContainsPoint compares LengthSq against Radius*Radius, so a radius obtained as
// Sqrt(furthestSq) goes through a square root and then back through a squaring.
// Neither rounding is exact, and the round trip can land a hair BELOW the value it
// started from — which puts the very point that defined the radius outside the
// sphere. A bounding volume that excludes its own input is not a bound, and a BVH
// built on one silently misses queries.
//
// 2^-21 is about 4.8e-7, comfortably above a float32 ulp doubled by the squaring
// and utterly negligible next to the roughly 15% looseness the centroid
// construction already costs. It is a relative figure so it scales with the
// sphere.
const radiusSlack = 1.0 / (1 << 21)

// enclosingRadius returns Sqrt(radiusSq) grown by radiusSlack, so that squaring
// the result is never below radiusSq. See radiusSlack.
func enclosingRadius[T Float](radiusSq T) T {
	return Sqrt(radiusSq) * (1 + radiusSlack)
}

// NewSphere returns a sphere with the given centre and radius.
func NewSphere[T Number](center Vector3[T], radius T) Sphere[T] {
	return Sphere[T]{Center: center, Radius: radius}
}

// NewSphereFromPoints returns a sphere containing every point, and false if there
// are none.
//
// It is NOT the minimal enclosing sphere: it takes the centroid and then the
// distance to the furthest point, which is cheap, stable and can be up to about
// 15% larger than optimal. Computing the true minimum needs Welzl's algorithm,
// which is worth it for a mesh baked once and not for a bound recomputed per
// frame.
func NewSphereFromPoints[T Float](points []Vector3[T]) (Sphere[T], bool) {
	if len(points) == 0 {
		return Sphere[T]{}, false
	}

	var centroid Vector3[T]
	for _, point := range points {
		centroid = centroid.Add(point)
	}
	centroid = centroid.DivScalar(T(len(points)))

	var radiusSq T
	for _, point := range points {
		if distanceSq := point.Sub(centroid).LengthSq(); distanceSq > radiusSq {
			radiusSq = distanceSq
		}
	}

	return Sphere[T]{Center: centroid, Radius: enclosingRadius(radiusSq)}, true
}

// NewSphereFromBox3 returns the sphere circumscribing a box, containing its
// corners.
func NewSphereFromBox3[T Float](box Box3[T]) Sphere[T] {
	return Sphere[T]{Center: box.Center, Radius: enclosingRadius(box.HalfSize.LengthSq())}
}

// IsEmpty reports whether the sphere encloses no volume.
func (s Sphere[T]) IsEmpty() bool {
	return s.Radius <= 0
}

// ContainsPoint reports whether the point is inside or on the surface.
func (s Sphere[T]) ContainsPoint(point Vector3[T]) bool {
	return point.Sub(s.Center).LengthSq() <= s.Radius*s.Radius
}

// ContainsSphere reports whether other lies entirely inside s.
func (s Sphere[T]) ContainsSphere(other Sphere[T]) bool {
	if other.Radius > s.Radius {
		return false
	}

	// The far side of other must still be within s, compared without a square
	// root by keeping both sides squared.
	slack := s.Radius - other.Radius

	return other.Center.Sub(s.Center).LengthSq() <= slack*slack
}

// ClosestPoint returns the point of the sphere nearest the given point, which is
// the point itself when it is inside.
func (s Sphere[T]) ClosestPoint(point Vector3[T]) Vector3[T] {
	offset := point.Sub(s.Center)

	distanceSq := offset.LengthSq()
	if distanceSq <= s.Radius*s.Radius {
		return point
	}

	return s.Center.Add(offset.MultiplyScalar(s.Radius / Sqrt(distanceSq)))
}

// BoundingBox returns the axis-aligned box enclosing the sphere.
func (s Sphere[T]) BoundingBox() Box3[T] {
	return Box3[T]{
		Center:   s.Center,
		HalfSize: Vector3[T]{X: s.Radius, Y: s.Radius, Z: s.Radius},
	}
}

// Volume returns the enclosed volume.
//
// The arithmetic goes through float64 and converts once at the end, matching
// Circle.Area: T may be an integer type, and T(4*Pi/3) would be a conversion of
// an untyped float constant, which does not compile for an integer T.
func (s Sphere[T]) Volume() T {
	radius := float64(s.Radius)

	return T(4.0 / 3.0 * math.Pi * radius * radius * radius)
}

// SurfaceArea returns the area of the surface.
func (s Sphere[T]) SurfaceArea() T {
	radius := float64(s.Radius)

	return T(4 * math.Pi * radius * radius)
}

// Union returns the smallest sphere containing both.
func (s Sphere[T]) Union(other Sphere[T]) Sphere[T] {
	if s.ContainsSphere(other) {
		return s
	}
	if other.ContainsSphere(s) {
		return other
	}

	offset := other.Center.Sub(s.Center)
	distance := offset.Length()

	if distance == 0 {
		// Concentric, and neither contains the other only if they are equal-sized:
		// keep the centre and take the larger radius. There is no direction to
		// displace along, which is also the division by zero avoided below.
		return Sphere[T]{Center: s.Center, Radius: Max(s.Radius, other.Radius)}
	}

	// Neither contains the other, so the result spans from the far side of one to
	// the far side of the other.
	span := distance + s.Radius + other.Radius
	radius := span / 2

	// Halving is exact for a float T, but truncates for an integer one, and
	// truncating here would produce a union too small to hold its own inputs. Round
	// up instead; the branch cannot fire for a float.
	if radius*2 < span {
		radius++
	}

	// No relative slack is applied here, unlike enclosingRadius: the constant does
	// not fit in every integer type T admits. The displaced centre below is
	// therefore accurate to rounding rather than exactly containing, so a caller
	// chaining unions into a deep hierarchy should expand by an epsilon of its own.

	// Move the centre so both far sides are on the surface.
	centre := s.Center.Add(offset.MultiplyScalar((radius - s.Radius) / distance))

	return Sphere[T]{Center: centre, Radius: radius}
}

// Expanded returns the sphere with its radius grown by amount.
func (s Sphere[T]) Expanded(amount T) Sphere[T] {
	s.Radius += amount

	return s
}

// Translate returns the sphere moved by the given offset.
func (s Sphere[T]) Translate(translation Vector3[T]) Sphere[T] {
	s.Center = s.Center.Add(translation)

	return s
}

// Equals reports exact equality of centre and radius.
func (s Sphere[T]) Equals(other Sphere[T]) bool {
	return s == other
}

// ApproxEquals reports whether centre and radius are within tol.
func (s Sphere[T]) ApproxEquals(other Sphere[T], tol T) bool {
	return Abs(s.Center.X-other.Center.X) <= tol &&
		Abs(s.Center.Y-other.Center.Y) <= tol &&
		Abs(s.Center.Z-other.Center.Z) <= tol &&
		Abs(s.Radius-other.Radius) <= tol
}

func (s Sphere[T]) String() string {
	return fmt.Sprintf("Sphere(center %v, radius %v)", s.Center, s.Radius)
}
