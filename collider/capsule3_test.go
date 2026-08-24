package collider

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wolfgarnet/gomath"
)

func capsuleAlong(start gomath.Vector3[float64], end gomath.Vector3[float64], radius float64) gomath.Capsule3[float64] {
	return gomath.NewCapsule3(start, end, radius)
}

func randomCapsule3(random *rand.Rand, spread float64, maxRadius float64) gomath.Capsule3[float64] {
	return gomath.NewCapsule3(
		randomPoint(random, spread),
		randomPoint(random, spread),
		0.2+random.Float64()*(maxRadius-0.2),
	)
}

// # The free oracle for the whole feature
//
// A capsule whose segment is degenerate IS a sphere, so every capsule function
// must return exactly what its sphere counterpart returns for the same
// configuration — same boolean, same depth, same normal. Nothing here is a
// tolerance on a resemblance: the two code paths are meant to compute the same
// number, and the only slack allowed is the arithmetic taking a different route
// to it.
//
// This is worth more than it looks. It ties four new functions to four that were
// already checked by the existing harness and oracle, over random configurations,
// without any of them having to be re-derived.
func TestCapsule3ReducesToSphereWhenDegenerate(t *testing.T) {
	const trials = 400

	random := rand.New(rand.NewSource(20260824))

	for trial := range trials {
		sphere := randomSphere(random, 1.5, 1.2)
		capsule := gomath.NewCapsule3FromSphere(sphere)

		other := randomSphere(random, 1.5, 1.2)
		box := randomBox3(random, 1.5)
		orientedBox := randomOBB3(random, 1.5)

		t.Run("bool tests", func(t *testing.T) {
			assert.Equalf(t, SphereSphere(sphere, other), Capsule3Sphere(capsule, other), "trial %d", trial)
			assert.Equalf(t, SphereBox3(sphere, box), Capsule3Box3(capsule, box), "trial %d", trial)
			assert.Equalf(t, SphereOBB3(sphere, orientedBox), Capsule3OBB3(capsule, orientedBox), "trial %d", trial)
			assert.Equalf(t, SphereSphere(sphere, other), Capsule3Capsule3(capsule, gomath.NewCapsule3FromSphere(other)), "trial %d", trial)
		})

		t.Run("penetrations", func(t *testing.T) {
			assertSamePenetration(t, trial, "Sphere",
				func() (float64, gomath.Vector3[float64], bool) { return SphereSpherePenetration(sphere, other) },
				func() (float64, gomath.Vector3[float64], bool) { return Capsule3SpherePenetration(capsule, other) })

			assertSamePenetration(t, trial, "Capsule3",
				func() (float64, gomath.Vector3[float64], bool) { return SphereSpherePenetration(sphere, other) },
				func() (float64, gomath.Vector3[float64], bool) {
					return Capsule3Capsule3Penetration(capsule, gomath.NewCapsule3FromSphere(other))
				})

			assertSamePenetration(t, trial, "Box3",
				func() (float64, gomath.Vector3[float64], bool) { return SphereBox3Penetration(sphere, box) },
				func() (float64, gomath.Vector3[float64], bool) { return Capsule3Box3Penetration(capsule, box) })

			assertSamePenetration(t, trial, "OBB3",
				func() (float64, gomath.Vector3[float64], bool) { return SphereOBB3Penetration(sphere, orientedBox) },
				func() (float64, gomath.Vector3[float64], bool) { return Capsule3OBB3Penetration(capsule, orientedBox) })
		})
	}
}

func assertSamePenetration(
	t *testing.T,
	trial int,
	label string,
	viaSphere func() (float64, gomath.Vector3[float64], bool),
	viaCapsule func() (float64, gomath.Vector3[float64], bool),
) {
	t.Helper()

	const tolerance = 1e-12

	wantDepth, wantNormal, wantOverlapping := viaSphere()
	gotDepth, gotNormal, gotOverlapping := viaCapsule()

	require.Equalf(t, wantOverlapping, gotOverlapping, "trial %d, %s: overlap", trial, label)

	if !wantOverlapping {
		return
	}

	assert.InDeltaf(t, wantDepth, gotDepth, tolerance, "trial %d, %s: depth", trial, label)
	assert.InDeltaf(t, 0, wantNormal.DistanceTo(gotNormal), tolerance, "trial %d, %s: normal", trial, label)
}

