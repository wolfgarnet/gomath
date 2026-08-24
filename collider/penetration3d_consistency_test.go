package collider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wolfgarnet/gomath"
)

// The 3D penetration matrix must agree on one MTV convention: the normal is unit
// length and points from the FIRST shape towards the second, and the penetration
// is the depth along it. This is gomath's CLAUDE.md rule, anchored for 2D by
// penetration_consistency_test.go and for 3D here.
//
// It lives in its own file rather than joining the 2D table because the two share
// no fixtures — Vector2 against Vector3 — and the 2D one is long enough already.
//
// The table below carries one row per ORDERED pair, so SphereBox3Penetration and
// Box3SpherePenetration are separate rows rather than one row checked twice. A
// reversed wrapper is exactly where a sign flip hides, so each direction is
// asserted in its own right.
//
// float32 is the consumer's instantiation (pixelmaster), so the whole harness runs
// in float32 and the tolerances are ones float32 can actually hold. A 1e-9
// epsilon written against float64 intuition would fail here for no real reason.
const (
	// consistencyTolerance is roughly 2^-17 relative to the unit-scale fixtures
	// below, which is comfortably above float32's ~1.2e-7 epsilon after the
	// square roots and dot products these functions run through.
	consistencyTolerance = 1e-5

	// separationTolerance is looser because it measures a residual after the
	// shapes have been moved: the shift itself rounds, and the re-test then
	// compounds that rounding through the same arithmetic again.
	separationTolerance = 1e-4
)

// penetrate3D calls a 3D penetration function with the pair's second shape — the
// one placed to the +X side — displaced by shift.
//
// Type-erasing to a closure is what lets one table cover pairs whose argument
// types differ. shift always moves the same physical shape, including in the
// swapped direction where that shape is the first argument, so the harness can
// push the pair apart without knowing which kind it holds.
type penetrate3D func(shift gomath.Vector3[float32]) (penetration float32, normal gomath.Vector3[float32], overlapping bool)

// pair3D is one ordered pair of the matrix.
type pair3D struct {
	name string

	// forward is the pair as named: shape1 at the origin, shape2 to the +X side.
	forward penetrate3D
	// swapped is the same two shapes through the reversed-argument function.
	swapped penetrate3D

	// wantPenetration is the overlap depth in the base configuration.
	wantPenetration float32
	// wantNormal is the expected forward normal. Every fixture is built so this
	// is determinate; see obliqueRotation for the one case that is not, which
	// leaves this zero and is checked by the direction invariant alone.
	wantNormal gomath.Vector3[float32]

	// clearOfEachOther displaces shape2 far enough that the pair cannot overlap.
	clearOfEachOther gomath.Vector3[float32]
}

// rotationAboutX returns a basis rotated by theta radians about the world X axis.
//
// Every rotated fixture below turns about X and is then separated along X, which
// keeps the least-overlap axis exactly +X: a rotation about X leaves the X axis
// fixed, so it stays a face normal of both boxes, while the Y and Z overlaps
// remain large because the centres are level on those axes. That makes the
// expected normal exact rather than something the test has to approximate, while
// still exercising a genuinely oriented box.
func rotationAboutX(theta float32) gomath.Quaternion[float32] {
	return gomath.QuaternionFromAxisAngle(gomath.Vector3[float32]{X: 1}, theta)
}

// obliqueRotation is a rotation about no principal axis, used for the one row
// whose MTV axis is not determinate by inspection.
func obliqueRotation() gomath.Quaternion[float32] {
	return gomath.QuaternionFromEulerXYZ[float32](0.4, 0.7, 0.3)
}

func unitHalfSize() gomath.Vector3[float32] {
	return gomath.Vector3[float32]{X: 1, Y: 1, Z: 1}
}

