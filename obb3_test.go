package gomath

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const obb3Tolerance = 1e-9

// onOrInside reports whether a point lies on or within the box, allowing for
// floating-point slack.
//
// ContainsPoint deliberately has no tolerance, and a point derived from the
// box's own geometry — a corner, or a clamped closest point — lands exactly on
// the boundary in exact arithmetic and a rounding step outside it in practice.
// Asserting ContainsPoint on such a point tests the arithmetic's luck, not the
// code.
func onOrInside(box OBB3[float64], point Vector3[float64], tolerance float64) bool {
	local := box.ToLocal(point)

	return math.Abs(local.X) <= box.HalfSize.X+tolerance &&
		math.Abs(local.Y) <= box.HalfSize.Y+tolerance &&
		math.Abs(local.Z) <= box.HalfSize.Z+tolerance
}

func quarterTurnAboutY[T Float](half Vector3[T]) OBB3[T] {
	return NewOBB3FromQuaternion(Vector3[T]{},
		half,
		QuaternionFromAxisAngle(Vector3[T]{Y: 1}, T(math.Pi/2)))
}

func TestOBB3AxisComesFromTheBasisColumns(t *testing.T) {
	// A quarter turn about Y takes local +X to world -Z and local +Z to world +X.
	box := quarterTurnAboutY(One3[float64]())

	xAxis := box.Axis(0)
	assert.InDelta(t, 0, xAxis.X, obb3Tolerance)
	assert.InDelta(t, -1, xAxis.Z, obb3Tolerance)

	zAxis := box.Axis(2)
	assert.InDelta(t, 1, zAxis.X, obb3Tolerance)
	assert.InDelta(t, 0, zAxis.Z, obb3Tolerance)

	// An identity box's axes are the world axes.
	identity := NewOBB3(Vector3[float64]{}, One3[float64](), Identity3[float64]())
	assert.Equal(t, Vector3[float64]{X: 1}, identity.Axis(0))
	assert.Equal(t, Vector3[float64]{Y: 1}, identity.Axis(1))
	assert.Equal(t, Vector3[float64]{Z: 1}, identity.Axis(2))
}

func TestOBB3LocalWorldRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(61))

	for i := 0; i < 200; i++ {
		axis := Vector3[float64]{
			X: rng.Float64()*2 - 1,
			Y: rng.Float64()*2 - 1,
			Z: rng.Float64()*2 - 1,
		}.Normalize()
		if axis.IsZero() {
			continue
		}

		box := NewOBB3FromQuaternion(
			Vector3[float64]{X: rng.Float64() * 10, Y: rng.Float64() * 10, Z: rng.Float64() * 10},
			Vector3[float64]{X: 1, Y: 2, Z: 3},
			QuaternionFromAxisAngle(axis, (rng.Float64()*2-1)*math.Pi))

		point := Vector3[float64]{rng.Float64() * 5, rng.Float64() * 5, rng.Float64() * 5}

		recovered := box.ToWorld(box.ToLocal(point))

		assert.InDelta(t, point.X, recovered.X, 1e-9)
		assert.InDelta(t, point.Y, recovered.Y, 1e-9)
		assert.InDelta(t, point.Z, recovered.Z, 1e-9)
	}
}

func TestOBB3ToLocalPutsTheCentreAtTheOrigin(t *testing.T) {
	box := NewOBB3FromQuaternion(
		Vector3[float64]{X: 4, Y: 5, Z: 6},
		One3[float64](),
		QuaternionFromAxisAngle(Vector3[float64]{Z: 1}, 0.7))

	local := box.ToLocal(box.Center)

	assert.InDelta(t, 0, local.X, obb3Tolerance)
	assert.InDelta(t, 0, local.Y, obb3Tolerance)
	assert.InDelta(t, 0, local.Z, obb3Tolerance)
}