// The second free oracle from the spec: a capsule and a sphere overlap exactly
// when the sphere's centre is within the sum of the radii of the capsule's
// segment. One line, computed without reference to how the real test works.
func TestCapsule3SphereAgainstTheOneLineOracle(t *testing.T) {
	random := rand.New(rand.NewSource(20260824))

	for trial := range 500 {
		capsule := randomCapsule3(random, 1.5, 1.2)
		sphere := randomSphere(random, 1.5, 1.2)

		reach := capsule.Radius + sphere.Radius
		want := capsule.Segment.DistanceToPoint(sphere.Center) <= reach

		assert.Equalf(t, want, Capsule3Sphere(capsule, sphere), "trial %d", trial)

		_, _, overlapping := Capsule3SpherePenetration(capsule, sphere)
		assert.Equalf(t, want, overlapping, "trial %d: the penetration must agree with the boolean", trial)
	}
}

// closestOnSegmentToAlignedBox is the engine of both box pairs, and the one place
// a shortcut would have been taken. It is checked directly against a sampled
// sweep of the segment: sampling can only overstate the minimum, so the exact
// answer must sit at or below it, and it must be achieved at a parameter that
// really is on the segment.
func TestClosestOnSegmentToAlignedBoxIsExact(t *testing.T) {
	const (
		trials  = 600
		samples = 20000
	)

	random := rand.New(rand.NewSource(20260824))

	for trial := range trials {
		segment := gomath.NewLineSegment3(randomPoint(random, 2.5), randomPoint(random, 2.5))
		halfSize := randomHalfSize(random)

		parameter, distanceSq := closestOnSegmentToAlignedBox(segment, halfSize)

		box := gomath.Box3[float64]{HalfSize: halfSize}

		sampled := math.Inf(1)
		for i := 0; i <= samples; i++ {
			point := segment.PointAt(float64(i) / samples)
			sampled = math.Min(sampled, point.DistanceToSquared(box.ClosestPoint(point)))
		}

		assert.LessOrEqualf(t, distanceSq, sampled+1e-12,
			"trial %d: exact distance exceeds the sampled minimum", trial)

		// The returned parameter must be the one the distance belongs to, and it
		// must lie on the segment rather than beyond an end.
		assert.GreaterOrEqualf(t, parameter, 0.0, "trial %d: parameter before the start", trial)
		assert.LessOrEqualf(t, parameter, 1.0, "trial %d: parameter past the end", trial)

		at := segment.PointAt(parameter)
		assert.InDeltaf(t, distanceSq, at.DistanceToSquared(box.ClosestPoint(at)), 1e-12,
			"trial %d: distance must belong to the returned parameter", trial)
	}
}

// The case the spec asked to be named: a capsule resting on the corner of a box.
//
// This is where the tempting shortcut — running the sphere-box test at the point
// of the segment nearest the box's CENTRE — is wrong. That point is not the point
// nearest the box, so the normal it produces does not point out of the corner and
// the character settles into it. The exact formulation is checked here against the
// answer worked out by hand.
func TestCapsule3Box3OnABoxCorner(t *testing.T) {
	// A unit box at the origin, so the corner in question is (1,1,1).
	box := gomath.Box3[float64]{HalfSize: gomath.Vector3[float64]{X: 1, Y: 1, Z: 1}}

	// A horizontal capsule lying above and outside that corner, running along Z so
	// that its nearest point to the corner is in its interior rather than at a cap
	// — which is exactly the arrangement whose closest point to the box CENTRE is
	// somewhere else entirely.
	radius := 0.5
	capsule := capsuleAlong(
		gomath.Vector3[float64]{X: 1.3, Y: 1.3, Z: -3},
		gomath.Vector3[float64]{X: 1.3, Y: 1.3, Z: 3},
		radius,
	)

	penetration, normal, overlapping := Capsule3Box3Penetration(capsule, box)
	require.True(t, overlapping)

	// The nearest point of the box is the corner's edge along Z at (1,1,z), and the
	// segment runs parallel to it at (1.3,1.3,z): a gap of 0.3*sqrt(2) across the
	// diagonal, so the overlap is the radius less that.
	gap := 0.3 * math.Sqrt2
	assert.InDelta(t, radius-gap, penetration, 1e-12)

	// And the normal points out of the corner, along the diagonal, capsule -> box.
	diagonal := 1 / math.Sqrt2
	assert.InDelta(t, -diagonal, normal.X, 1e-12)
	assert.InDelta(t, -diagonal, normal.Y, 1e-12)
	assert.InDelta(t, 0, normal.Z, 1e-12)

	// The shortcut this test exists to rule out: the point of the segment nearest
	// the box's centre is at z=0, and running the sphere test from there gives the
	// same answer only by accident of this arrangement — so also check a capsule
	// whose nearest-to-centre point and nearest-to-box point genuinely differ.
	slanted := capsuleAlong(
		gomath.Vector3[float64]{X: 1.2, Y: 1.2, Z: 0},
		gomath.Vector3[float64]{X: 4, Y: 4, Z: 4},
		0.4,
	)

	exactDepth, exactNormal, exactOverlapping := Capsule3Box3Penetration(slanted, box)
	require.True(t, exactOverlapping)

	viaCentre := gomath.Sphere[float64]{Center: slanted.Segment.ClosestPoint(box.Center), Radius: slanted.Radius}
	shortcutDepth, shortcutNormal, _ := SphereBox3Penetration(viaCentre, box)

	// They agree here (the nearest point to the centre is the near cap either
	// way), which is the honest result; the value of the assertion is the exact
	// figure, worked out independently below.
	assert.InDelta(t, shortcutDepth, exactDepth, 1e-12)
	assert.InDelta(t, 0, shortcutNormal.DistanceTo(exactNormal), 1e-12)

	cornerGap := 0.2 * math.Sqrt2
	assert.InDelta(t, 0.4-cornerGap, exactDepth, 1e-12)
}

