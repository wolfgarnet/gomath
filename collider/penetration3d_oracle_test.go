package collider

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wolfgarnet/gomath"
)

// A brute-force oracle for the 3D penetration functions, in the same spirit as
// ellipsoid_oracle_test.go's for 2D ellipse distance: an independent, deliberately
// naive computation of the answer, checked against the analytic one over randomly
// generated configurations.
//
// # What the oracle actually computes
//
// For two convex shapes the penetration depth is
//
//	min over unit directions n of w(n)
//
// where w(n) is the support function of the Minkowski difference, evaluated on
// n. That needs nothing but a support function per shape — no closest points, no
// separating axes, none of the machinery the analytic functions use — which is
// what makes it independent evidence rather than the same code written twice.
//
// # Why this pins the answer from both sides
//
// A finite sample of directions can only ever OVERSTATE the minimum, so
// `analytic <= sampled` is a bound that holds without any tolerance and catches a
// depth that is too large. The other side comes from self-consistency: the
// returned depth must equal w(n) for the returned normal n, and w of anything is
// at least the true minimum, so that catches a depth that is too small. Together
// they trap the analytic answer, which neither check does alone.
//
// # float64 here, float32 in the consistency harness
//
// The oracle's job is exact arithmetic, so it runs the generic functions at
// float64 to leave headroom for the sampling to be the dominant error rather than
// the type. penetration3d_consistency_test.go covers the consumer's float32
// instantiation.

// supportFunc gives the extent of a shape's projection onto a unit direction.
type supportFunc func(direction gomath.Vector3[float64]) (lo float64, hi float64)

func sphereSupport(sphere gomath.Sphere[float64]) supportFunc {
	return func(direction gomath.Vector3[float64]) (float64, float64) {
		centre := sphere.Center.Dot(direction)

		return centre - sphere.Radius, centre + sphere.Radius
	}
}

func box3Support(box gomath.Box3[float64]) supportFunc {
	return func(direction gomath.Vector3[float64]) (float64, float64) {
		centre := box.Center.Dot(direction)
		reach := box.HalfSize.X*math.Abs(direction.X) +
			box.HalfSize.Y*math.Abs(direction.Y) +
			box.HalfSize.Z*math.Abs(direction.Z)

		return centre - reach, centre + reach
	}
}

// A capsule's support is its segment's, widened by the radius on both sides — a
// Minkowski sum with a ball adds the radius along every direction at once. That
// makes the capsule rows independent evidence in the same sense as the others:
// two lines that know nothing of segments, parameters or separating axes.
func capsule3Support(capsule gomath.Capsule3[float64]) supportFunc {
	return func(direction gomath.Vector3[float64]) (float64, float64) {
		first := capsule.Segment.P1.Dot(direction)
		second := capsule.Segment.P2.Dot(direction)

		return math.Min(first, second) - capsule.Radius, math.Max(first, second) + capsule.Radius
	}
}

func obb3Support(box gomath.OBB3[float64]) supportFunc {
	return func(direction gomath.Vector3[float64]) (float64, float64) {
		centre := box.Center.Dot(direction)
		reach := box.HalfSize.X*math.Abs(direction.Dot(box.Axis(0))) +
			box.HalfSize.Y*math.Abs(direction.Dot(box.Axis(1))) +
			box.HalfSize.Z*math.Abs(direction.Dot(box.Axis(2)))

		return centre - reach, centre + reach
	}
}

// depthAlong is how far the second shape would have to travel along direction to
// clear the first, negative when the direction already separates them.
//
// It is the support function of the Minkowski difference: h(n) = h_first(n) +
// h_second(-n), which is hi_first(n) - lo_second(n) in the terms above.
//
// Note what this is NOT. The obvious thing to write is the length of the
// intersection of the two projected intervals — the quantity SAT uses to decide
// separation. That is the same number in the ordinary case, and it is WRONG when
// one interval contains the other: a small sphere deep inside a large one gives
// an interval intersection of the small sphere's diameter, when the distance it
// actually has to travel to escape is very much further. Only the support-function
// form is the penetration depth.
func depthAlong(first supportFunc, second supportFunc, direction gomath.Vector3[float64]) float64 {
	_, firstHi := first(direction)
	secondLo, _ := second(direction)

	return firstHi - secondLo
}

