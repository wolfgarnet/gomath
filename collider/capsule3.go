package collider

import (
	"slices"

	"github.com/wolfgarnet/gomath"
)

// Capsule overlap in 3D. Normals point from the FIRST shape towards the second,
// as everywhere else here.
//
// A capsule is a segment grown by a radius, so every test below is a distance
// from that segment compared against a sum of radii — which is why the capsule is
// the cheap shape to give a character. See gomath.Capsule3 for why it is also the
// right one.
//
// # Everything here is exact
//
// Nothing in this file is an approximation, including the box pairs, where one
// would have been tempting. The section above capsuleAlignedBoxPenetration says
// how that is achieved and why the obvious shortcut — running the sphere-box test
// at the segment's closest point to the box centre — is wrong exactly where a
// character meets a corner.
//
// # Why these are Float where the sphere tests are Number
//
// A capsule test has to find a PARAMETER along its segment, which is a division,
// where SphereBox3 only ever clamps a coordinate. Truncating that parameter to an
// integer snaps the closest point to an endpoint and silently answers a different
// question, so an integer T is not offered rather than being offered wrong.

// crossAxisDegeneracy is the relative squared length below which an edge-cross
// axis is discarded.
//
// The axes concerned are boxAxis x segmentDirection, whose squared length is
// |segmentDirection|^2 sin^2(theta) — so dividing by the direction's squared
// length makes this a threshold on the ANGLE, independent of how long the segment
// is. At 1e-10 that is about 1e-5 radians.
//
// Such an axis carries no direction worth normalising: its components are pure
// cancellation error. Discarding it is safe because a degenerate axis cannot be a
// separating axis, and when the segment is degenerate all three vanish at once —
// which is exactly how the capsule tests collapse to the sphere ones.
const crossAxisDegeneracy = 1e-10

// Capsule3Sphere reports whether a capsule overlaps a sphere.
//
// The whole test: the sphere's centre is within radius1+radius2 of the capsule's
// segment. It is also the independent one-line oracle the capsule tests are
// checked against.
func Capsule3Sphere[T gomath.Float](capsule gomath.Capsule3[T], sphere gomath.Sphere[T]) bool {
	reach := capsule.Radius + sphere.Radius

	return capsule.Segment.DistanceToPointSquared(sphere.Center) <= reach*reach
}