// The fixtures all put shape1 at the origin and shape2 at (1.5, 0, 0) with unit
// extents or radii, giving a 0.5 overlap along X in every row. One shared
// geometry across the whole matrix is what makes the rows comparable: a row that
// disagrees is disagreeing about the convention, not about its own arrangement.
func penetration3DPairs() []pair3D {
	const gap = 1.5

	offsetCentre := func(shift gomath.Vector3[float32]) gomath.Vector3[float32] {
		return gomath.Vector3[float32]{X: gap}.Add(shift)
	}

	origin := gomath.Vector3[float32]{}
	plusX := gomath.Vector3[float32]{X: 1}
	wellClear := gomath.Vector3[float32]{X: 10}

	sphereAt := func(centre gomath.Vector3[float32]) gomath.Sphere[float32] {
		return gomath.Sphere[float32]{Center: centre, Radius: 1}
	}
	box3At := func(centre gomath.Vector3[float32]) gomath.Box3[float32] {
		return gomath.Box3[float32]{Center: centre, HalfSize: unitHalfSize()}
	}
	obb3At := func(centre gomath.Vector3[float32], rotation gomath.Quaternion[float32]) gomath.OBB3[float32] {
		return gomath.NewOBB3FromQuaternion(centre, unitHalfSize(), rotation)
	}
	// The capsule fixture is a unit-radius capsule whose segment runs along Y, so
	// it reaches exactly 1 along X like every other shape here and the row is
	// directly comparable with the rest of the table. A segment along X would have
	// made the capsule longer than the others in the very direction the fixtures
	// separate along.
	capsule3At := func(centre gomath.Vector3[float32]) gomath.Capsule3[float32] {
		return gomath.NewCapsule3(
			centre.Sub(gomath.Vector3[float32]{Y: 0.5}),
			centre.Add(gomath.Vector3[float32]{Y: 0.5}),
			1,
		)
	}

	return []pair3D{
		{
			name: "Sphere/Sphere",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return SphereSpherePenetration(sphereAt(origin), sphereAt(offsetCentre(shift)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return SphereSpherePenetration(sphereAt(offsetCentre(shift)), sphereAt(origin))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "Sphere/Box3",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return SphereBox3Penetration(sphereAt(origin), box3At(offsetCentre(shift)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Box3SpherePenetration(box3At(offsetCentre(shift)), sphereAt(origin))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "Box3/Sphere",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Box3SpherePenetration(box3At(origin), sphereAt(offsetCentre(shift)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return SphereBox3Penetration(sphereAt(offsetCentre(shift)), box3At(origin))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "Sphere/OBB3",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return SphereOBB3Penetration(sphereAt(origin), obb3At(offsetCentre(shift), rotationAboutX(0.5)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return OBB3SpherePenetration(obb3At(offsetCentre(shift), rotationAboutX(0.5)), sphereAt(origin))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "OBB3/Sphere",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return OBB3SpherePenetration(obb3At(origin, rotationAboutX(0.5)), sphereAt(offsetCentre(shift)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return SphereOBB3Penetration(sphereAt(offsetCentre(shift)), obb3At(origin, rotationAboutX(0.5)))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "Box3/Box3",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Box3Box3Penetration(box3At(origin), box3At(offsetCentre(shift)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Box3Box3Penetration(box3At(offsetCentre(shift)), box3At(origin))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "Box3/OBB3",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Box3OBB3Penetration(box3At(origin), obb3At(offsetCentre(shift), rotationAboutX(0.5)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return OBB3Box3Penetration(obb3At(offsetCentre(shift), rotationAboutX(0.5)), box3At(origin))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "OBB3/Box3",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return OBB3Box3Penetration(obb3At(origin, rotationAboutX(0.5)), box3At(offsetCentre(shift)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Box3OBB3Penetration(box3At(offsetCentre(shift)), obb3At(origin, rotationAboutX(0.5)))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "OBB3/OBB3",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return OBB3OBB3Penetration(
					obb3At(origin, rotationAboutX(0.5)),
					obb3At(offsetCentre(shift), rotationAboutX(-0.3)),
				)
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return OBB3OBB3Penetration(
					obb3At(offsetCentre(shift), rotationAboutX(-0.3)),
					obb3At(origin, rotationAboutX(0.5)),
				)
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			// The one row whose MTV axis is not determinate by inspection: both
			// boxes are turned about all three axes, so the minimum could fall on
			// a face normal or an edge axis. wantNormal is left zero and only the
			// direction invariant is asserted — which is the part that matters,
			// since an oblique pair is exactly where a sign convention that only
			// holds for axis-aligned cases would go unnoticed.
			name: "OBB3/OBB3 oblique",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return OBB3OBB3Penetration(
					obb3At(origin, obliqueRotation()),
					obb3At(offsetCentre(shift), rotationAboutX(-0.3)),
				)
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return OBB3OBB3Penetration(
					obb3At(offsetCentre(shift), rotationAboutX(-0.3)),
					obb3At(origin, obliqueRotation()),
				)
			},
			clearOfEachOther: wellClear,
		},
		{
			// The centre-inside case, which is where a closest-point formulation
			// degenerates: the closest point is the centre, the offset vector is
			// zero, and normalising it gives NaN or a zero normal. A resolver has
			// to dig bodies out of exactly this state.
			name: "Sphere/OBB3 centre inside",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return SphereOBB3Penetration(
					gomath.Sphere[float32]{Center: gomath.Vector3[float32]{X: -0.8}, Radius: 0.25},
					obb3At(shift, rotationAboutX(0.5)),
				)
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return OBB3SpherePenetration(
					obb3At(shift, rotationAboutX(0.5)),
					gomath.Sphere[float32]{Center: gomath.Vector3[float32]{X: -0.8}, Radius: 0.25},
				)
			},
			// The centre sits 0.2 inside the -X face, which is the shallowest, so
			// the sphere is pushed out along -X and the normal — pointing sphere
			// towards box — is +X.
			wantPenetration:  0.45,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "Capsule3/Sphere",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Capsule3SpherePenetration(capsule3At(origin), sphereAt(offsetCentre(shift)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return SphereCapsule3Penetration(sphereAt(offsetCentre(shift)), capsule3At(origin))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "Sphere/Capsule3",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return SphereCapsule3Penetration(sphereAt(origin), capsule3At(offsetCentre(shift)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Capsule3SpherePenetration(capsule3At(offsetCentre(shift)), sphereAt(origin))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			// Both segments run along Y, so this row also exercises the parallel
			// branch of LineSegment3.ClosestParameters — the arrangement two
			// standing characters produce, and the one with no unique closest pair.
			name: "Capsule3/Capsule3",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Capsule3Capsule3Penetration(capsule3At(origin), capsule3At(offsetCentre(shift)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Capsule3Capsule3Penetration(capsule3At(offsetCentre(shift)), capsule3At(origin))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "Capsule3/Box3",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Capsule3Box3Penetration(capsule3At(origin), box3At(offsetCentre(shift)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Box3Capsule3Penetration(box3At(offsetCentre(shift)), capsule3At(origin))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "Box3/Capsule3",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Box3Capsule3Penetration(box3At(origin), capsule3At(offsetCentre(shift)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Capsule3Box3Penetration(capsule3At(offsetCentre(shift)), box3At(origin))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "Capsule3/OBB3",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Capsule3OBB3Penetration(capsule3At(origin), obb3At(offsetCentre(shift), rotationAboutX(0.5)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return OBB3Capsule3Penetration(obb3At(offsetCentre(shift), rotationAboutX(0.5)), capsule3At(origin))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			name: "OBB3/Capsule3",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return OBB3Capsule3Penetration(obb3At(origin, rotationAboutX(0.5)), capsule3At(offsetCentre(shift)))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Capsule3OBB3Penetration(capsule3At(offsetCentre(shift)), obb3At(origin, rotationAboutX(0.5)))
			},
			wantPenetration:  0.5,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
		{
			// The capsule counterpart of the row above: the segment lies INSIDE the
			// box, so there is no closest pair to take a direction from and the
			// answer comes from the separating-axis branch instead. The segment runs
			// along Z at x = -0.8, which puts it 0.2 inside the -X face — the
			// shallowest — so the capsule leaves along -X and the normal, pointing
			// capsule towards box, is +X.
			name: "Capsule3/Box3 segment inside",
			forward: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Capsule3Box3Penetration(piercingCapsule(), box3At(shift))
			},
			swapped: func(shift gomath.Vector3[float32]) (float32, gomath.Vector3[float32], bool) {
				return Box3Capsule3Penetration(box3At(shift), piercingCapsule())
			},
			wantPenetration:  0.45,
			wantNormal:       plusX,
			clearOfEachOther: wellClear,
		},
	}
}

// piercingCapsule is a capsule whose segment lies wholly inside the unit box at
// the origin, used by the row that exercises the separating-axis branch.
func piercingCapsule() gomath.Capsule3[float32] {
	return gomath.NewCapsule3(
		gomath.Vector3[float32]{X: -0.8, Z: -0.4},
		gomath.Vector3[float32]{X: -0.8, Z: 0.4},
		0.25,
	)
}

// TestPenetration3D_NormalPointsShape1ToShape2 is part (a): every function agrees
// that the normal is unit length and points from the first shape towards the
// second.
func TestPenetration3D_NormalPointsShape1ToShape2(t *testing.T) {
	for _, pair := range penetration3DPairs() {
		t.Run(pair.name, func(t *testing.T) {
			penetration, normal, overlapping := pair.forward(gomath.Vector3[float32]{})
			require.True(t, overlapping, "fixture is meant to overlap")

			assert.InDelta(t, 1, normal.Length(), consistencyTolerance, "normal must be unit length")
			assert.Positive(t, penetration, "penetration must be positive while overlapping")

			// The invariant itself: the pair's second shape lies to the +X side,
			// so a shape1->shape2 normal has a positive X component.
			assert.Positive(t, normal.X, "normal must point towards the second shape")

			if !pair.wantNormal.IsZero() {
				assertVector3InDelta(t, pair.wantNormal, normal, consistencyTolerance, "normal")
				assert.InDelta(t, pair.wantPenetration, penetration, consistencyTolerance, "penetration")
			}
		})
	}
}

// TestPenetration3D_MTVSeparates is part (b): the returned vector really is a
// minimum TRANSLATION vector — moving the second shape along it by the returned
// depth ends the overlap.
//
// Touching counts as overlapping throughout this package, so an exactly-resolved
// pair legitimately still reports true with a depth of ~0. The assertion is that
// the residual depth is gone, not that the flag has flipped.
func TestPenetration3D_MTVSeparates(t *testing.T) {
	for _, pair := range penetration3DPairs() {
		t.Run(pair.name, func(t *testing.T) {
			penetration, normal, overlapping := pair.forward(gomath.Vector3[float32]{})
			require.True(t, overlapping, "fixture is meant to overlap")

			residual, _, stillOverlapping := pair.forward(normal.MultiplyScalar(penetration))
			if stillOverlapping {
				assert.LessOrEqual(t, residual, float32(separationTolerance),
					"moving by penetration*normal must leave at most a rounding residual")
			}

			// And a hair further must clear it outright, which rules out a depth
			// that happens to be too small.
			_, _, closeButClear := pair.forward(normal.MultiplyScalar(penetration + 0.01))
			assert.False(t, closeButClear, "moving slightly past the MTV must separate the pair")
		})
	}
}

// TestPenetration3D_SwappedArgumentsNegateNormal is part (c): calling the pair the
// other way round returns the same depth and the opposite normal.
//
// This is the assertion the reversed wrappers exist to satisfy, and the one that
// proves the convention is uniform enough for a consumer to flip it once,
// centrally, instead of at every call site.
func TestPenetration3D_SwappedArgumentsNegateNormal(t *testing.T) {
	for _, pair := range penetration3DPairs() {
		t.Run(pair.name, func(t *testing.T) {
			forwardPenetration, forwardNormal, forwardOverlapping := pair.forward(gomath.Vector3[float32]{})
			swappedPenetration, swappedNormal, swappedOverlapping := pair.swapped(gomath.Vector3[float32]{})

			require.True(t, forwardOverlapping)
			assert.Equal(t, forwardOverlapping, swappedOverlapping, "both orders must agree on overlap")

			assert.InDelta(t, forwardPenetration, swappedPenetration, consistencyTolerance,
				"depth is a distance and does not depend on argument order")
			assertVector3InDelta(t, forwardNormal.Negate(), swappedNormal, consistencyTolerance,
				"swapped normal")
		})
	}
}

// TestPenetration3D_NonOverlappingReturnsZero is part (d): a separated pair
// reports false and leaves the other two returns at zero.
//
// The zeroes matter because callers must not read them, and a stale non-zero
// normal is precisely what a caller that forgets to check the flag would sail
// past without noticing.
func TestPenetration3D_NonOverlappingReturnsZero(t *testing.T) {
	for _, pair := range penetration3DPairs() {
		t.Run(pair.name, func(t *testing.T) {
			for _, direction := range []struct {
				name string
				call penetrate3D
			}{
				{"forward", pair.forward},
				{"swapped", pair.swapped},
			} {
				t.Run(direction.name, func(t *testing.T) {
					penetration, normal, overlapping := direction.call(pair.clearOfEachOther)

					assert.False(t, overlapping, "shapes are well clear of each other")
					assert.Zero(t, penetration, "penetration must be zero when not overlapping")
					assert.True(t, normal.IsZero(), "normal must be zero when not overlapping, got %v", normal)
				})
			}
		})
	}
}

func assertVector3InDelta(t *testing.T, want gomath.Vector3[float32], got gomath.Vector3[float32], delta float64, label string) {
	t.Helper()

	assert.InDeltaf(t, want.X, got.X, delta, "%s.X", label)
	assert.InDeltaf(t, want.Y, got.Y, delta, "%s.Y", label)
	assert.InDeltaf(t, want.Z, got.Z, delta, "%s.Z", label)
}
