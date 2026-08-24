package gomath

import "fmt"

// LineSegment3 is a finite piece of a line in 3D: the points between P1 and P2.
//
// It is the 3D counterpart of LineSegment, and carries the same field names so
// that code ported between the two dimensions reads the same. The suffix follows
// Box3, OBB3, Ray3 and Plane3 — LineSegment already claimed the unsuffixed name
// for 2D.
//
// Its reason to exist is Capsule3, which is exactly this segment grown by a
// radius: every capsule test reduces to a distance from this type, so the
// distance routines belong here rather than hidden inside the capsule.
//
// Helpers that need a true magnitude or a parameter along the segment (Length,
// UnitDirection, ClosestPoint, DistanceTo*, ClosestParameters and everything
// built on them) go through a division or a square root and truncate for an
// integer T, per the library-wide rule. The element-wise helpers (Midpoint,
// LengthSquared, Reversed, Translate, BoundingBox, Equals) are exact for
// integers.
type LineSegment3[T Number] struct {
	P1, P2 Vector3[T]
}

// NewLineSegment3 returns the segment between two points.
func NewLineSegment3[T Number](p1 Vector3[T], p2 Vector3[T]) LineSegment3[T] {
	return LineSegment3[T]{P1: p1, P2: p2}
}

// segmentParallelDivisor sets the relative threshold below which two segments are
// treated as parallel: their closest-approach denominator must exceed
// 1/segmentParallelDivisor of the product of their squared lengths.
//
// The denominator is |d1|^2|d2|^2 sin^2(theta) between the two directions, so
// dividing the product out makes the test a threshold on the ANGLE rather than on
// the lengths — the same comparison for a pair of millimetre segments and a pair
// of kilometre ones. An absolute epsilon would call every short pair parallel and
// every long pair skew.
//
// It is written as a divisor rather than as a 1e-6 constant because T ranges over
// the integers too, where an untyped float constant does not convert. Dividing
// the product also cannot overflow the way multiplying the denominator by a
// million could, for a quantity that is already a fourth power of a coordinate.
//
// A million is about a milliradian of misalignment, which is generous on purpose:
// this is what decides whether ClosestParameters takes its midpoint-of-the-overlap
// branch, and a pair that close to parallel has no meaningfully unique closest
// pair to lose. It is comfortably above what float32 resolves after the
// cancellation in the denominator. For an integer T the quotient floors to zero,
// which leaves the branch firing only on exactly-parallel segments — the honest
// answer for a type whose geometry is storage-only anyway.
const segmentParallelDivisor = 1000000

// parallelThreshold returns the smallest closest-approach denominator that still
// counts as skew, for two segments of the given squared lengths.
//
// The arithmetic goes through float64 and converts once. T ranges over every
// signed and floating type the library admits and the divisor does not convert to
// the narrow ones — 1000000 does not fit in an int8 — and widening keeps the
// product of two squared lengths from overflowing a narrow T on the way in.
func parallelThreshold[T Number](lengthSq1 T, lengthSq2 T) T {
	return T(float64(lengthSq1) * float64(lengthSq2) / segmentParallelDivisor)
}

// Direction returns the vector from P1 to P2. It is NOT normalised; its length
// is the segment's length, which is what the parameterised helpers divide by.
func (segment LineSegment3[T]) Direction() Vector3[T] {
	return segment.P2.Sub(segment.P1)
}

// UnitDirection returns the segment's direction as a unit vector (P1 -> P2). A
// degenerate segment has no direction and yields a zero vector.
func (segment LineSegment3[T]) UnitDirection() Vector3[T] {
	return segment.Direction().Normalize()
}

// Length returns the distance between the endpoints.
func (segment LineSegment3[T]) Length() T {
	return segment.P1.DistanceTo(segment.P2)
}

// LengthSquared returns the squared distance between the endpoints — cheaper
// than Length, and exact for integer T.
func (segment LineSegment3[T]) LengthSquared() T {
	return segment.P1.DistanceToSquared(segment.P2)
}

// IsDegenerate reports whether the segment is a single point.
//
// It is worth asking explicitly: a degenerate segment is what turns a Capsule3
// into a Sphere, and it is the case every routine below has to branch on to
// avoid dividing by a zero length.
func (segment LineSegment3[T]) IsDegenerate() bool {
	return segment.P1 == segment.P2
}

// PointAt returns the point a fraction t along the segment: t=0 is P1, t=1 is
// P2. Values outside [0,1] extrapolate along the line.
func (segment LineSegment3[T]) PointAt(t T) Vector3[T] {
	return segment.P1.Lerp(segment.P2, t)
}

// Midpoint returns the point halfway between P1 and P2.
func (segment LineSegment3[T]) Midpoint() Vector3[T] {
	return Vector3[T]{
		X: (segment.P1.X + segment.P2.X) / 2,
		Y: (segment.P1.Y + segment.P2.Y) / 2,
		Z: (segment.P1.Z + segment.P2.Z) / 2,
	}
}

// ClosestParameter returns the fraction along the segment, in [0,1], of the point
// nearest to the given point. A degenerate segment returns 0.
func (segment LineSegment3[T]) ClosestParameter(point Vector3[T]) T {
	direction := segment.Direction()

	lengthSq := direction.LengthSq()
	if lengthSq == 0 {
		return 0
	}

	return Clamp(point.Sub(segment.P1).Dot(direction)/lengthSq, 0, 1)
}

// ClosestPoint returns the point on the segment nearest to the given point.
func (segment LineSegment3[T]) ClosestPoint(point Vector3[T]) Vector3[T] {
	return segment.PointAt(segment.ClosestParameter(point))
}

