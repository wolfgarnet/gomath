package gomath

import (
	"fmt"
	"math"
)

// Capsule3 is a segment grown by a radius: every point within Radius of the
// segment. It is the shape a walking character wants.
//
// # Why a capsule and not a sphere or a box
//
// A sphere wide enough to hold a character's shoulders is that wide at the feet
// too, so the character can never get near a wall. A capsule is the right width
// at every height.
//
// A box has corners, and corners catch. Walking a box along a wall snags on every
// seam in the geometry. A capsule's surface is smooth everywhere, so it slides
// over small steps and along walls rather than stopping dead at them.
//
// It is also the cheapest of the three to test. OBB3OBB3 examines fifteen
// separating axes; capsule-against-anything reduces to a distance from the
// segment, which is a handful of dot products. The most common entity in a game
// should not be paying for the most expensive test.
//
// # No orientation field
//
// The segment is stored in world space, so the capsule's orientation is implicit
// in its two endpoints and there is nothing to keep in sync. A character capsule
// is upright anyway; NewCapsule3Upright is the constructor for that case.
//
// # The reduction that keeps the tests honest
//
// A capsule whose segment is degenerate is exactly a Sphere of the same radius,
// so every capsule test must agree with its sphere counterpart there. That is a
// free oracle for the whole feature, and it is why the type stores a segment
// rather than a centre, a half-height and an axis — the latter has to special-case
// a zero axis where this one simply collapses.
//
// Overlap between a capsule and any other shape lives in the collider package;
// the root types carry only self-contained helpers.
type Capsule3[T Number] struct {
	Segment LineSegment3[T]
	Radius  T
}

// NewCapsule3 returns the capsule of the given radius about the segment between
// two points.
func NewCapsule3[T Number](start Vector3[T], end Vector3[T], radius T) Capsule3[T] {
	return Capsule3[T]{Segment: NewLineSegment3(start, end), Radius: radius}
}

// NewCapsule3Upright returns an upright capsule standing on base, of the given
// TOTAL height and radius. Y is up.
//
// height is the full extent from the soles to the crown, caps included — the
// figure a designer means by "this character is 1.8 metres tall" — so the segment
// runs from base+radius to base+height-radius and the capsule's bounding box is
// exactly height tall.
//
// A height below twice the radius cannot be built: the caps alone are already
// that tall. Rather than return an error for a case a caller can neither
// anticipate nor usefully handle mid-frame, the segment collapses to a point at
// mid-height and the capsule becomes a sphere of the given radius. That is the
// nearest shape that exists, it is never smaller than requested, and it keeps
// every test below well defined.
func NewCapsule3Upright[T Number](base Vector3[T], height T, radius T) Capsule3[T] {
	low := base.Y + radius
	high := base.Y + height - radius

	if high < low {
		middle := base.Y + height/2
		low, high = middle, middle
	}

	return Capsule3[T]{
		Segment: LineSegment3[T]{
			P1: Vector3[T]{X: base.X, Y: low, Z: base.Z},
			P2: Vector3[T]{X: base.X, Y: high, Z: base.Z},
		},
		Radius: radius,
	}
}

// NewCapsule3FromSphere returns the capsule identical to a sphere: the same
// radius about a degenerate segment at its centre.
//
// It exists to make the reduction in the type's doc comment expressible in code,
// which is what the tests compare against.
func NewCapsule3FromSphere[T Number](sphere Sphere[T]) Capsule3[T] {
	return Capsule3[T]{
		Segment: LineSegment3[T]{P1: sphere.Center, P2: sphere.Center},
		Radius:  sphere.Radius,
	}
}

// IsEmpty reports whether the capsule encloses no volume.
func (capsule Capsule3[T]) IsEmpty() bool {
	return capsule.Radius <= 0
}

// IsSphere reports whether the segment is degenerate, so the capsule is exactly a
// sphere. See Sphere.
func (capsule Capsule3[T]) IsSphere() bool {
	return capsule.Segment.IsDegenerate()
}

// Sphere returns the capsule as a Sphere, and false unless its segment is
// degenerate.
func (capsule Capsule3[T]) Sphere() (Sphere[T], bool) {
	if !capsule.IsSphere() {
		return Sphere[T]{}, false
	}

	return Sphere[T]{Center: capsule.Segment.P1, Radius: capsule.Radius}, true
}