// Capsule3SpherePenetration returns the minimum translation separating a capsule
// from a sphere, with the normal pointing from the capsule towards the sphere.
//
// The direction is the one of closest approach between the capsule's SEGMENT and
// the sphere's centre — not between the two centres. A centre-to-centre normal
// looks plausible and is wrong whenever the contact is near an end cap, which is
// exactly where a character touches the ground.
func Capsule3SpherePenetration[T gomath.Float](capsule gomath.Capsule3[T], sphere gomath.Sphere[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	onAxis := capsule.Segment.ClosestPoint(sphere.Center)
	offset := sphere.Center.Sub(onAxis)
	reach := capsule.Radius + sphere.Radius

	distanceSq := offset.LengthSq()
	if distanceSq > reach*reach {
		return 0, normal, false
	}

	if distanceSq > 0 {
		distance := gomath.Sqrt(distanceSq)

		return reach - distance, offset.DivScalar(distance), true
	}

	// The centre lies ON the capsule's axis, so there is no closest-approach
	// direction to read off. Every direction perpendicular to the axis escapes by
	// the same amount, and that amount is the full reach.
	return reach, anyPerpendicular(capsule.Segment.Direction()), true
}

// SphereCapsule3Penetration returns the minimum translation separating a sphere
// from a capsule, with the normal pointing from the sphere towards the capsule.
// It is Capsule3SpherePenetration with the arguments the other way round.
func SphereCapsule3Penetration[T gomath.Float](sphere gomath.Sphere[T], capsule gomath.Capsule3[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	penetration, normal, overlapping = Capsule3SpherePenetration(capsule, sphere)
	if !overlapping {
		return 0, gomath.Vector3[T]{}, false
	}

	return penetration, normal.Negate(), true
}

// Capsule3Capsule3 reports whether two capsules overlap: their segments are
// within the sum of their radii.
func Capsule3Capsule3[T gomath.Float](a gomath.Capsule3[T], b gomath.Capsule3[T]) bool {
	reach := a.Radius + b.Radius

	return a.Segment.DistanceToSegmentSquared(b.Segment) <= reach*reach
}

// Capsule3Capsule3Penetration returns the minimum translation separating two
// overlapping capsules, with the normal pointing from a towards b.
//
// Two capsules standing side by side are a pair of parallel segments, which have
// no unique closest pair — the common arrangement for a crowd, not an exotic one.
// gomath.LineSegment3.ClosestParameters resolves it to the middle of the
// overlapping interval; see its comment for why that particular choice.
func Capsule3Capsule3Penetration[T gomath.Float](a gomath.Capsule3[T], b gomath.Capsule3[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	onA, onB := a.Segment.ClosestPoints(b.Segment)
	offset := onB.Sub(onA)
	reach := a.Radius + b.Radius

	distanceSq := offset.LengthSq()
	if distanceSq > reach*reach {
		return 0, normal, false
	}

	if distanceSq > 0 {
		distance := gomath.Sqrt(distanceSq)

		return reach - distance, offset.DivScalar(distance), true
	}

	// The segments touch, so the axes themselves give the direction. Two crossing
	// segments span a plane and the shortest way out is across it; parallel
	// segments that touch are collinear, and then any direction perpendicular to
	// the shared axis will do.
	//
	// Fully coincident axes have no preferred direction at all, and the answer
	// here is arbitrary — as it is for concentric spheres in
	// SphereSpherePenetration. In that one case the swapped call does NOT return
	// the negated normal, because there is nothing in the configuration to
	// distinguish the two orders by.
	return reach, separatingAxisOf(a.Segment.Direction(), b.Segment.Direction()), true
}

// Capsule3Box3 reports whether a capsule overlaps an axis-aligned box.
func Capsule3Box3[T gomath.Float](capsule gomath.Capsule3[T], box gomath.Box3[T]) bool {
	_, distanceSq := closestOnSegmentToAlignedBox(capsule.Segment.Translate(box.Center.Negate()), box.HalfSize)

	return distanceSq <= capsule.Radius*capsule.Radius
}

// Capsule3OBB3 reports whether a capsule overlaps an oriented box.
func Capsule3OBB3[T gomath.Float](capsule gomath.Capsule3[T], box gomath.OBB3[T]) bool {
	_, distanceSq := closestOnSegmentToAlignedBox(segmentToLocal(capsule.Segment, box), box.HalfSize)

	return distanceSq <= capsule.Radius*capsule.Radius
}

// Capsule3Box3Penetration returns the minimum translation separating a capsule
// from an axis-aligned box, with the normal pointing from the capsule towards the
// box.
func Capsule3Box3Penetration[T gomath.Float](capsule gomath.Capsule3[T], box gomath.Box3[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	return capsuleAlignedBoxPenetration(capsule.Segment.Translate(box.Center.Negate()), capsule.Radius, box.HalfSize)
}

// Box3Capsule3Penetration returns the minimum translation separating an
// axis-aligned box from a capsule, with the normal pointing from the box towards
// the capsule. It is Capsule3Box3Penetration with the arguments the other way
// round.
func Box3Capsule3Penetration[T gomath.Float](box gomath.Box3[T], capsule gomath.Capsule3[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	penetration, normal, overlapping = Capsule3Box3Penetration(capsule, box)
	if !overlapping {
		return 0, gomath.Vector3[T]{}, false
	}

	return penetration, normal.Negate(), true
}

// Capsule3OBB3Penetration returns the minimum translation separating a capsule
// from an oriented box, with the normal pointing from the capsule towards the box.
//
// As in SphereOBB3Penetration, the whole problem is the axis-aligned one once the
// capsule's segment is expressed in the box's frame. Only the direction has to
// come back out to world space; the depth is a length, and a rotation does not
// change lengths.
func Capsule3OBB3Penetration[T gomath.Float](capsule gomath.Capsule3[T], box gomath.OBB3[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	penetration, local, overlapping := capsuleAlignedBoxPenetration(segmentToLocal(capsule.Segment, box), capsule.Radius, box.HalfSize)
	if !overlapping {
		return 0, gomath.Vector3[T]{}, false
	}

	return penetration, localDirectionToWorld(box, local), true
}

// OBB3Capsule3Penetration returns the minimum translation separating an oriented
// box from a capsule, with the normal pointing from the box towards the capsule.
// It is Capsule3OBB3Penetration with the arguments the other way round.
func OBB3Capsule3Penetration[T gomath.Float](box gomath.OBB3[T], capsule gomath.Capsule3[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	penetration, normal, overlapping = Capsule3OBB3Penetration(capsule, box)
	if !overlapping {
		return 0, gomath.Vector3[T]{}, false
	}

	return penetration, normal.Negate(), true
}

// segmentToLocal expresses a segment in an oriented box's frame, where the box is
// the axis-aligned span [-HalfSize, +HalfSize].
func segmentToLocal[T gomath.Float](segment gomath.LineSegment3[T], box gomath.OBB3[T]) gomath.LineSegment3[T] {
	return gomath.LineSegment3[T]{P1: box.ToLocal(segment.P1), P2: box.ToLocal(segment.P2)}
}

// capsuleAlignedBoxPenetration solves the capsule-against-box problem in the
// box's own frame, where the box is the axis-aligned span [-halfSize, +halfSize].
// Both box pairs route through it; only the mapping in and out differs.
//
// # Why this is exact, when a capsule-box test usually is not
//
// The tempting shortcut is to find the point on the segment nearest the box's
// centre and run the sphere-box test there. It is right for a face contact and
// wrong for a corner: the closest point to the CENTRE is not the closest point to
// the box, so a character standing on the corner of a step gets a normal that
// points somewhere other than out of the corner, and sinks into it.
//
// The exact answer needs two branches, and each is closed-form.
//
// Where the segment misses the box, closestOnSegmentToAlignedBox gives the true
// closest pair, and the minimum translation is along it with depth radius minus
// distance. That is not obvious, so: write D for the Minkowski difference
// box - segment, a convex set, and c for its point nearest the origin, at
// distance d. Moving the capsule by lambda along -c/d puts the origin on the far
// side of D's supporting hyperplane at c, so the separation is at least d+lambda,
// and lambda = radius-d therefore clears it. No shorter move can, because any
// translation t reaches a separation of at most d+|t|. So radius-d is the depth
// and c/d the direction, whichever features the closest pair happens to lie on.
//
// Where the segment PIERCES the box, the origin is inside D and the depth is its
// distance to D's boundary — which for a convex polytope is the least distance to
// a face plane, so a separating-axis minimum over D's face normals is exact.
// D is a box swept along a segment, and the face normals of a Minkowski sum are
// the face normals of each operand plus their edge-edge cross products; a segment
// contributes no faces and one edge, leaving the box's three axes and the three
// crosses with the segment's direction. Six axes, and the rounding by radius adds
// the same amount to every one of them, so it does not change which is least.
//
// A degenerate segment kills all three cross axes at once, leaving the three face
// normals — which is exactly shallowestFaceNormal, so the capsule test collapses
// onto the sphere one rather than merely agreeing with it.
//
// # The one place accuracy is soft
//
// A segment passing a corner at a gap of a few ulps takes the first branch, where
// the difference between the closest points is a subtraction of nearly equal
// coordinates. Its LENGTH survives that, so the depth stays accurate — but its
// direction does not, and the returned normal can be several degrees out at a
// float32 gap below about a millionth of a unit. No threshold is applied to hide
// it: a pair that close is touching, the depth is right either way, and rounding
// the gap up to contact would trade a bounded direction error for an unbounded
// choice of which branch answers. Contact proper is exact, which is what
// closestOnSegmentToAlignedBox's middle-of-the-interval return buys.
func capsuleAlignedBoxPenetration[T gomath.Float](segment gomath.LineSegment3[T], radius T, halfSize gomath.Vector3[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	parameter, distanceSq := closestOnSegmentToAlignedBox(segment, halfSize)
	if distanceSq > radius*radius {
		return 0, normal, false
	}

	if distanceSq > 0 {
		onSegment := segment.PointAt(parameter)
		onBox := onSegment.Clamp(halfSize.Negate(), halfSize)
		distance := gomath.Sqrt(distanceSq)

		return radius - distance, onBox.Sub(onSegment).DivScalar(distance), true
	}

	return piercedAlignedBoxPenetration(segment, radius, halfSize)
}

// piercedAlignedBoxPenetration is the second branch above: the segment meets the
// box, so the minimum is found by a separating-axis sweep rather than from a
// closest pair that has collapsed to a point.
//
// Unlike OBB3OBB3Penetration this applies no bias towards face axes. That bias
// exists to stop a resting contact's normal flipping between frames, and nothing
// rests in this state — the segment is through the box, which is a deep overlap a
// solver is digging out of, not a contact it is holding.
func piercedAlignedBoxPenetration[T gomath.Float](segment gomath.LineSegment3[T], radius T, halfSize gomath.Vector3[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	direction := segment.Direction()

	axes := [6]gomath.Vector3[T]{
		{X: 1},
		{Y: 1},
		{Z: 1},
		// The three box axes crossed with the segment's direction.
		{Y: -direction.Z, Z: direction.Y},
		{X: direction.Z, Z: -direction.X},
		{X: -direction.Y, Y: direction.X},
	}

	directionLengthSq := direction.LengthSq()
	found := false

	for index, axis := range axes {
		if index >= 3 && axis.LengthSq() <= directionLengthSq*T(crossAxisDegeneracy) {
			continue
		}

		unit := axis.Normalize()

		// The capsule's extent along the axis, and the box's, which is centred on
		// the origin in this frame.
		low := gomath.Min(segment.P1.Dot(unit), segment.P2.Dot(unit))
		high := gomath.Max(segment.P1.Dot(unit), segment.P2.Dot(unit))
		reach := halfSize.X*gomath.Abs(unit.X) + halfSize.Y*gomath.Abs(unit.Y) + halfSize.Z*gomath.Abs(unit.Z)

		// How far the box would have to travel along the axis to clear the
		// capsule, in each of the two directions along it. Written from the
		// support functions rather than as an interval overlap, which is a
		// different number as soon as one shape contains the other.
		depth, candidate := high+radius+reach, unit
		if opposite := radius + reach - low; opposite < depth {
			depth, candidate = opposite, unit.Negate()
		}

		if !found || depth < penetration {
			penetration, normal, found = depth, candidate, true
		}
	}

	return penetration, normal, true
}

// closestOnSegmentToAlignedBox returns the parameter along the segment of the
// point nearest a box spanning [-halfSize, +halfSize], and the squared distance
// there. A segment meeting the box gives zero.
//
// # Why this is exact and cheap at the same time
//
// Distance to a convex set is a convex function, and the segment is affine in its
// parameter, so the squared distance is a CONVEX function of t on [0,1]. It is
// also piecewise quadratic: within any range of t over which each coordinate stays
// on the same side of its slab, the difference from the clamped point is affine in
// t, so the sum of squares is one ordinary quadratic to be minimised in closed
// form.
//
// The pieces are delimited by the parameters at which a coordinate crosses a slab
// face — six of them at most, two per axis. So the whole minimisation is at most
// eight breakpoints and seven quadratics, with no iteration and no tolerance to
// pick. Every interval is examined rather than the convexity being used to
// descend, which costs nothing at this size and leaves the result independent of
// the rounding in the breakpoints.
func closestOnSegmentToAlignedBox[T gomath.Float](segment gomath.LineSegment3[T], halfSize gomath.Vector3[T]) (parameter T, distanceSq T) {
	origin := [3]T{segment.P1.X, segment.P1.Y, segment.P1.Z}
	direction := segment.Direction()
	slope := [3]T{direction.X, direction.Y, direction.Z}
	half := [3]T{halfSize.X, halfSize.Y, halfSize.Z}

	// The endpoints, plus every slab crossing strictly between them.
	breakpoints := [8]T{0, 1}
	count := 2

	for axis := range 3 {
		if slope[axis] == 0 {
			continue
		}

		inverse := 1 / slope[axis]
		for _, face := range [2]T{-half[axis], half[axis]} {
			if crossing := (face - origin[axis]) * inverse; crossing > 0 && crossing < 1 {
				breakpoints[count] = crossing
				count++
			}
		}
	}

	slices.Sort(breakpoints[:count])

	parameter = 0
	distanceSq = gapToAlignedBoxSquared(origin, slope, half, 0)

	for index := 0; index+1 < count; index++ {
		low, high := breakpoints[index], breakpoints[index+1]
		if !(low < high) {
			continue
		}

		candidate := minimiseGapOnInterval(origin, slope, half, low, high)

		if gap := gapToAlignedBoxSquared(origin, slope, half, candidate); gap < distanceSq {
			parameter, distanceSq = candidate, gap
		}
	}

	return parameter, distanceSq
}

// minimiseGapOnInterval returns the parameter minimising the squared gap over
// [low, high], on which the clamping pattern is assumed constant.
//
// The pattern is read once at the interval's midpoint, which is what makes the
// gap a single quadratic here: an axis outside its slab contributes an affine
// residual, and one inside contributes nothing at all.
func minimiseGapOnInterval[T gomath.Float](origin [3]T, slope [3]T, half [3]T, low T, high T) T {
	var quadratic, linear T

	middle := (low + high) / 2

	for axis := range 3 {
		position := origin[axis] + middle*slope[axis]

		var constant T
		switch {
		case position < -half[axis]:
			constant = -half[axis] - origin[axis]
		case position > half[axis]:
			constant = half[axis] - origin[axis]
		default:
			continue
		}

		quadratic += slope[axis] * slope[axis]
		linear += constant * slope[axis]
	}

	if quadratic == 0 {
		// The gap does not vary over this interval, so any parameter in it is a
		// minimum: the segment is inside the box here, or degenerate, or running
		// parallel to every face it is outside of.
		//
		// The MIDDLE is returned rather than an end, and the difference is not
		// cosmetic. An end of the interval is a slab crossing, computed by a
		// division, so a segment that passes through the box reports a gap there of
		// a few times nothing rather than of nothing — and a few times nothing is
		// still greater than zero, which sends the caller down its
		// closest-pair branch to normalise a difference vector made entirely of
		// rounding error. A point in the interior evaluates to an exact zero.
		return middle
	}

	// The residual is constant - slope*t on each contributing axis, so the sum of
	// squares is least where its derivative vanishes.
	return gomath.Clamp(linear/quadratic, low, high)
}

// gapToAlignedBoxSquared returns the squared distance from the point at parameter
// to the box spanning [-half, +half], which is zero inside it.
func gapToAlignedBoxSquared[T gomath.Float](origin [3]T, slope [3]T, half [3]T, parameter T) T {
	var total T

	for axis := range 3 {
		position := origin[axis] + parameter*slope[axis]

		switch {
		case position < -half[axis]:
			total += gomath.Sqr(position + half[axis])
		case position > half[axis]:
			total += gomath.Sqr(position - half[axis])
		}
	}

	return total
}

// separatingAxisOf returns a unit direction perpendicular to both segment
// directions, for the degenerate case where two capsules' axes touch.
//
// Crossing axes span a plane and their common perpendicular is the way out of it.
// When the cross product vanishes the axes are parallel and touching, so they are
// collinear, and any direction perpendicular to the shared one escapes equally.
func separatingAxisOf[T gomath.Float](direction1 gomath.Vector3[T], direction2 gomath.Vector3[T]) gomath.Vector3[T] {
	cross := direction1.Cross(direction2)

	// Relative, so the test is on the angle between the axes rather than on their
	// lengths. See crossAxisDegeneracy.
	if cross.LengthSq() > direction1.LengthSq()*direction2.LengthSq()*T(crossAxisDegeneracy) {
		return cross.Normalize()
	}

	if !direction1.IsZero() {
		return anyPerpendicular(direction1)
	}

	return anyPerpendicular(direction2)
}

// anyPerpendicular returns some unit vector at right angles to the given one, and
// the +X axis for a zero vector — matching the arbitrary choice
// SphereSpherePenetration makes for concentric spheres.
//
// Which perpendicular is returned does not matter; that it is well away from
// degenerate does. Crossing with the axis the vector leans on LEAST is what
// guarantees that: the cross product is then at its longest, so normalising it
// costs no precision.
func anyPerpendicular[T gomath.Float](direction gomath.Vector3[T]) gomath.Vector3[T] {
	if direction.IsZero() {
		return gomath.Vector3[T]{X: 1}
	}

	leaning := direction.Abs()

	axis := gomath.Vector3[T]{X: 1}
	if leaning.Y < leaning.X && leaning.Y <= leaning.Z {
		axis = gomath.Vector3[T]{Y: 1}
	} else if leaning.Z < leaning.X && leaning.Z < leaning.Y {
		axis = gomath.Vector3[T]{Z: 1}
	}

	return direction.Cross(axis).Normalize()
}