// DistanceToPointSquared returns the squared distance from point to the segment.
func (segment LineSegment3[T]) DistanceToPointSquared(point Vector3[T]) T {
	return segment.ClosestPoint(point).DistanceToSquared(point)
}

// DistanceToPoint returns the distance from point to the segment.
func (segment LineSegment3[T]) DistanceToPoint(point Vector3[T]) T {
	return Sqrt(segment.DistanceToPointSquared(point))
}

// ClosestParameters returns the fractions s along this segment and t along other
// at their closest approach, both in [0,1].
//
// This is the engine behind every capsule test: two capsules overlap exactly when
// the distance between their segments is under the sum of their radii, and the
// MTV direction is the direction of this closest approach.
//
// # The parallel case is the common one, not the exotic one
//
// When the segments are parallel and their projections overlap there is no unique
// closest pair — every point of the overlap is equally close. Two upright
// character capsules standing side by side are exactly that, so this is the
// arrangement a crowd produces constantly.
//
// The textbook formulation divides by a*e - b*b, which is zero there, and the
// usual repair is to clamp s to 0 and accept whatever endpoint falls out. That
// makes the contact point jump to an end cap, and for two capsules of equal
// height standing level it puts the contact at their feet. This returns the
// MIDPOINT of the overlapping interval instead, which is stable under small
// perturbations and is the point a solver wants to apply the impulse at.
// collider/obb3.go records the same lesson for parallel boxes.
func (segment LineSegment3[T]) ClosestParameters(other LineSegment3[T]) (s T, t T) {
	direction1 := segment.Direction()
	direction2 := other.Direction()
	offset := segment.P1.Sub(other.P1)

	lengthSq1 := direction1.LengthSq()
	lengthSq2 := direction2.LengthSq()
	projected2 := direction2.Dot(offset)

	// Both degenerate: two points, and the only parameters that exist are 0.
	if lengthSq1 == 0 && lengthSq2 == 0 {
		return 0, 0
	}

	// This one is a point: project it onto the other.
	if lengthSq1 == 0 {
		return 0, Clamp(projected2/lengthSq2, 0, 1)
	}

	projected1 := direction1.Dot(offset)

	// The other is a point: project it onto this one.
	if lengthSq2 == 0 {
		return Clamp(-projected1/lengthSq1, 0, 1), 0
	}

	alignment := direction1.Dot(direction2)
	denominator := lengthSq1*lengthSq2 - alignment*alignment

	if denominator > parallelThreshold(lengthSq1, lengthSq2) {
		s = Clamp((alignment*projected2-projected1*lengthSq2)/denominator, 0, 1)
	} else {
		s = parallelOverlapMidpoint(-projected1/lengthSq1, (alignment-projected1)/lengthSq1)
	}

	// Where s lands on the other segment, and the correction if that overshoots
	// its ends — moving t back changes which point of this segment is nearest.
	t = (alignment*s + projected2) / lengthSq2
	if t < 0 {
		return Clamp(-projected1/lengthSq1, 0, 1), 0
	}
	if t > 1 {
		return Clamp((alignment-projected1)/lengthSq1, 0, 1), 1
	}

	return s, t
}

// parallelOverlapMidpoint returns the parameter to use for a pair of parallel
// segments, given where the other segment's endpoints project onto this one.
//
// Where the projected span meets [0,1] the answer is the middle of the shared
// part. Where it does not, the segments lie end to end and the nearest point is
// simply the end they face each other across, which clamping the projected
// midpoint picks out.
func parallelOverlapMidpoint[T Number](projectedStart T, projectedEnd T) T {
	low := Min(projectedStart, projectedEnd)
	high := Max(projectedStart, projectedEnd)

	if overlapLow, overlapHigh := Max(low, 0), Min(high, 1); overlapLow <= overlapHigh {
		return (overlapLow + overlapHigh) / 2
	}

	return Clamp((low+high)/2, 0, 1)
}

// ClosestPoints returns the pair of points, one on each segment, at their closest
// approach.
func (segment LineSegment3[T]) ClosestPoints(other LineSegment3[T]) (onSegment Vector3[T], onOther Vector3[T]) {
	s, t := segment.ClosestParameters(other)

	return segment.PointAt(s), other.PointAt(t)
}

// DistanceToSegmentSquared returns the squared distance between the two segments.
func (segment LineSegment3[T]) DistanceToSegmentSquared(other LineSegment3[T]) T {
	onSegment, onOther := segment.ClosestPoints(other)

	return onSegment.DistanceToSquared(onOther)
}

// DistanceToSegment returns the distance between the two segments.
func (segment LineSegment3[T]) DistanceToSegment(other LineSegment3[T]) T {
	return Sqrt(segment.DistanceToSegmentSquared(other))
}

// Reversed returns the segment with its endpoints swapped.
func (segment LineSegment3[T]) Reversed() LineSegment3[T] {
	return LineSegment3[T]{P1: segment.P2, P2: segment.P1}
}

// Translate returns a copy of the segment moved by delta.
func (segment LineSegment3[T]) Translate(delta Vector3[T]) LineSegment3[T] {
	return LineSegment3[T]{P1: segment.P1.Add(delta), P2: segment.P2.Add(delta)}
}

// BoundingBox returns the axis-aligned box that tightly encloses the segment.
func (segment LineSegment3[T]) BoundingBox() Box3[T] {
	return NewBox3FromMinMax(segment.P1, segment.P2)
}

// Equals reports whether both segments have the same endpoints in the same order.
func (segment LineSegment3[T]) Equals(other LineSegment3[T]) bool {
	return segment == other
}

func (segment LineSegment3[T]) String() string {
	return fmt.Sprintf("%v -> %v", segment.P1, segment.P2)
}
