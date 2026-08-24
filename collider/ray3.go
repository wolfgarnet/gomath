package collider

import (
	"math"

	"github.com/wolfgarnet/gomath"
)

// Ray casting in 3D.
//
// Every test here returns the distance along the ray to the FIRST surface it
// meets, plus whether it met one at all. Distances are true world distances
// because Ray3 keeps a unit direction, so results from different shapes compare
// directly — which is the whole mechanism behind "what did I click on": cast at
// everything, keep the smallest.
//
// # Rays start where they start
//
// A hit behind the origin is not a hit. Getting this wrong is not a subtle
// numerical matter, it is the difference between picking the object under the
// cursor and picking the one behind the camera.
//
// # A ray that begins inside a shape
//
// Every test reports a hit at distance 0 rather than at the exit point. The
// alternative — reporting where the ray leaves — means a camera inside an
// object's bounds picks its far wall, and a query about whether the origin is
// enclosed cannot be answered from the result at all.

// Ray3Plane3 returns the distance to where a ray meets an infinite plane.
//
// This is the workhorse of ground picking: cast the cursor's ray at the y=0
// plane and the hit point is where the player is pointing on the map.
//
// A ray parallel to the plane never meets it, and a ray pointing away from it
// meets it only behind the origin — both report no hit. A ray whose ORIGIN lies
// on the plane hits at distance 0, whichever way it points, since the origin is
// genuinely a point of the plane.
func Ray3Plane3[T gomath.Float](ray gomath.Ray3[T], plane gomath.Plane3[T]) (distance T, hit bool) {
	if !ray.IsValid() {
		return 0, false
	}

	alignment := plane.Normal.Dot(ray.Direction)
	fromPlane := plane.SignedDistance(ray.Origin)

	// Starting on the plane counts, and is checked before the parallel test so
	// that a ray lying IN the plane still reports its origin rather than
	// nothing.
	if fromPlane == 0 {
		return 0, true
	}

	if alignment == 0 {
		// Parallel and not on it: no meeting point exists.
		return 0, false
	}

	distance = -fromPlane / alignment
	if distance < 0 {
		return 0, false
	}

	return distance, true
}

// Ray3Sphere returns the distance to where a ray first meets a sphere.
//
// A ray starting inside reports 0, per the note above.
func Ray3Sphere[T gomath.Float](ray gomath.Ray3[T], sphere gomath.Sphere[T]) (distance T, hit bool) {
	if !ray.IsValid() {
		return 0, false
	}

	toCentre := sphere.Center.Sub(ray.Origin)

	// Inside, including exactly on the surface.
	if toCentre.LengthSq() <= sphere.Radius*sphere.Radius {
		return 0, true
	}

	// How far along the ray the sphere's centre projects. Negative means the
	// centre is behind, and since the origin is outside, the sphere is too.
	along := toCentre.Dot(ray.Direction)
	if along < 0 {
		return 0, false
	}

	// The squared distance from the centre to the ray at its closest approach,
	// by Pythagoras — no square root needed to reject a miss.
	closestSq := toCentre.LengthSq() - along*along
	radiusSq := sphere.Radius * sphere.Radius
	if closestSq > radiusSq {
		return 0, false
	}

	// Back up from the closest approach to where the surface is crossed.
	halfChord := gomath.Sqrt(radiusSq - closestSq)

	return along - halfChord, true
}

// Ray3Box3 returns the distance to where a ray first meets an axis-aligned box.
//
// The slab method: the box is the intersection of three pairs of parallel
// planes, and the ray is inside the box exactly where it is inside all three
// slabs at once. So the entry is the LATEST of the three slab entries and the
// exit the EARLIEST of the exits, and they cross over when the ray misses.
func Ray3Box3[T gomath.Float](ray gomath.Ray3[T], box gomath.Box3[T]) (distance T, hit bool) {
	if !ray.IsValid() {
		return 0, false
	}

	minimum, maximum := box.Min(), box.Max()

	origin := [3]T{ray.Origin.X, ray.Origin.Y, ray.Origin.Z}
	direction := [3]T{ray.Direction.X, ray.Direction.Y, ray.Direction.Z}
	low := [3]T{minimum.X, minimum.Y, minimum.Z}
	high := [3]T{maximum.X, maximum.Y, maximum.Z}

	var (
		entry T
		// Seeded at infinity so the first slab sets it; T is a float type,
		// so the conversion keeps the infinity for both widths.
		exit = T(math.Inf(1))
	)

	for axis := range 3 {
		if direction[axis] == 0 {
			// Parallel to this pair of planes: it can never enter, so it must
			// already be between them or it misses entirely.
			if origin[axis] < low[axis] || origin[axis] > high[axis] {
				return 0, false
			}

			continue
		}

		inverse := 1 / direction[axis]
		near := (low[axis] - origin[axis]) * inverse
		far := (high[axis] - origin[axis]) * inverse

		// A negative direction crosses the far plane first.
		if near > far {
			near, far = far, near
		}

		entry = max(entry, near)
		exit = min(exit, far)

		if entry > exit {
			return 0, false
		}
	}

	// entry starts at 0, so a ray beginning inside the box reports 0 and the
	// exit-behind-the-origin case is already excluded by entry > exit.
	if exit < 0 {
		return 0, false
	}

	return entry, true
}

