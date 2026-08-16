package collider

import "github.com/wolfgarnet/gomath"

// Sphere overlap. Normals point from the FIRST shape towards the second, as
// everywhere else here.
//
// A sphere test is one distance comparison, which is why it belongs first in any
// rejection chain: it costs a fraction of a box test and is invariant under
// rotation, so a bounding sphere never needs recomputing as an object turns.

// SphereSphere reports whether two spheres overlap. Touching counts as
// overlapping, matching the other same-shape tests here.
func SphereSphere[T gomath.Number](a gomath.Sphere[T], b gomath.Sphere[T]) bool {
	reach := a.Radius + b.Radius

	// Squared throughout, so no square root is needed for a yes/no answer.
	return b.Center.Sub(a.Center).LengthSq() <= reach*reach
}

// SphereSpherePenetration returns the minimum translation separating two
// overlapping spheres, with the normal pointing from a towards b.
//
// Concentric spheres have no unique separation direction, so an arbitrary axis is
// chosen rather than returning a zero normal a solver would divide by.
func SphereSpherePenetration[T gomath.Float](a gomath.Sphere[T], b gomath.Sphere[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	offset := b.Center.Sub(a.Center)
	reach := a.Radius + b.Radius

	distanceSq := offset.LengthSq()
	if distanceSq > reach*reach {
		return 0, normal, false
	}

	distance := gomath.Sqrt(distanceSq)
	if distance == 0 {
		// Same centre: fully overlapping with no direction implied.
		return reach, gomath.Vector3[T]{X: 1}, true
	}

	return reach - distance, offset.DivScalar(distance), true
}

// SphereBox3 reports whether a sphere overlaps an axis-aligned box.
//
// The test is against the box's closest point to the centre, which is what makes
// it exact rather than conservative: a sphere near a box corner is correctly
// rejected, where comparing bounding volumes would not be.
func SphereBox3[T gomath.Number](sphere gomath.Sphere[T], box gomath.Box3[T]) bool {
	closest := box.ClosestPoint(sphere.Center)

	return closest.Sub(sphere.Center).LengthSq() <= sphere.Radius*sphere.Radius
}

// SphereBox3Penetration returns the minimum translation separating a sphere from
// an axis-aligned box, with the normal pointing from the sphere towards the box.
//
// A centre inside the box has no closest surface point to push away from, so the
// shallowest face is used instead — the same choice Box3Box3Penetration makes.
func SphereBox3Penetration[T gomath.Float](sphere gomath.Sphere[T], box gomath.Box3[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	closest := box.ClosestPoint(sphere.Center)
	offset := closest.Sub(sphere.Center)

	distanceSq := offset.LengthSq()
	if distanceSq > sphere.Radius*sphere.Radius {
		return 0, normal, false
	}

	if distanceSq > 0 {
		distance := gomath.Sqrt(distanceSq)

		return sphere.Radius - distance, offset.DivScalar(distance), true
	}

	// The centre is inside the box: fall back to the shallowest face.
	shallowest, normal := shallowestFaceNormal(box.HalfSize, sphere.Center.Sub(box.Center))

	return sphere.Radius + shallowest, normal, true
}

// SphereOBB3 reports whether a sphere overlaps an oriented box.
func SphereOBB3[T gomath.Float](sphere gomath.Sphere[T], box gomath.OBB3[T]) bool {
	closest := box.ClosestPoint(sphere.Center)

	return closest.Sub(sphere.Center).LengthSq() <= sphere.Radius*sphere.Radius
}

// SphereOBB3Penetration returns the minimum translation separating a sphere from
// an oriented box, with the normal pointing from the sphere towards the box.
//
// The whole problem is SphereBox3Penetration's once the sphere's centre is
// expressed in the box's frame, where the box is the axis-aligned span
// [-HalfSize, +HalfSize]. Only the resulting direction has to come back out to
// world space; the depth is a length, and a rotation does not change lengths.
//
// This is the pair a character-against-scenery test produces most often, which is
// why it is worth solving in the box's frame rather than falling back to
// OBB3OBB3Penetration on the sphere's bounding box — that would answer a
// different, looser question.
func SphereOBB3Penetration[T gomath.Float](sphere gomath.Sphere[T], box gomath.OBB3[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	local := box.ToLocal(sphere.Center)
	closest := local.Clamp(box.HalfSize.Negate(), box.HalfSize)
	offset := closest.Sub(local)

	distanceSq := offset.LengthSq()
	if distanceSq > sphere.Radius*sphere.Radius {
		return 0, normal, false
	}

	if distanceSq > 0 {
		distance := gomath.Sqrt(distanceSq)

		return sphere.Radius - distance, localDirectionToWorld(box, offset.DivScalar(distance)), true
	}

	// The centre is inside the box, so the closest point IS the centre and the
	// offset above carries no direction. The axis has to come from the box.
	shallowest, localNormal := shallowestFaceNormal(box.HalfSize, local)

	return sphere.Radius + shallowest, localDirectionToWorld(box, localNormal), true
}

// OBB3SpherePenetration returns the minimum translation separating an oriented
// box from a sphere, with the normal pointing from the box towards the sphere.
//
// It exists so that neither argument order is a trap: the pair resolves whichever
// way round the caller holds it, and the normal always points first towards
// second.
func OBB3SpherePenetration[T gomath.Float](box gomath.OBB3[T], sphere gomath.Sphere[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	penetration, normal, overlapping = SphereOBB3Penetration(sphere, box)
	if !overlapping {
		return 0, gomath.Vector3[T]{}, false
	}

	return penetration, normal.Negate(), true
}

// Box3SpherePenetration returns the minimum translation separating an
// axis-aligned box from a sphere, with the normal pointing from the box towards
// the sphere. It is SphereBox3Penetration with the arguments the other way round.
func Box3SpherePenetration[T gomath.Float](box gomath.Box3[T], sphere gomath.Sphere[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	penetration, normal, overlapping = SphereBox3Penetration(sphere, box)
	if !overlapping {
		return 0, gomath.Vector3[T]{}, false
	}

	return penetration, normal.Negate(), true
}

// shallowestFaceNormal returns the distance from a point inside a box to its
// nearest face, and the unit direction pointing from that point deeper INTO the
// box — which is the shape1->shape2 normal when shape1 is the thing inside.
//
// Both sphere-box penetrations need this. The closest-point formulation they are
// otherwise built on degenerates when the centre is inside: the closest point is
// the centre itself, the difference vector is zero, and normalising it gives NaN
// or a zero normal depending on the implementation. Neither is a direction a
// solver can push along, so the direction has to be recovered from the box's own
// geometry instead, and the least-penetrated face is the one that separates them
// by moving the shortest distance.
//
// The point is given in the box's frame, so an oriented box passes its local
// coordinates and rotates the result back afterwards.
func shallowestFaceNormal[T gomath.Float](halfSize gomath.Vector3[T], point gomath.Vector3[T]) (depth T, normal gomath.Vector3[T]) {
	depth = halfSize.X - gomath.Abs(point.X)
	normal = gomath.Vector3[T]{X: 1}
	if point.X > 0 {
		normal.X = -1
	}

	if overlapY := halfSize.Y - gomath.Abs(point.Y); overlapY < depth {
		depth = overlapY
		normal = gomath.Vector3[T]{Y: 1}
		if point.Y > 0 {
			normal.Y = -1
		}
	}

	if overlapZ := halfSize.Z - gomath.Abs(point.Z); overlapZ < depth {
		depth = overlapZ
		normal = gomath.Vector3[T]{Z: 1}
		if point.Z > 0 {
			normal.Z = -1
		}
	}

	return depth, normal
}

// localDirectionToWorld maps a direction from a box's frame back to world space.
//
// It is OBB3.ToWorld without the translation, since a direction has no origin to
// be offset from. The basis is orthonormal, so the result is still unit length.
func localDirectionToWorld[T gomath.Float](box gomath.OBB3[T], local gomath.Vector3[T]) gomath.Vector3[T] {
	return box.Axis(0).MultiplyScalar(local.X).
		Add(box.Axis(1).MultiplyScalar(local.Y)).
		Add(box.Axis(2).MultiplyScalar(local.Z))
}