// fibonacciDirections returns count roughly-uniform unit vectors on the sphere.
//
// The spiral is used rather than random sampling so the coarse pass has no gaps
// to get lucky or unlucky about, and rather than a lat/long grid because that
// clusters samples at the poles and would resolve some axes far better than
// others.
func fibonacciDirections(count int) []gomath.Vector3[float64] {
	directions := make([]gomath.Vector3[float64], 0, count)
	goldenAngle := math.Pi * (3 - math.Sqrt(5))

	for index := 0; index < count; index++ {
		z := 1 - 2*(float64(index)+0.5)/float64(count)
		ringRadius := math.Sqrt(math.Max(0, 1-z*z))
		theta := goldenAngle * float64(index)

		directions = append(directions, gomath.Vector3[float64]{
			X: ringRadius * math.Cos(theta),
			Y: ringRadius * math.Sin(theta),
			Z: z,
		})
	}

	return directions
}

// minimumDepth is the oracle proper: a coarse sweep over the sphere of directions
// followed by a stochastic descent from the best one.
//
// The refinement matters. depthAlong has a CORNER at its minimum for box-shaped
// operands, not a smooth basin, so the coarse sweep's error falls off only
// linearly in the sample spacing — an unrefined sweep overstates the minimum by
// enough that the `analytic <= sampled` bound would pass for answers that are
// visibly wrong. Descending afterwards makes the bound tight enough to be worth
// asserting.
func minimumDepth(first supportFunc, second supportFunc, random *rand.Rand) float64 {
	best := gomath.Vector3[float64]{X: 1}
	bestDepth := math.Inf(1)

	for _, direction := range fibonacciDirections(1500) {
		if depth := depthAlong(first, second, direction); depth < bestDepth {
			best, bestDepth = direction, depth
		}
	}

	searchRadius := 0.2
	for step := 0; step < 3000; step++ {
		candidate := best.Add(gomath.Vector3[float64]{
			X: random.NormFloat64() * searchRadius,
			Y: random.NormFloat64() * searchRadius,
			Z: random.NormFloat64() * searchRadius,
		}).Normalize()

		if depth := depthAlong(first, second, candidate); depth < bestDepth {
			best, bestDepth = candidate, depth
		}

		if step%150 == 149 {
			searchRadius *= 0.75
		}
	}

	return bestDepth
}

// oracleTrial is one randomly generated configuration: the analytic answer, and
// the two support functions the oracle re-derives it from.
type oracleTrial struct {
	penetration float64
	normal      gomath.Vector3[float64]
	overlapping bool

	first  supportFunc
	second supportFunc
}

// randomHalfSize keeps every extent comfortably away from zero. A degenerate
// extent makes a box a plane or a line, where the minimum-overlap direction stops
// being unique and the oracle has nothing definite to compare against.
func randomHalfSize(random *rand.Rand) gomath.Vector3[float64] {
	return gomath.Vector3[float64]{
		X: 0.3 + random.Float64(),
		Y: 0.3 + random.Float64(),
		Z: 0.3 + random.Float64(),
	}
}

func randomRotation(random *rand.Rand) gomath.Quaternion[float64] {
	return gomath.QuaternionFromEulerXYZ(
		random.Float64()*2*math.Pi,
		random.Float64()*2*math.Pi,
		random.Float64()*2*math.Pi,
	)
}

func randomBox3(random *rand.Rand, spread float64) gomath.Box3[float64] {
	return gomath.Box3[float64]{
		Center:   randomPoint(random, spread),
		HalfSize: randomHalfSize(random),
	}
}

func randomOBB3(random *rand.Rand, spread float64) gomath.OBB3[float64] {
	return gomath.NewOBB3FromQuaternion(randomPoint(random, spread), randomHalfSize(random), randomRotation(random))
}

// The spread is deliberately small relative to the shapes' extents, so that a
// large fraction of the random configurations actually overlap. A wider spread
// would spend the trial budget confirming that distant shapes miss each other,
// which is the case least likely to be wrong.
const (
	oracleSpread = 1.2

	// oracleMaxRadius matches randomHalfSize's range, so spheres and boxes are
	// comparable in size and neither routinely swallows the other.
	oracleMaxRadius = 1.3
)