// Ray3OBB3 returns the distance to where a ray first meets an oriented box.
//
// The ray is moved into the box's own space and tested against an axis-aligned
// one, rather than the box being rotated into the world. That is the cheaper
// direction — a ray is two vectors where a box is eight corners — and it reuses
// the slab test exactly rather than restating it against arbitrary axes.
//
// The rotation is rigid, so distances measured in the box's space are the same
// distances in the world and need no conversion back.
func Ray3OBB3[T gomath.Float](ray gomath.Ray3[T], box gomath.OBB3[T]) (distance T, hit bool) {
	if !ray.IsValid() {
		return 0, false
	}

	// Into the box's frame: shift so the centre is the origin, then unrotate by
	// projecting onto its axes.
	offset := ray.Origin.Sub(box.Center)

	local := gomath.Ray3[T]{
		Origin: gomath.Vector3[T]{
			X: offset.Dot(box.Axis(0)),
			Y: offset.Dot(box.Axis(1)),
			Z: offset.Dot(box.Axis(2)),
		},
		Direction: gomath.Vector3[T]{
			X: ray.Direction.Dot(box.Axis(0)),
			Y: ray.Direction.Dot(box.Axis(1)),
			Z: ray.Direction.Dot(box.Axis(2)),
		},
	}

	// The axes are orthonormal, so the projected direction is still unit length
	// and the local ray needs no renormalising.
	aligned := gomath.Box3[T]{
		Center:   gomath.Vector3[T]{},
		HalfSize: box.HalfSize,
	}

	return Ray3Box3(local, aligned)
}

// Ray3Capsule3 returns the distance to where a ray first meets a capsule.
//
// A ray starting inside reports 0, per the note at the top of this file.
//
// # Three candidates, not a cascade
//
// A capsule is a cylindrical body with a hemisphere at each end, and a ray can
// enter through any of the three. The textbook version tests the body first and
// falls through to the caps only when the body's hit lands outside the segment's
// span; that is correct but easy to get subtly wrong, because the fall-through
// has to reason about which cap the ray could still reach.
//
// This computes all three candidate entry distances and keeps the smallest, which
// needs no such reasoning: the capsule is convex and the ray has already been
// shown to start outside it, so its first crossing of the surface is simply the
// nearest of the crossings of the pieces. The price is two sphere tests on every
// call, and a capsule is not what a picking loop spends its time on — the boxes
// of the scenery are.
func Ray3Capsule3[T gomath.Float](ray gomath.Ray3[T], capsule gomath.Capsule3[T]) (distance T, hit bool) {
	if !ray.IsValid() {
		return 0, false
	}

	if capsule.ContainsPoint(ray.Origin) {
		return 0, true
	}

	nearest := T(math.Inf(1))
	found := false

	keep := func(candidate T, ok bool) {
		if ok && candidate < nearest {
			nearest, found = candidate, true
		}
	}

	keep(rayCapsuleBody(ray, capsule))
	keep(Ray3Sphere(ray, gomath.Sphere[T]{Center: capsule.Segment.P1, Radius: capsule.Radius}))
	keep(Ray3Sphere(ray, gomath.Sphere[T]{Center: capsule.Segment.P2, Radius: capsule.Radius}))

	if !found {
		return 0, false
	}

	return nearest, true
}

// rayCapsuleBody returns the distance at which a ray enters the cylindrical part
// of a capsule — the part strictly between the two caps.
//
// Substituting the ray into the infinite cylinder about the segment's line gives
// a quadratic. Every coefficient below is that quadratic's, multiplied through by
// the segment's squared length so that no division by it is needed and a
// degenerate segment simply produces a quadratic with no solutions rather than a
// NaN. Its caller covers that case with the caps, which for a degenerate segment
// are the whole shape.
//
// A vanishing leading coefficient means the ray runs parallel to the axis, and
// such a ray can only ever enter through a cap.
func rayCapsuleBody[T gomath.Float](ray gomath.Ray3[T], capsule gomath.Capsule3[T]) (distance T, hit bool) {
	axis := capsule.Segment.Direction()
	toOrigin := ray.Origin.Sub(capsule.Segment.P1)

	axisLengthSq := axis.LengthSq()
	alongAxis := toOrigin.Dot(axis)
	directionAlongAxis := ray.Direction.Dot(axis)

	quadratic := axisLengthSq - directionAlongAxis*directionAlongAxis
	if quadratic <= 0 {
		return 0, false
	}

	linear := axisLengthSq*toOrigin.Dot(ray.Direction) - alongAxis*directionAlongAxis
	constant := axisLengthSq*(toOrigin.LengthSq()-capsule.Radius*capsule.Radius) - alongAxis*alongAxis

	discriminant := linear*linear - quadratic*constant
	if discriminant < 0 {
		return 0, false
	}

	distance = (-linear - gomath.Sqrt(discriminant)) / quadratic
	if distance < 0 {
		return 0, false
	}

	// Where that crossing sits along the axis, in units of its squared length.
	// Outside [0, length^2] the ray met the infinite cylinder beyond a cap, so it
	// is not a point of the capsule.
	projected := alongAxis + distance*directionAlongAxis
	if projected < 0 || projected > axisLengthSq {
		return 0, false
	}

	return distance, true
}