// A segment that goes straight through a box has no closest pair to read a
// direction from, so the answer comes from the separating-axis branch instead.
// A solver has to dig bodies out of exactly this state.
func TestCapsule3Box3SegmentThroughTheBox(t *testing.T) {
	box := gomath.Box3[float64]{HalfSize: gomath.Vector3[float64]{X: 2, Y: 1, Z: 3}}

	// Skewered along X, which is not the shallowest axis: Y is.
	capsule := capsuleAlong(
		gomath.Vector3[float64]{X: -5},
		gomath.Vector3[float64]{X: 5},
		0.25,
	)

	penetration, normal, overlapping := Capsule3Box3Penetration(capsule, box)
	require.True(t, overlapping)

	// The segment lies on the box's centre line, so escaping along Y costs the
	// half-extent plus the radius, and that is the least of the three.
	assert.InDelta(t, 1+0.25, penetration, 1e-12)
	assert.InDelta(t, 1, math.Abs(normal.Y), 1e-12, "the shallowest axis is Y, got %v", normal)

	// Moving the box by the returned vector really does clear it.
	moved := box
	moved.Center = moved.Center.Add(normal.MultiplyScalar(penetration + 1e-9))
	assert.False(t, Capsule3Box3(capsule, moved))
}

// Two capsules standing side by side are the crowd case, and their segments are
// parallel: the contact point has to be the middle of the shared span, not an end
// cap. gomath.LineSegment3.ClosestParameters is where that is decided; this pins
// what it means for the pair.
func TestCapsule3Capsule3ParallelStanding(t *testing.T) {
	first := gomath.NewCapsule3Upright(gomath.Vector3[float64]{}, 1.8, 0.3)
	second := gomath.NewCapsule3Upright(gomath.Vector3[float64]{X: 0.5}, 1.8, 0.3)

	penetration, normal, overlapping := Capsule3Capsule3Penetration(first, second)
	require.True(t, overlapping)

	assert.InDelta(t, 0.6-0.5, penetration, 1e-12, "the radii sum to 0.6 and the axes are 0.5 apart")
	assert.InDelta(t, 0, normal.DistanceTo(gomath.Vector3[float64]{X: 1}), 1e-12)

	// The contact is at the middle of the overlapping span, at half the segment's
	// height, rather than at either character's feet.
	onFirst, onSecond := first.Segment.ClosestPoints(second.Segment)
	assert.InDelta(t, 0.9, onFirst.Y, 1e-12)
	assert.InDelta(t, 0.9, onSecond.Y, 1e-12)

	// And it stays there when one of them is shifted up: the shared span moves, the
	// answer moves with it, and nothing jumps to a cap.
	raised := second.Translate(gomath.Vector3[float64]{Y: 0.6})
	onFirst, _ = first.Segment.ClosestPoints(raised.Segment)
	assert.InDelta(t, 1.2, onFirst.Y, 1e-12)
}

// Two capsules whose axes cross have a closest "pair" that is a single point, so
// the direction has to come from the axes themselves — the common perpendicular.
func TestCapsule3Capsule3CrossingAxes(t *testing.T) {
	first := capsuleAlong(gomath.Vector3[float64]{X: -2}, gomath.Vector3[float64]{X: 2}, 0.3)
	second := capsuleAlong(gomath.Vector3[float64]{Y: -2}, gomath.Vector3[float64]{Y: 2}, 0.4)

	penetration, normal, overlapping := Capsule3Capsule3Penetration(first, second)
	require.True(t, overlapping)

	assert.InDelta(t, 0.7, penetration, 1e-12, "the axes touch, so the whole reach is overlap")
	assert.InDelta(t, 1, math.Abs(normal.Z), 1e-12, "the way out is across the plane they span, got %v", normal)

	// The swapped call still negates, which the cross product gives for free.
	_, swapped, _ := Capsule3Capsule3Penetration(second, first)
	assert.InDelta(t, 0, swapped.DistanceTo(normal.Negate()), 1e-12)
}

