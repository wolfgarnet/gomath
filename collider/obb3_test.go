package collider

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wolfgarnet/gomath"
)

// The optimised SAT test is cross-checked against a deliberately naive oracle
// below. The two share no arithmetic — one works in scalars from a precomputed
// rotation, the other builds cross products as vectors — so agreement over
// random configurations is real evidence rather than the code agreeing with
// itself.

const obbTolerance = 1e-4

// naiveOBB3OBB3 is the direct transcription of the separating-axis theorem:
// build each of the fifteen axes as a vector, project both boxes onto it, and
// report separation if the projections miss.
//
// It normalises and skips degenerate axes, which is what the optimised version
// achieves with an epsilon instead. Written for clarity, not speed, and used
// only as a test oracle.
func naiveOBB3OBB3(a gomath.OBB3[float64], b gomath.OBB3[float64]) bool {
	offset := b.Center.Sub(a.Center)

	axes := make([]gomath.Vector3[float64], 0, 15)
	for i := 0; i < 3; i++ {
		axes = append(axes, a.Axis(i))
	}
	for j := 0; j < 3; j++ {
		axes = append(axes, b.Axis(j))
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			axes = append(axes, a.Axis(i).Cross(b.Axis(j)))
		}
	}

	for _, axis := range axes {
		// A parallel edge pair gives no direction, so it cannot separate.
		if axis.LengthSq() < 1e-12 {
			continue
		}
		unit := axis.Normalize()

		radiusA := math.Abs(unit.Dot(a.Axis(0)))*a.HalfSize.X +
			math.Abs(unit.Dot(a.Axis(1)))*a.HalfSize.Y +
			math.Abs(unit.Dot(a.Axis(2)))*a.HalfSize.Z
		radiusB := math.Abs(unit.Dot(b.Axis(0)))*b.HalfSize.X +
			math.Abs(unit.Dot(b.Axis(1)))*b.HalfSize.Y +
			math.Abs(unit.Dot(b.Axis(2)))*b.HalfSize.Z

		if math.Abs(unit.Dot(offset)) > radiusA+radiusB {
			return false
		}
	}

	return true
}

func unitBox(center gomath.Vector3[float64]) gomath.OBB3[float64] {
	return gomath.NewOBB3FromBox3(gomath.NewBox3(center, gomath.One3[float64]()))
}

func rotatedBox(center gomath.Vector3[float64], axis gomath.Vector3[float64], angle float64, half gomath.Vector3[float64]) gomath.OBB3[float64] {
	return gomath.NewOBB3FromQuaternion(center, half,
		gomath.QuaternionFromAxisAngle(axis.Normalize(), angle))
}

// TestOBB3ParallelBoxesOverlap is the case the naive formulation gets wrong.
//
// With identical orientations every edge cross-product is exactly zero. A test
// written with `>=` on the projection comparison reports these as SEPARATED,
// because zero distance is not less than zero radius — a silent false negative
// that only appears once two boxes rotate into alignment.
func TestOBB3ParallelBoxesOverlap(t *testing.T) {
	cases := []struct {
		name    string
		offset  gomath.Vector3[float64]
		overlap bool
	}{
		{"coincident", gomath.Vector3[float64]{}, true},
		{"overlapping along X", gomath.Vector3[float64]{X: 1.5}, true},
		{"overlapping along Y", gomath.Vector3[float64]{Y: 1.5}, true},
		{"overlapping diagonally", gomath.Vector3[float64]{X: 1, Y: 1, Z: 1}, true},
		{"just touching", gomath.Vector3[float64]{X: 2}, true},
		{"separated along X", gomath.Vector3[float64]{X: 2.5}, false},
		{"separated diagonally", gomath.Vector3[float64]{X: 2.5, Y: 2.5, Z: 2.5}, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			a := unitBox(gomath.Vector3[float64]{})
			b := unitBox(testCase.offset)

			assert.Equal(t, testCase.overlap, OBB3OBB3(a, b), "optimised test")
			assert.Equal(t, testCase.overlap, naiveOBB3OBB3(a, b), "naive oracle")
		})
	}
}