func TestPenetration3D_AgainstBruteForceOracle(t *testing.T) {
	const (
		trials = 120

		// Configurations whose minimum overlap sits inside this band of zero are
		// skipped: the oracle's descent cannot reliably tell a hair of overlap
		// from a hair of separation, and neither verdict would be evidence.
		tangentBand = 1e-6

		// The analytic and oracle depths are computed by completely different
		// routes, so they agree to sampling error rather than to float64 epsilon.
		depthTolerance = 1e-6
	)

	cases := []struct {
		name string
		// depthSlack is how far above the true minimum the analytic answer is
		// permitted to sit. It is 1 for the exact pairs and slightly above 1 for
		// anything routed through OBB3OBB3Penetration, which deliberately prefers
		// a face axis over an edge axis unless the edge beats it by
		// faceAxisPreference — a stability choice, paid for in exactness.
		depthSlack float64
		build      func(random *rand.Rand) oracleTrial
	}{
		{
			name:       "SphereSpherePenetration",
			depthSlack: 1,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomSphere(random, oracleSpread, oracleMaxRadius), randomSphere(random, oracleSpread, oracleMaxRadius)
				penetration, normal, overlapping := SphereSpherePenetration(first, second)

				return oracleTrial{penetration, normal, overlapping, sphereSupport(first), sphereSupport(second)}
			},
		},
		{
			name:       "SphereBox3Penetration",
			depthSlack: 1,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomSphere(random, oracleSpread, oracleMaxRadius), randomBox3(random, oracleSpread)
				penetration, normal, overlapping := SphereBox3Penetration(first, second)

				return oracleTrial{penetration, normal, overlapping, sphereSupport(first), box3Support(second)}
			},
		},
		{
			name:       "Box3SpherePenetration",
			depthSlack: 1,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomBox3(random, oracleSpread), randomSphere(random, oracleSpread, oracleMaxRadius)
				penetration, normal, overlapping := Box3SpherePenetration(first, second)

				return oracleTrial{penetration, normal, overlapping, box3Support(first), sphereSupport(second)}
			},
		},
		{
			name:       "SphereOBB3Penetration",
			depthSlack: 1,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomSphere(random, oracleSpread, oracleMaxRadius), randomOBB3(random, oracleSpread)
				penetration, normal, overlapping := SphereOBB3Penetration(first, second)

				return oracleTrial{penetration, normal, overlapping, sphereSupport(first), obb3Support(second)}
			},
		},
		{
			name:       "OBB3SpherePenetration",
			depthSlack: 1,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomOBB3(random, oracleSpread), randomSphere(random, oracleSpread, oracleMaxRadius)
				penetration, normal, overlapping := OBB3SpherePenetration(first, second)

				return oracleTrial{penetration, normal, overlapping, obb3Support(first), sphereSupport(second)}
			},
		},
		{
			name:       "Box3Box3Penetration",
			depthSlack: 1,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomBox3(random, oracleSpread), randomBox3(random, oracleSpread)
				penetration, normal, overlapping := Box3Box3Penetration(first, second)

				return oracleTrial{penetration, normal, overlapping, box3Support(first), box3Support(second)}
			},
		},
		{
			name:       "Box3OBB3Penetration",
			depthSlack: 1 / faceAxisPreference,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomBox3(random, oracleSpread), randomOBB3(random, oracleSpread)
				penetration, normal, overlapping := Box3OBB3Penetration(first, second)

				return oracleTrial{penetration, normal, overlapping, box3Support(first), obb3Support(second)}
			},
		},
		{
			name:       "OBB3Box3Penetration",
			depthSlack: 1 / faceAxisPreference,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomOBB3(random, oracleSpread), randomBox3(random, oracleSpread)
				penetration, normal, overlapping := OBB3Box3Penetration(first, second)

				return oracleTrial{penetration, normal, overlapping, obb3Support(first), box3Support(second)}
			},
		},
		{
			name:       "OBB3OBB3Penetration",
			depthSlack: 1 / faceAxisPreference,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomOBB3(random, oracleSpread), randomOBB3(random, oracleSpread)
				penetration, normal, overlapping := OBB3OBB3Penetration(first, second)

				return oracleTrial{penetration, normal, overlapping, obb3Support(first), obb3Support(second)}
			},
		},
		// Every capsule pair runs at depthSlack 1: none of them buys any slack, the
		// box ones included. See capsuleAlignedBoxPenetration for why the box pairs
		// can be exact rather than the approximation they are usually written as —
		// and note that it is this file, not the consistency harness, that would
		// catch a normal mapped back to world space by the wrong rotation.
		{
			name:       "Capsule3SpherePenetration",
			depthSlack: 1,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomCapsule3(random, oracleSpread, oracleMaxRadius), randomSphere(random, oracleSpread, oracleMaxRadius)
				penetration, normal, overlapping := Capsule3SpherePenetration(first, second)

				return oracleTrial{penetration, normal, overlapping, capsule3Support(first), sphereSupport(second)}
			},
		},
		{
			name:       "SphereCapsule3Penetration",
			depthSlack: 1,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomSphere(random, oracleSpread, oracleMaxRadius), randomCapsule3(random, oracleSpread, oracleMaxRadius)
				penetration, normal, overlapping := SphereCapsule3Penetration(first, second)

				return oracleTrial{penetration, normal, overlapping, sphereSupport(first), capsule3Support(second)}
			},
		},
		{
			name:       "Capsule3Capsule3Penetration",
			depthSlack: 1,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomCapsule3(random, oracleSpread, oracleMaxRadius), randomCapsule3(random, oracleSpread, oracleMaxRadius)
				penetration, normal, overlapping := Capsule3Capsule3Penetration(first, second)

				return oracleTrial{penetration, normal, overlapping, capsule3Support(first), capsule3Support(second)}
			},
		},
		{
			name:       "Capsule3Box3Penetration",
			depthSlack: 1,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomCapsule3(random, oracleSpread, oracleMaxRadius), randomBox3(random, oracleSpread)
				penetration, normal, overlapping := Capsule3Box3Penetration(first, second)

				return oracleTrial{penetration, normal, overlapping, capsule3Support(first), box3Support(second)}
			},
		},
		{
			name:       "Box3Capsule3Penetration",
			depthSlack: 1,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomBox3(random, oracleSpread), randomCapsule3(random, oracleSpread, oracleMaxRadius)
				penetration, normal, overlapping := Box3Capsule3Penetration(first, second)

				return oracleTrial{penetration, normal, overlapping, box3Support(first), capsule3Support(second)}
			},
		},
		{
			name:       "Capsule3OBB3Penetration",
			depthSlack: 1,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomCapsule3(random, oracleSpread, oracleMaxRadius), randomOBB3(random, oracleSpread)
				penetration, normal, overlapping := Capsule3OBB3Penetration(first, second)

				return oracleTrial{penetration, normal, overlapping, capsule3Support(first), obb3Support(second)}
			},
		},
		{
			name:       "OBB3Capsule3Penetration",
			depthSlack: 1,
			build: func(random *rand.Rand) oracleTrial {
				first, second := randomOBB3(random, oracleSpread), randomCapsule3(random, oracleSpread, oracleMaxRadius)
				penetration, normal, overlapping := OBB3Capsule3Penetration(first, second)

				return oracleTrial{penetration, normal, overlapping, obb3Support(first), capsule3Support(second)}
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			// Seeded per case so a failure reproduces and cases do not perturb
			// each other's sequences.
			random := rand.New(rand.NewSource(20260815))

			checked := 0

			for trial := 0; trial < trials; trial++ {
				result := testCase.build(random)
				trueDepth := minimumDepth(result.first, result.second, random)

				if math.Abs(trueDepth) < tangentBand {
					// Too close to call; the oracle has nothing to say.
					continue
				}

				if trueDepth < 0 {
					assert.False(t, result.overlapping,
						"trial %d: oracle separates the pair by %g", trial, -trueDepth)

					continue
				}

				if !assert.True(t, result.overlapping,
					"trial %d: oracle overlaps the pair by %g", trial, trueDepth) {
					continue
				}

				checked++

				assert.InDeltaf(t, 1, result.normal.Length(), depthTolerance,
					"trial %d: normal must be unit length", trial)

				// The returned depth must be the overlap along the returned
				// normal — otherwise the pair of values is not a translation that
				// resolves anything, whatever the depth alone says.
				assert.InDeltaf(t, depthAlong(result.first, result.second, result.normal), result.penetration,
					depthTolerance, "trial %d: depth must equal the overlap along its own normal", trial)

				// And it must be minimal: a finite sample can only overstate the
				// true minimum, so this bound needs no tolerance beyond the slack
				// the implementation openly buys.
				assert.LessOrEqualf(t, result.penetration, trueDepth*testCase.depthSlack+depthTolerance,
					"trial %d: depth exceeds the oracle's minimum of %g", trial, trueDepth)
			}

			// A guard against the whole case quietly degenerating into skips and
			// separations, which would make every assertion above vacuous.
			assert.Greaterf(t, checked, trials/4,
				"only %d of %d trials actually overlapped; the fixtures are not exercising the function", checked, trials)
		})
	}
}