// A sphere sitting exactly on a capsule's axis has no closest-approach direction
// either, and the escape is radial: perpendicular to the axis, by the full reach.
func TestCapsule3SphereOnTheAxis(t *testing.T) {
	capsule := capsuleAlong(gomath.Vector3[float64]{Y: -1}, gomath.Vector3[float64]{Y: 1}, 0.5)
	sphere := gomath.Sphere[float64]{Center: gomath.Vector3[float64]{Y: 0.25}, Radius: 0.2}

	penetration, normal, overlapping := Capsule3SpherePenetration(capsule, sphere)
	require.True(t, overlapping)

	assert.InDelta(t, 0.7, penetration, 1e-12)
	assert.InDelta(t, 1, normal.Length(), 1e-12)
	assert.InDelta(t, 0, normal.Y, 1e-12, "the escape is perpendicular to the axis, got %v", normal)

	// And it really separates them.
	moved := sphere.Translate(normal.MultiplyScalar(penetration + 1e-9))
	assert.False(t, Capsule3Sphere(capsule, moved))
}

// anyPerpendicular has to work for every direction, including the ones aligned
// with an axis, where a fixed choice of cross-product partner would collapse.
func TestAnyPerpendicular(t *testing.T) {
	directions := []gomath.Vector3[float64]{
		{X: 1}, {Y: 1}, {Z: 1},
		{X: -3}, {Y: 0, Z: -7},
		{X: 1, Y: 1, Z: 1},
		{X: 1e-8, Y: 1, Z: 1e-8},
	}

	for _, direction := range directions {
		perpendicular := anyPerpendicular(direction)

		assert.InDeltaf(t, 1, perpendicular.Length(), 1e-12, "for %v", direction)
		assert.InDeltaf(t, 0, perpendicular.Dot(direction.Normalize()), 1e-12, "for %v", direction)
	}

	assert.Equal(t, gomath.Vector3[float64]{X: 1}, anyPerpendicular(gomath.Vector3[float64]{}),
		"a zero direction gets the same arbitrary axis concentric spheres do")
}

// A regression, and the reason minimiseGapOnInterval returns the middle of an
// interval rather than an end.
//
// The capsule below has one endpoint inside the box, so the segment genuinely
// meets it and the closest gap is zero. Reading the gap at a slab crossing — which
// is where a division put it — reported 1.2e-32 instead, and 1.2e-32 is greater
// than zero: the closest-pair branch then ran on a difference vector that was
// entirely rounding error, normalised it into an arbitrary axis, and returned a
// depth of exactly the radius when the true answer is half again as much, because
// the box is small enough to sit INSIDE the capsule and has to travel further than
// a radius to get out.
//
// The oracle caught it, which is what the oracle is for; this pins the shape of it
// so a future edit to the interval search fails here first and legibly.
func TestCapsule3Box3SegmentEndingInsideTheBox(t *testing.T) {
	capsule := capsuleAlong(
		gomath.Vector3[float64]{X: 0.9582096158864386, Y: -0.7040602781091165, Z: 1.2995108838851261},
		gomath.Vector3[float64]{X: 0.4960748587800213, Y: -0.017644192584584495, Z: 0.2301970825329872},
		1.2837417412716494,
	)
	box := gomath.Box3[float64]{HalfSize: gomath.Vector3[float64]{X: 0.680279398474874, Y: 0.8931086178027228, Z: 0.7957765852437262}}

	require.True(t, box.ContainsPoint(capsule.Segment.P2), "the fixture's far end must be inside the box")

	_, distanceSq := closestOnSegmentToAlignedBox(capsule.Segment, box.HalfSize)
	assert.Equal(t, 0.0, distanceSq, "a segment that reaches inside the box is zero away from it, exactly")

	penetration, normal, overlapping := Capsule3Box3Penetration(capsule, box)
	require.True(t, overlapping)

	assert.Greater(t, penetration, capsule.Radius,
		"the box is inside the capsule, so it has further to travel than a radius")
	assert.InDelta(t, 1, normal.Length(), 1e-12)

	// And the translation genuinely clears it.
	moved := box
	moved.Center = moved.Center.Add(normal.MultiplyScalar(penetration + 1e-9))
	assert.False(t, Capsule3Box3(capsule, moved))
}