// TestOBB3NearlyParallelBoxesOverlap covers the other half of the same bug: a
// tiny rotation makes the cross products near-zero rather than zero, so they are
// dominated by cancellation error. A test without the epsilon reports sporadic
// separation for boxes that plainly overlap.
func TestOBB3NearlyParallelBoxesOverlap(t *testing.T) {
	for _, angle := range []float64{1e-9, 1e-7, 1e-5, 1e-3, 0.01} {
		a := unitBox(gomath.Vector3[float64]{})
		b := rotatedBox(gomath.Vector3[float64]{X: 0.5},
			gomath.Vector3[float64]{Y: 1}, angle, gomath.One3[float64]())

		assert.True(t, OBB3OBB3(a, b),
			"boxes offset by 0.5 and rotated by %v radians clearly overlap", angle)
	}
}

func TestOBB3RotatedSeparation(t *testing.T) {
	// A box rotated 45 degrees about Y has a longer diagonal in XZ, so it reaches
	// further along X than an axis-aligned one — the case face-normal-only tests
	// get wrong.
	a := unitBox(gomath.Vector3[float64]{})
	rotated := rotatedBox(gomath.Vector3[float64]{X: 2.2},
		gomath.Vector3[float64]{Y: 1}, math.Pi/4, gomath.One3[float64]())

	// Reach along X is 1 + sqrt(2) ~ 2.414, so at 2.2 they overlap.
	assert.True(t, OBB3OBB3(a, rotated))
	assert.True(t, naiveOBB3OBB3(a, rotated))

	// An unrotated box at the same distance does not.
	assert.False(t, OBB3OBB3(a, unitBox(gomath.Vector3[float64]{X: 2.2})))
}

// faceOnlySeparates reports whether any of the six FACE normals separates the
// pair. It deliberately omits the nine edge axes, so comparing it against the
// full test shows what those axes are for.
func faceOnlySeparates(a gomath.OBB3[float64], b gomath.OBB3[float64]) bool {
	offset := b.Center.Sub(a.Center)

	for i := 0; i < 3; i++ {
		for _, axis := range []gomath.Vector3[float64]{a.Axis(i), b.Axis(i)} {
			radiusA := math.Abs(axis.Dot(a.Axis(0)))*a.HalfSize.X +
				math.Abs(axis.Dot(a.Axis(1)))*a.HalfSize.Y +
				math.Abs(axis.Dot(a.Axis(2)))*a.HalfSize.Z
			radiusB := math.Abs(axis.Dot(b.Axis(0)))*b.HalfSize.X +
				math.Abs(axis.Dot(b.Axis(1)))*b.HalfSize.Y +
				math.Abs(axis.Dot(b.Axis(2)))*b.HalfSize.Z

			if math.Abs(axis.Dot(offset)) > radiusA+radiusB {
				return true
			}
		}
	}

	return false
}

// TestEdgeAxesAreNecessary is why the nine cross-product axes exist.
//
// It searches for configurations where every face normal reports overlap but an
// edge axis separates the boxes. A face-normals-only test calls these
// intersecting and is simply wrong — the boxes are disjoint.
//
// The configuration is found by search rather than hand-picked, because these
// cases are surprisingly hard to construct by eye: my first three attempts were
// all separated by a face normal after all, which would have made the test pass
// while proving nothing.
func TestEdgeAxesAreNecessary(t *testing.T) {
	rng := rand.New(rand.NewSource(45))

	randomBox := func(spread float64) gomath.OBB3[float64] {
		axis := gomath.Vector3[float64]{
			X: rng.Float64()*2 - 1,
			Y: rng.Float64()*2 - 1,
			Z: rng.Float64()*2 - 1,
		}
		if axis.IsZero() {
			axis = gomath.Vector3[float64]{Y: 1}
		}

		return gomath.NewOBB3FromQuaternion(
			gomath.Vector3[float64]{
				X: (rng.Float64()*2 - 1) * spread,
				Y: (rng.Float64()*2 - 1) * spread,
				Z: (rng.Float64()*2 - 1) * spread,
			},
			// Elongated boxes, since edge-edge separation needs long thin shapes
			// crossing at an angle.
			gomath.Vector3[float64]{
				X: rng.Float64()*3 + 0.2,
				Y: rng.Float64()*0.4 + 0.1,
				Z: rng.Float64()*0.4 + 0.1,
			},
			gomath.QuaternionFromAxisAngle(axis.Normalize(), (rng.Float64()*2-1)*math.Pi),
		)
	}

	found := 0

	for i := 0; i < 200000 && found < 25; i++ {
		a := randomBox(2)
		b := randomBox(2)

		// The interesting case: faces all overlap, but the boxes are disjoint.
		if faceOnlySeparates(a, b) {
			continue
		}
		if OBB3OBB3(a, b) {
			continue
		}

		found++

		// The full test says separated; the oracle must agree, and a
		// face-only test would have got it wrong.
		assert.False(t, naiveOBB3OBB3(a, b),
			"oracle disagrees on an edge-separated pair:\na=%v\nb=%v", a, b)

		// And the penetration test must also report no overlap.
		_, _, overlapping := OBB3OBB3Penetration(a, b)
		assert.False(t, overlapping,
			"penetration reports an overlap on an edge-separated pair:\na=%v\nb=%v", a, b)
	}

	require.Greater(t, found, 0,
		"found no edge-separated configurations, so this test proves nothing about the edge axes")
	t.Logf("found %d configurations separated only by an edge axis", found)
}