// Center returns the midpoint of the segment, which is the capsule's centre of
// symmetry.
func (capsule Capsule3[T]) Center() Vector3[T] {
	return capsule.Segment.Midpoint()
}

// Height returns the capsule's total extent along its own axis: the segment plus
// a radius at each end.
//
// For a capsule built by NewCapsule3Upright this returns the height it was given.
func (capsule Capsule3[T]) Height() T {
	return capsule.Segment.Length() + 2*capsule.Radius
}

// ContainsPoint reports whether the point is inside or on the surface.
func (capsule Capsule3[T]) ContainsPoint(point Vector3[T]) bool {
	return capsule.Segment.DistanceToPointSquared(point) <= capsule.Radius*capsule.Radius
}

// ClosestPoint returns the point of the capsule nearest the given point, which is
// the point itself when it is inside.
func (capsule Capsule3[T]) ClosestPoint(point Vector3[T]) Vector3[T] {
	onAxis := capsule.Segment.ClosestPoint(point)
	offset := point.Sub(onAxis)

	distanceSq := offset.LengthSq()
	if distanceSq <= capsule.Radius*capsule.Radius {
		return point
	}

	return onAxis.Add(offset.MultiplyScalar(capsule.Radius / Sqrt(distanceSq)))
}

// BoundingBox returns the axis-aligned box enclosing the capsule.
//
// It is the segment's bounds grown by the radius on every axis, which is exact:
// the extreme point on each axis lies on one of the two caps.
func (capsule Capsule3[T]) BoundingBox() Box3[T] {
	radius := Vector3[T]{X: capsule.Radius, Y: capsule.Radius, Z: capsule.Radius}

	return NewBox3FromMinMax(
		capsule.Segment.P1.Min(capsule.Segment.P2).Sub(radius),
		capsule.Segment.P1.Max(capsule.Segment.P2).Add(radius),
	)
}

// Volume returns the enclosed volume: a cylinder along the segment, plus the two
// caps, which together are one whole sphere.
//
// The arithmetic goes through float64 and converts once at the end, matching
// Sphere.Volume: T may be an integer type, for which an untyped float constant
// does not convert.
func (capsule Capsule3[T]) Volume() T {
	radius := float64(capsule.Radius)
	length := float64(capsule.Segment.Length())

	return T(math.Pi*radius*radius*length + 4.0/3.0*math.Pi*radius*radius*radius)
}

// SurfaceArea returns the area of the surface: the cylinder's side, plus the two
// caps, which together are one whole sphere.
func (capsule Capsule3[T]) SurfaceArea() T {
	radius := float64(capsule.Radius)
	length := float64(capsule.Segment.Length())

	return T(2*math.Pi*radius*length + 4*math.Pi*radius*radius)
}

// Expanded returns the capsule with its radius grown by amount, leaving the
// segment where it is.
//
// This is the Minkowski sum with a ball, which is what a swept or padded query
// wants — note it grows the capsule at both ends as well as at the sides, so the
// total Height grows by twice amount.
func (capsule Capsule3[T]) Expanded(amount T) Capsule3[T] {
	capsule.Radius += amount

	return capsule
}

// Translate returns the capsule moved by the given offset.
func (capsule Capsule3[T]) Translate(translation Vector3[T]) Capsule3[T] {
	capsule.Segment = capsule.Segment.Translate(translation)

	return capsule
}

// Equals reports exact equality of segment and radius. Note that a capsule and
// its reverse are the same set of points but are NOT equal by this test.
func (capsule Capsule3[T]) Equals(other Capsule3[T]) bool {
	return capsule == other
}

// ApproxEquals reports whether both endpoints and the radius are within tol.
func (capsule Capsule3[T]) ApproxEquals(other Capsule3[T], tol T) bool {
	return approxEqualsVector3(capsule.Segment.P1, other.Segment.P1, tol) &&
		approxEqualsVector3(capsule.Segment.P2, other.Segment.P2, tol) &&
		Abs(capsule.Radius-other.Radius) <= tol
}

func approxEqualsVector3[T Number](a Vector3[T], b Vector3[T], tol T) bool {
	return Abs(a.X-b.X) <= tol && Abs(a.Y-b.Y) <= tol && Abs(a.Z-b.Z) <= tol
}

func (capsule Capsule3[T]) String() string {
	return fmt.Sprintf("Capsule3(%v, radius %v)", capsule.Segment, capsule.Radius)
}
