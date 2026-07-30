package collider

import "github.com/wolfgarnet/gomath"

// 3D box overlap: axis-aligned against axis-aligned, oriented against oriented,
// and the mixed case.
//
// Normals point from the FIRST shape towards the second, matching the
// convention every other penetration function here follows.

// Box3Box3 reports whether two axis-aligned boxes overlap.
//
// Boxes that merely touch count as overlapping, matching AABBAABB in 2D.
//
// This exists separately from OBB3OBB3 rather than routing through it with
// identity rotations because it is six comparisons against fifteen axis tests.
// Axis-aligned pairs are the common case in a broadphase, and paying the
// oriented price for them is the single easiest way to make collision slow.
func Box3Box3[T gomath.Number](a gomath.Box3[T], b gomath.Box3[T]) bool {
	offset := b.Center.Sub(a.Center)

	return gomath.Abs(offset.X) <= a.HalfSize.X+b.HalfSize.X &&
		gomath.Abs(offset.Y) <= a.HalfSize.Y+b.HalfSize.Y &&
		gomath.Abs(offset.Z) <= a.HalfSize.Z+b.HalfSize.Z
}

// Box3Box3Penetration returns the minimum translation that separates two
// overlapping axis-aligned boxes.
//
// The normal is the world axis of least overlap and points from a towards b.
func Box3Box3Penetration[T gomath.Number](a gomath.Box3[T], b gomath.Box3[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	offset := b.Center.Sub(a.Center)

	overlapX := a.HalfSize.X + b.HalfSize.X - gomath.Abs(offset.X)
	overlapY := a.HalfSize.Y + b.HalfSize.Y - gomath.Abs(offset.Y)
	overlapZ := a.HalfSize.Z + b.HalfSize.Z - gomath.Abs(offset.Z)

	if overlapX < 0 || overlapY < 0 || overlapZ < 0 {
		return 0, normal, false
	}

	penetration = overlapX
	axis := 0
	if overlapY < penetration {
		penetration, axis = overlapY, 1
	}
	if overlapZ < penetration {
		penetration, axis = overlapZ, 2
	}

	switch axis {
	case 0:
		normal = gomath.Vector3[T]{X: 1}
		if offset.X < 0 {
			normal.X = -1
		}
	case 1:
		normal = gomath.Vector3[T]{Y: 1}
		if offset.Y < 0 {
			normal.Y = -1
		}
	default:
		normal = gomath.Vector3[T]{Z: 1}
		if offset.Z < 0 {
			normal.Z = -1
		}
	}

	return penetration, normal, true
}

// Box3OBB3 reports whether an axis-aligned box overlaps an oriented one.
func Box3OBB3[T gomath.Float](a gomath.Box3[T], b gomath.OBB3[T]) bool {
	return OBB3OBB3(gomath.NewOBB3FromBox3(a), b)
}

// OBB3Box3 reports whether an oriented box overlaps an axis-aligned one.
func OBB3Box3[T gomath.Float](a gomath.OBB3[T], b gomath.Box3[T]) bool {
	return OBB3OBB3(a, gomath.NewOBB3FromBox3(b))
}

// Box3OBB3Penetration returns the minimum translation separating an axis-aligned
// box from an oriented one, with the normal pointing from a towards b.
func Box3OBB3Penetration[T gomath.Float](a gomath.Box3[T], b gomath.OBB3[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	return OBB3OBB3Penetration(gomath.NewOBB3FromBox3(a), b)
}