// TestOBB3AgreesWithOracle is the property test: random boxes, both tests must
// give the same answer.
func TestOBB3AgreesWithOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(41))

	randomBox := func() gomath.OBB3[float64] {
		axis := gomath.Vector3[float64]{
			X: rng.Float64()*2 - 1,
			Y: rng.Float64()*2 - 1,
			Z: rng.Float64()*2 - 1,
		}
		if axis.IsZero() {
			axis = gomath.Vector3[float64]{Y: 1}
		}

		return gomath.NewOBB3FromQuaternion(
			gomath.Vector3[float64]{
				X: (rng.Float64()*2 - 1) * 4,
				Y: (rng.Float64()*2 - 1) * 4,
				Z: (rng.Float64()*2 - 1) * 4,
			},
			gomath.Vector3[float64]{
				X: rng.Float64()*1.5 + 0.25,
				Y: rng.Float64()*1.5 + 0.25,
				Z: rng.Float64()*1.5 + 0.25,
			},
			gomath.QuaternionFromAxisAngle(axis.Normalize(), (rng.Float64()*2-1)*math.Pi),
		)
	}

	disagreements := 0
	overlaps := 0

	for i := 0; i < 5000; i++ {
		a := randomBox()
		b := randomBox()

		optimised := OBB3OBB3(a, b)
		oracle := naiveOBB3OBB3(a, b)

		if optimised {
			overlaps++
		}
		if optimised != oracle {
			disagreements++
			if disagreements <= 3 {
				t.Errorf("disagreement %d: optimised=%v oracle=%v\na=%v\nb=%v",
					disagreements, optimised, oracle, a, b)
			}
		}
	}

	assert.Zero(t, disagreements, "the two implementations disagreed %d times", disagreements)

	// A test where nothing ever overlaps proves nothing.
	assert.Greater(t, overlaps, 100, "the random configurations should include overlaps")
	assert.Less(t, overlaps, 4900, "the random configurations should include separations")
}

// TestOBB3MatchesBox3ForAxisAlignedBoxes ties the oriented path to the cheap
// axis-aligned one: with identity rotations they must always agree.
func TestOBB3MatchesBox3ForAxisAlignedBoxes(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < 2000; i++ {
		randomBox := func() gomath.Box3[float64] {
			return gomath.NewBox3(
				gomath.Vector3[float64]{
					X: (rng.Float64()*2 - 1) * 3,
					Y: (rng.Float64()*2 - 1) * 3,
					Z: (rng.Float64()*2 - 1) * 3,
				},
				gomath.Vector3[float64]{
					X: rng.Float64() + 0.25,
					Y: rng.Float64() + 0.25,
					Z: rng.Float64() + 0.25,
				},
			)
		}

		a, b := randomBox(), randomBox()

		assert.Equal(t, Box3Box3(a, b),
			OBB3OBB3(gomath.NewOBB3FromBox3(a), gomath.NewOBB3FromBox3(b)),
			"a=%v b=%v", a, b)
	}
}

func TestOBB3PenetrationSeparatedReportsFalse(t *testing.T) {
	a := unitBox(gomath.Vector3[float64]{})
	b := unitBox(gomath.Vector3[float64]{X: 5})

	penetration, _, overlapping := OBB3OBB3Penetration(a, b)

	assert.False(t, overlapping)
	assert.Zero(t, penetration)
}

