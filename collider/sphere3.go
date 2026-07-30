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
	toCentre := sphere.Center.Sub(box.Center)

	overlapX := box.HalfSize.X - gomath.Abs(toCentre.X)
	overlapY := box.HalfSize.Y - gomath.Abs(toCentre.Y)
	overlapZ := box.HalfSize.Z - gomath.Abs(toCentre.Z)

	shallowest := overlapX
	axis := 0
	if overlapY < shallowest {
		shallowest, axis = overlapY, 1
	}
	if overlapZ < shallowest {
		shallowest, axis = overlapZ, 2
	}

	switch axis {
	case 0:
		normal = gomath.Vector3[T]{X: 1}
		if toCentre.X > 0 {
			normal.X = -1
		}
	case 1:
		normal = gomath.Vector3[T]{Y: 1}
		if toCentre.Y > 0 {
			normal.Y = -1
		}
	default:
		normal = gomath.Vector3[T]{Z: 1}
		if toCentre.Z > 0 {
			normal.Z = -1
		}
	}

	return sphere.Radius + shallowest, normal, true
}

// SphereOBB3 reports whether a sphere overlaps an oriented box.
func SphereOBB3[T gomath.Float](sphere gomath.Sphere[T], box gomath.OBB3[T]) bool {
	closest := box.ClosestPoint(sphere.Center)

	return closest.Sub(sphere.Center).LengthSq() <= sphere.Radius*sphere.Radius
}