func TestOBB3ContainsPoint(t *testing.T) {
	// Rotated a quarter turn about Y, the box's local X extent lies along world Z.
	box := quarterTurnAboutY(Vector3[float64]{X: 3, Y: 1, Z: 1})

	assert.True(t, box.ContainsPoint(Vector3[float64]{}))
	// Local X of 3 maps to world -Z of 3.
	assert.True(t, box.ContainsPoint(Vector3[float64]{Z: -2.9}))
	assert.False(t, box.ContainsPoint(Vector3[float64]{Z: -3.1}))
	// World X is now the short axis.
	assert.True(t, box.ContainsPoint(Vector3[float64]{X: 0.9}))
	assert.False(t, box.ContainsPoint(Vector3[float64]{X: 1.1}))
}

func TestOBB3ClosestPoint(t *testing.T) {
	box := NewOBB3(Vector3[float64]{}, One3[float64](), Identity3[float64]())

	// A point inside is its own closest point.
	inside := Vector3[float64]{X: 0.5, Y: -0.5}
	closest := box.ClosestPoint(inside)
	assert.InDelta(t, inside.X, closest.X, obb3Tolerance)
	assert.InDelta(t, inside.Y, closest.Y, obb3Tolerance)

	// A point outside clamps onto the surface, and the result must be on the box.
	outside := Vector3[float64]{X: 5, Y: 0.25, Z: -9}
	onSurface := box.ClosestPoint(outside)
	assert.InDelta(t, 1, onSurface.X, obb3Tolerance)
	assert.InDelta(t, 0.25, onSurface.Y, obb3Tolerance)
	assert.InDelta(t, -1, onSurface.Z, obb3Tolerance)
	assert.True(t, box.ContainsPoint(onSurface))
}

func TestOBB3ClosestPointOnARotatedBox(t *testing.T) {
	rng := rand.New(rand.NewSource(62))

	for i := 0; i < 200; i++ {
		axis := Vector3[float64]{
			X: rng.Float64()*2 - 1,
			Y: rng.Float64()*2 - 1,
			Z: rng.Float64()*2 - 1,
		}.Normalize()
		if axis.IsZero() {
			continue
		}

		box := NewOBB3FromQuaternion(
			Vector3[float64]{X: rng.Float64() * 4},
			Vector3[float64]{X: 1, Y: 2, Z: 0.5},
			QuaternionFromAxisAngle(axis, (rng.Float64()*2-1)*math.Pi))

		point := Vector3[float64]{
			X: (rng.Float64()*2 - 1) * 10,
			Y: (rng.Float64()*2 - 1) * 10,
			Z: (rng.Float64()*2 - 1) * 10,
		}

		closest := box.ClosestPoint(point)

		// The closest point is always on or in the box...
		assert.True(t, onOrInside(box, closest, 1e-9),
			"the closest point must lie on the box, got %v", closest)

		// ...and no corner is nearer than it.
		for _, corner := range box.Corners() {
			assert.LessOrEqual(t,
				closest.DistanceTo(point)-1e-9,
				corner.DistanceTo(point),
				"a corner was closer than the reported closest point")
		}
	}
}

func TestOBB3Corners(t *testing.T) {
	// An identity box's corners match the equivalent Box3's exactly, which ties
	// the two enumerations together.
	box := NewBox3(Vector3[float64]{X: 1, Y: 2, Z: 3}, Vector3[float64]{X: 1, Y: 2, Z: 3})
	oriented := NewOBB3FromBox3(box)

	boxCorners := box.Corners()
	orientedCorners := oriented.Corners()

	for index := range boxCorners {
		assert.InDelta(t, boxCorners[index].X, orientedCorners[index].X, obb3Tolerance)
		assert.InDelta(t, boxCorners[index].Y, orientedCorners[index].Y, obb3Tolerance)
		assert.InDelta(t, boxCorners[index].Z, orientedCorners[index].Z, obb3Tolerance)
	}
}