func TestOBB3PenetrationDepthAndNormal(t *testing.T) {
	// Two unit boxes overlapping by 0.5 along X: the minimum translation is 0.5
	// along +X, and the normal points from a towards b.
	a := unitBox(gomath.Vector3[float64]{})
	b := unitBox(gomath.Vector3[float64]{X: 1.5})

	penetration, normal, overlapping := OBB3OBB3Penetration(a, b)

	require.True(t, overlapping)
	assert.InDelta(t, 0.5, penetration, obbTolerance)
	assert.InDelta(t, 1, normal.X, obbTolerance)
	assert.InDelta(t, 0, normal.Y, obbTolerance)
	assert.InDelta(t, 0, normal.Z, obbTolerance)
}

// TestOBB3PenetrationNormalPointsFromFirstToSecond pins the package-wide MTV
// convention for the 3D functions too.
func TestOBB3PenetrationNormalPointsFromFirstToSecond(t *testing.T) {
	offsets := []gomath.Vector3[float64]{
		{X: 1.5}, {X: -1.5}, {Y: 1.5}, {Y: -1.5}, {Z: 1.5}, {Z: -1.5},
	}

	for _, offset := range offsets {
		a := unitBox(gomath.Vector3[float64]{})
		b := unitBox(offset)

		_, normal, overlapping := OBB3OBB3Penetration(a, b)
		require.True(t, overlapping, "offset %v should overlap", offset)

		assert.Greater(t, normal.Dot(offset), 0.0,
			"the normal must point from a towards b, got %v for offset %v", normal, offset)
	}
}

// TestOBB3PenetrationResolvesTheOverlap is the property that makes the result
// useful: translating b along the normal by the penetration depth must separate
// the boxes.
func TestOBB3PenetrationResolvesTheOverlap(t *testing.T) {
	rng := rand.New(rand.NewSource(43))

	resolved := 0

	for i := 0; i < 2000; i++ {
		makeBox := func(spread float64) gomath.OBB3[float64] {
			axis := gomath.Vector3[float64]{
				X: rng.Float64()*2 - 1,
				Y: rng.Float64()*2 - 1,
				Z: rng.Float64()*2 - 1,
			}
			if axis.IsZero() {
				axis = gomath.Vector3[float64]{Y: 1}
			}

			return gomath.NewOBB3FromQuaternion(
				gomath.Vector3[float64]{
					X: (rng.Float64()*2 - 1) * spread,
					Y: (rng.Float64()*2 - 1) * spread,
					Z: (rng.Float64()*2 - 1) * spread,
				},
				gomath.Vector3[float64]{
					X: rng.Float64() + 0.4,
					Y: rng.Float64() + 0.4,
					Z: rng.Float64() + 0.4,
				},
				gomath.QuaternionFromAxisAngle(axis.Normalize(), (rng.Float64()*2-1)*math.Pi),
			)
		}

		a := makeBox(1.2)
		b := makeBox(1.2)

		penetration, normal, overlapping := OBB3OBB3Penetration(a, b)
		if !overlapping {
			continue
		}
		resolved++

		// Push past the reported depth by an ABSOLUTE margin, not a relative
		// one. The boolean test is deliberately conservative by about a
		// millionth, so for a penetration of 1e-5 a relative nudge of 0.1%
		// lands well inside that epsilon and the boxes still read as touching.
		pushed := b.Translate(normal.MultiplyScalar(penetration + 1e-4))

		assert.False(t, OBB3OBB3(a, pushed),
			"translating by the reported MTV should separate the boxes: depth %v normal %v",
			penetration, normal)
	}

	assert.Greater(t, resolved, 200, "the sample should contain plenty of overlaps")
}

func TestOBB3PenetrationAgreesWithBooleanTest(t *testing.T) {
	rng := rand.New(rand.NewSource(44))

	for i := 0; i < 3000; i++ {
		makeBox := func() gomath.OBB3[float64] {
			axis := gomath.Vector3[float64]{
				X: rng.Float64()*2 - 1,
				Y: rng.Float64()*2 - 1,
				Z: rng.Float64()*2 - 1,
			}
			if axis.IsZero() {
				axis = gomath.Vector3[float64]{Z: 1}
			}

			return gomath.NewOBB3FromQuaternion(
				gomath.Vector3[float64]{
					X: (rng.Float64()*2 - 1) * 3,
					Y: (rng.Float64()*2 - 1) * 3,
					Z: (rng.Float64()*2 - 1) * 3,
				},
				gomath.Vector3[float64]{
					X: rng.Float64() + 0.3,
					Y: rng.Float64() + 0.3,
					Z: rng.Float64() + 0.3,
				},
				gomath.QuaternionFromAxisAngle(axis.Normalize(), (rng.Float64()*2-1)*math.Pi),
			)
		}

		a, b := makeBox(), makeBox()

		_, _, overlapping := OBB3OBB3Penetration(a, b)

		// The boolean test is slightly conservative because of its epsilon, so it
		// may report an overlap the penetration test does not. The reverse would
		// be a real inconsistency.
		if overlapping {
			assert.True(t, OBB3OBB3(a, b),
				"penetration found an overlap the boolean test missed:\na=%v\nb=%v", a, b)
		}
	}
}

func TestOBB3MixedWithBox3(t *testing.T) {
	box := gomath.NewBox3(gomath.Vector3[float64]{}, gomath.One3[float64]())
	oriented := rotatedBox(gomath.Vector3[float64]{X: 2.2},
		gomath.Vector3[float64]{Y: 1}, math.Pi/4, gomath.One3[float64]())

	assert.True(t, Box3OBB3(box, oriented))
	assert.True(t, OBB3Box3(oriented, box))

	far := rotatedBox(gomath.Vector3[float64]{X: 10},
		gomath.Vector3[float64]{Y: 1}, math.Pi/4, gomath.One3[float64]())
	assert.False(t, Box3OBB3(box, far))

	// The mixed penetration normal follows the same convention.
	_, normal, overlapping := Box3OBB3Penetration(box, oriented)
	require.True(t, overlapping)
	assert.Greater(t, normal.X, 0.0, "the normal should point from the box towards the oriented one")
}

func TestOBB3FromMatrix4FoldsScaleIntoExtents(t *testing.T) {
	// A transform carrying scale must produce the box that scale implies, rather
	// than a basis that quietly violates the orthonormality the tests assume.
	local := gomath.NewBox3(gomath.Vector3[float64]{}, gomath.One3[float64]())
	transform := gomath.TranslationMatrix4[float64](5, 0, 0).
		Mul(gomath.ScaleMatrix4[float64](2, 3, 4))

	oriented, ok := gomath.NewOBB3FromMatrix4(local, transform)
	require.True(t, ok)

	assert.InDelta(t, 5, oriented.Center.X, obbTolerance)
	assert.InDelta(t, 2, oriented.HalfSize.X, obbTolerance)
	assert.InDelta(t, 3, oriented.HalfSize.Y, obbTolerance)
	assert.InDelta(t, 4, oriented.HalfSize.Z, obbTolerance)

	// The basis must be orthonormal after the scale is factored out.
	for i := 0; i < 3; i++ {
		assert.InDelta(t, 1, oriented.Axis(i).Length(), obbTolerance)
	}

	// A collapsed axis has no valid orientation.
	_, ok = gomath.NewOBB3FromMatrix4(local, gomath.ScaleMatrix4[float64](1, 0, 1))
	assert.False(t, ok)
}

// TestOBB3PenetrationWithParallelBoxes exercises the degenerate-axis skip in the
// penetration path — the same configuration that breaks a naive boolean test.
//
// With identical orientations all nine edge axes are exactly zero. They must be
// skipped rather than contributing a zero-depth "minimum", which would otherwise
// be reported as the answer and resolve to no movement at all.
func TestOBB3PenetrationWithParallelBoxes(t *testing.T) {
	a := unitBox(gomath.Vector3[float64]{})
	b := unitBox(gomath.Vector3[float64]{X: 1.5, Y: 0.25})

	penetration, normal, overlapping := OBB3OBB3Penetration(a, b)

	require.True(t, overlapping)

	// Shallowest axis is X, overlapping by 0.5.
	assert.InDelta(t, 0.5, penetration, obbTolerance)
	assert.InDelta(t, 1, normal.X, obbTolerance)
	assert.Greater(t, penetration, 0.0,
		"a zero depth would mean a degenerate axis was accepted as the minimum")

	// And the result still resolves the overlap. The margin has to exceed the
	// boolean test parallelEpsilon, or the push lands exactly on its boundary.
	pushed := b.Translate(normal.MultiplyScalar(penetration + 1e-4))
	assert.False(t, OBB3OBB3(a, pushed))
}