func TestOBB3CornersAreOnTheBox(t *testing.T) {
	box := NewOBB3FromQuaternion(
		Vector3[float64]{X: 2, Y: -1, Z: 3},
		Vector3[float64]{X: 1, Y: 2, Z: 0.5},
		QuaternionFromAxisAngle(Vector3[float64]{X: 1, Y: 1, Z: 1}.Normalize(), 0.9))

	for index, corner := range box.Corners() {
		assert.True(t, onOrInside(box, corner, 1e-9),
			"corner %d at %v should be on the box", index, corner)

		// A corner is the point where all three local coordinates are at their
		// extreme, which is a stronger claim than merely being inside.
		local := box.ToLocal(corner)
		assert.InDelta(t, box.HalfSize.X, math.Abs(local.X), 1e-9)
		assert.InDelta(t, box.HalfSize.Y, math.Abs(local.Y), 1e-9)
		assert.InDelta(t, box.HalfSize.Z, math.Abs(local.Z), 1e-9)
	}
}

func TestOBB3BoundingBoxEnclosesEveryCorner(t *testing.T) {
	rng := rand.New(rand.NewSource(63))

	for i := 0; i < 300; i++ {
		axis := Vector3[float64]{
			X: rng.Float64()*2 - 1,
			Y: rng.Float64()*2 - 1,
			Z: rng.Float64()*2 - 1,
		}.Normalize()
		if axis.IsZero() {
			continue
		}

		box := NewOBB3FromQuaternion(
			Vector3[float64]{X: rng.Float64() * 3, Y: rng.Float64() * 3, Z: rng.Float64() * 3},
			Vector3[float64]{X: rng.Float64() + 0.2, Y: rng.Float64() + 0.2, Z: rng.Float64() + 0.2},
			QuaternionFromAxisAngle(axis, (rng.Float64()*2-1)*math.Pi))

		bounds := box.BoundingBox()

		// Every corner inside, and the bound tight enough that shrinking it
		// slightly would exclude one — otherwise a trivially huge box would pass.
		corners := box.Corners()
		relaxed := bounds.Expanded(1e-9)
		for _, corner := range corners {
			assert.True(t, relaxed.ContainsPoint(corner),
				"the bounding box should contain corner %v", corner)
		}

		rebuilt, ok := NewBox3FromPoints(corners[:])
		require.True(t, ok)
		assert.True(t, bounds.ApproxEquals(rebuilt, 1e-9),
			"the bounding box should be tight: got %v want %v", bounds, rebuilt)
	}
}

func TestOBB3BoundingBoxOfAnAxisAlignedBoxIsItself(t *testing.T) {
	box := NewBox3(Vector3[float64]{X: 1, Y: 2, Z: 3}, Vector3[float64]{X: 4, Y: 5, Z: 6})

	assert.True(t, NewOBB3FromBox3(box).BoundingBox().ApproxEquals(box, obb3Tolerance))
}

func TestOBB3VolumeIsRotationInvariant(t *testing.T) {
	half := Vector3[float64]{X: 1, Y: 2, Z: 3}

	upright := NewOBB3(Vector3[float64]{}, half, Identity3[float64]())
	turned := quarterTurnAboutY(half)

	assert.InDelta(t, 48, upright.Volume(), obb3Tolerance)
	assert.InDelta(t, upright.Volume(), turned.Volume(), obb3Tolerance,
		"rotating a box does not change its volume")
}

func TestOBB3TranslateAndRotate(t *testing.T) {
	box := NewOBB3(Vector3[float64]{}, One3[float64](), Identity3[float64]())

	moved := box.Translate(Vector3[float64]{X: 5, Y: 6, Z: 7})
	assert.Equal(t, Vector3[float64]{X: 5, Y: 6, Z: 7}, moved.Center)
	assert.Equal(t, Vector3[float64]{}, box.Center, "Translate must not mutate the receiver")

	quarter := QuaternionFromAxisAngle(Vector3[float64]{Y: 1}, math.Pi/2).ToMatrix3()
	turned := box.Rotated(quarter)

	// Local +X now points along world -Z.
	assert.InDelta(t, -1, turned.Axis(0).Z, 1e-9)
	assert.Equal(t, Vector3[float64]{X: 1}, box.Axis(0), "Rotated must not mutate the receiver")
}

func TestOBB3String(t *testing.T) {
	box := NewOBB3(Vector3[float64]{}, One3[float64](), Identity3[float64]())

	assert.Contains(t, box.String(), "OBB3")
}
