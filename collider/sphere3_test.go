package collider

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wolfgarnet/gomath"
)

// The sphere tests are checked against point sampling rather than against a
// rearrangement of the same formula.
//
// If a shared point exists, the shapes overlap; that is the definition of overlap
// and it is computed here without touching either overlap function. Sampling can
// only ever prove overlap, never disprove it, so the assertions run in that
// direction: a found shared point demands a true answer, and a false answer demands
// that no sample was ever shared.

func randomSphere(rng *rand.Rand, spread float64, maxRadius float64) gomath.Sphere[float64] {
	return gomath.NewSphere(gomath.Vector3[float64]{
		X: (rng.Float64()*2 - 1) * spread,
		Y: (rng.Float64()*2 - 1) * spread,
		Z: (rng.Float64()*2 - 1) * spread,
	}, rng.Float64()*maxRadius+0.05)
}

func TestSphereSphereAgreesWithPointSampling(t *testing.T) {
	rng := rand.New(rand.NewSource(401))

	overlaps, separate := 0, 0

	for i := 0; i < 4000; i++ {
		a := randomSphere(rng, 3, 2)
		b := randomSphere(rng, 3, 2)

		reported := SphereSphere(a, b)
		if reported {
			overlaps++
		} else {
			separate++
		}

		// Sample the smaller sphere, since a shared point is likeliest to be found
		// there.
		sampled := a
		other := b
		if b.Radius < a.Radius {
			sampled, other = b, a
		}

		shared := false
		for j := 0; j < 400; j++ {
			if point := randomPointIn(rng, sampled); other.ContainsPoint(point) {
				shared = true

				break
			}
		}

		if shared {
			require.True(t, reported,
				"a point lies in both %v and %v, but they were reported separate", a, b)
		}
		if !reported {
			require.False(t, shared)
		}
	}

	// Both answers must appear, or the test is only checking one branch.
	assert.Greater(t, overlaps, 200)
	assert.Greater(t, separate, 200)
}

func TestSphereSphereTouchingCountsAsOverlap(t *testing.T) {
	a := gomath.NewSphere(gomath.Vector3[float64]{}, 1)

	// Exactly tangent along an axis, where the distance is representable exactly.
	assert.True(t, SphereSphere(a, gomath.NewSphere(gomath.Vector3[float64]{X: 3}, 2)),
		"tangent spheres touch, and touching counts")
	assert.False(t, SphereSphere(a, gomath.NewSphere(gomath.Vector3[float64]{X: 3.0001}, 2)))

	// A sphere fully inside another overlaps it.
	assert.True(t, SphereSphere(a, gomath.NewSphere(gomath.Vector3[float64]{}, 0.1)))
	assert.True(t, SphereSphere(gomath.NewSphere(gomath.Vector3[float64]{}, 0.1), a),
		"overlap is symmetric")
}

func TestSphereSpherePenetrationSeparatesExactly(t *testing.T) {
	rng := rand.New(rand.NewSource(402))

	resolved := 0

	for i := 0; i < 3000; i++ {
		a := randomSphere(rng, 2, 2)
		b := randomSphere(rng, 2, 2)

		penetration, normal, overlapping := SphereSpherePenetration(a, b)
		require.Equal(t, SphereSphere(a, b), overlapping,
			"the boolean and the penetration test must agree on %v / %v", a, b)

		if !overlapping {
			assert.Zero(t, penetration)

			continue
		}
		resolved++

		require.InDelta(t, 1, normal.Length(), 1e-12, "the normal must be a unit vector")
		require.GreaterOrEqual(t, penetration, 0.0)

		// Pushing a backwards along the normal by the penetration must leave the
		// two exactly touching: still reported overlapping, but any further nudge
		// separates them. This is the property a solver depends on.
		pushed := a.Translate(normal.MultiplyScalar(-penetration))
		gap := b.Center.Sub(pushed.Center).Length() - (a.Radius + b.Radius)
		require.InDelta(t, 0, gap, 1e-9,
			"after separating by the reported penetration the spheres should be tangent")

		// And the normal must point from a towards b, not the other way, or a solver
		// pushes the bodies together.
		if offset := b.Center.Sub(a.Center); offset.LengthSq() > 1e-12 {
			require.Greater(t, normal.Dot(offset.Normalize()), 0.999,
				"the normal should point from a towards b")
		}
	}

	assert.Greater(t, resolved, 300, "the sample should contain overlapping pairs")
}

func TestSphereSpherePenetrationOfConcentricSpheres(t *testing.T) {
	// Concentric spheres have no separating direction. Returning a zero normal
	// would hand a solver a vector it divides by, so an arbitrary unit axis is
	// used instead and the penetration is the full reach.
	a := gomath.NewSphere(gomath.Vector3[float64]{X: 5}, 1)
	b := gomath.NewSphere(gomath.Vector3[float64]{X: 5}, 2)

	penetration, normal, overlapping := SphereSpherePenetration(a, b)

	require.True(t, overlapping)
	assert.Equal(t, 3.0, penetration)
	assert.InDelta(t, 1, normal.Length(), 1e-12, "the fallback normal must still be unit length")
}

func TestSphereBox3AgreesWithPointSampling(t *testing.T) {
	rng := rand.New(rand.NewSource(403))

	overlaps, separate := 0, 0

	for i := 0; i < 4000; i++ {
		sphere := randomSphere(rng, 3, 1.5)
		box := gomath.NewBox3(
			gomath.Vector3[float64]{
				X: (rng.Float64()*2 - 1) * 3,
				Y: (rng.Float64()*2 - 1) * 3,
				Z: (rng.Float64()*2 - 1) * 3,
			},
			gomath.Vector3[float64]{
				X: rng.Float64()*1.5 + 0.1,
				Y: rng.Float64()*1.5 + 0.1,
				Z: rng.Float64()*1.5 + 0.1,
			})

		reported := SphereBox3(sphere, box)
		if reported {
			overlaps++
		} else {
			separate++
		}

		shared := false
		for j := 0; j < 400; j++ {
			if point := randomPointIn(rng, sphere); box.ContainsPoint(point) {
				shared = true

				break
			}
		}

		if shared {
			require.True(t, reported,
				"a point lies in both %v and %v, but they were reported separate", sphere, box)
		}
		if !reported {
			require.False(t, shared)
		}
	}

	assert.Greater(t, overlaps, 200)
	assert.Greater(t, separate, 200)
}

func TestSphereBox3RejectsNearACorner(t *testing.T) {
	// This is the case that separates an exact test from a cheap axis-wise one. The
	// sphere's centre is 0.5 past the box on every axis, so per-axis the gaps are
	// each smaller than the radius — but the corner is sqrt(3)/2 away, which is
	// further, so there is no overlap.
	box := gomath.NewBox3(gomath.Vector3[float64]{}, gomath.Vector3[float64]{X: 1, Y: 1, Z: 1})
	sphere := gomath.NewSphere(gomath.Vector3[float64]{X: 1.5, Y: 1.5, Z: 1.5}, 0.8)

	require.Less(t, 0.5, sphere.Radius, "setup: each axis gap is inside the radius")
	require.Greater(t, math.Sqrt(3)*0.5, sphere.Radius, "setup: the corner is outside the radius")

	assert.False(t, SphereBox3(sphere, box),
		"a sphere past a corner does not overlap, even though every axis overlaps")

	// Growing the radius past the corner distance must flip the answer, or the
	// assertion above would pass for the wrong reason.
	assert.True(t, SphereBox3(sphere.Expanded(0.2), box))
}

func TestSphereBox3PenetrationSeparates(t *testing.T) {
	rng := rand.New(rand.NewSource(404))

	fromOutside, fromInside := 0, 0

	for i := 0; i < 3000; i++ {
		sphere := randomSphere(rng, 2, 1.5)
		box := gomath.NewBox3(
			gomath.Vector3[float64]{
				X: (rng.Float64()*2 - 1) * 2,
				Y: (rng.Float64()*2 - 1) * 2,
				Z: (rng.Float64()*2 - 1) * 2,
			},
			gomath.Vector3[float64]{X: rng.Float64() + 0.2, Y: rng.Float64() + 0.2, Z: rng.Float64() + 0.2})

		penetration, normal, overlapping := SphereBox3Penetration(sphere, box)
		require.Equal(t, SphereBox3(sphere, box), overlapping,
			"the boolean and the penetration test must agree on %v / %v", sphere, box)

		if !overlapping {
			assert.Zero(t, penetration)

			continue
		}

		require.InDelta(t, 1, normal.Length(), 1e-12, "the normal must be a unit vector")
		require.Greater(t, penetration, 0.0)

		if box.ContainsPoint(sphere.Center) {
			fromInside++
		} else {
			fromOutside++
		}

		// Moving the sphere back along the normal by slightly more than the
		// penetration must separate the pair. Slightly, because the reported
		// distance leaves them exactly touching, which still counts as overlapping.
		pushed := sphere.Translate(normal.MultiplyScalar(-(penetration + 1e-6)))
		require.False(t, SphereBox3(pushed, box),
			"separating %v from %v by %v along %v did not resolve the overlap",
			sphere, box, penetration, normal)
	}

	// Both branches of the function must be exercised: the closest-point case and
	// the shallowest-face fallback for a centre inside the box.
	assert.Greater(t, fromOutside, 200, "the sample should contain centres outside the box")
	assert.Greater(t, fromInside, 20, "the sample should contain centres inside the box")
}

func TestSphereBox3PenetrationNormalPointsAtTheBox(t *testing.T) {
	// A sphere overlapping the +X face of a box centred at the origin: the normal
	// points from the sphere towards the box, so along -X.
	box := gomath.NewBox3(gomath.Vector3[float64]{}, gomath.Vector3[float64]{X: 1, Y: 1, Z: 1})
	sphere := gomath.NewSphere(gomath.Vector3[float64]{X: 1.5}, 0.8)

	penetration, normal, overlapping := SphereBox3Penetration(sphere, box)

	require.True(t, overlapping)
	assert.InDelta(t, 0.3, penetration, 1e-12, "0.8 of radius against a gap of 0.5")
	assert.InDelta(t, -1, normal.X, 1e-12, "got %v", normal)
	assert.InDelta(t, 0, normal.Y, 1e-12)
	assert.InDelta(t, 0, normal.Z, 1e-12)
}

func TestSphereBox3PenetrationFromInsidePicksTheShallowestFace(t *testing.T) {
	// The centre is inside a flat box, nearest the +Y face. There is no closest
	// surface point to push away from, so the shallowest axis decides — and the
	// distance must carry the sphere entirely out, which is the radius plus the
	// remaining overlap on that axis.
	box := gomath.NewBox3(gomath.Vector3[float64]{}, gomath.Vector3[float64]{X: 4, Y: 1, Z: 4})
	sphere := gomath.NewSphere(gomath.Vector3[float64]{Y: 0.6}, 0.5)

	penetration, normal, overlapping := SphereBox3Penetration(sphere, box)

	require.True(t, overlapping)
	// The +Y face is nearest, so resolving moves the sphere towards +Y — which
	// means the normal, being the direction a solver pushes AGAINST, is -Y. Inside
	// the box there is no "towards the box" to speak of, so the fallback is defined
	// as whatever keeps that sign convention continuous with the outside case.
	assert.InDelta(t, -1, normal.Y, 1e-12, "got %v", normal)
	assert.InDelta(t, 0.9, penetration, 1e-12, "0.5 of radius plus 0.4 still inside the face")

	// Applying it must clear the box completely.
	pushed := sphere.Translate(normal.MultiplyScalar(-(penetration + 1e-6)))
	assert.False(t, SphereBox3(pushed, box))
}

func TestSphereOBB3MatchesSphereBox3WhenUnrotated(t *testing.T) {
	// An unrotated OBB is an AABB, so the two functions must agree exactly. That
	// pins the oriented path against the simpler one before trusting it rotated.
	rng := rand.New(rand.NewSource(405))

	agreed := 0

	for i := 0; i < 3000; i++ {
		sphere := randomSphere(rng, 2.5, 1.5)
		halfSize := gomath.Vector3[float64]{X: rng.Float64() + 0.2, Y: rng.Float64() + 0.2, Z: rng.Float64() + 0.2}
		centre := gomath.Vector3[float64]{
			X: (rng.Float64()*2 - 1) * 2,
			Y: (rng.Float64()*2 - 1) * 2,
			Z: (rng.Float64()*2 - 1) * 2,
		}

		box := gomath.NewBox3(centre, halfSize)
		oriented := gomath.NewOBB3FromQuaternion(centre, halfSize,
			gomath.QuaternionFromAxisAngle(gomath.Vector3[float64]{Y: 1}, 0))

		require.Equal(t, SphereBox3(sphere, box), SphereOBB3(sphere, oriented),
			"an unrotated OBB must behave as its AABB for %v", sphere)
		agreed++
	}

	assert.Equal(t, 3000, agreed)
}

func TestSphereOBB3AgreesWithPointSampling(t *testing.T) {
	rng := rand.New(rand.NewSource(406))

	overlaps, separate := 0, 0

	for i := 0; i < 3000; i++ {
		sphere := randomSphere(rng, 2.5, 1.2)

		axis := gomath.Vector3[float64]{
			X: rng.Float64()*2 - 1,
			Y: rng.Float64()*2 - 1,
			Z: rng.Float64()*2 - 1,
		}.Normalize()
		if axis.IsZero() {
			continue
		}

		oriented := gomath.NewOBB3FromQuaternion(
			gomath.Vector3[float64]{
				X: (rng.Float64()*2 - 1) * 2.5,
				Y: (rng.Float64()*2 - 1) * 2.5,
				Z: (rng.Float64()*2 - 1) * 2.5,
			},
			gomath.Vector3[float64]{X: rng.Float64() + 0.2, Y: rng.Float64()*0.6 + 0.1, Z: rng.Float64() + 0.2},
			gomath.QuaternionFromAxisAngle(axis, (rng.Float64()*2-1)*math.Pi))

		reported := SphereOBB3(sphere, oriented)
		if reported {
			overlaps++
		} else {
			separate++
		}

		shared := false
		for j := 0; j < 400; j++ {
			if point := randomPointIn(rng, sphere); oriented.ContainsPoint(point) {
				shared = true

				break
			}
		}

		if shared {
			require.True(t, reported,
				"a point lies in both %v and the oriented box, but they were reported separate", sphere)
		}
		if !reported {
			require.False(t, shared)
		}
	}

	assert.Greater(t, overlaps, 150)
	assert.Greater(t, separate, 150)
}

func TestSphereTestsAreCheaperThanBoxTests(t *testing.T) {
	// The whole argument for a bounding sphere is that rejecting with one is
	// cheaper than rejecting with a box. If that is not true here, the type has no
	// reason to be the first link in a cull chain.
	rng := rand.New(rand.NewSource(407))

	spheres := make([]gomath.Sphere[float64], 512)
	boxes := make([]gomath.OBB3[float64], 512)
	for i := range spheres {
		spheres[i] = randomSphere(rng, 20, 2)

		axis := gomath.Vector3[float64]{X: rng.Float64(), Y: rng.Float64(), Z: rng.Float64()}.Normalize()
		boxes[i] = gomath.NewOBB3FromQuaternion(spheres[i].Center,
			gomath.Vector3[float64]{X: 1, Y: 1, Z: 1},
			gomath.QuaternionFromAxisAngle(axis, rng.Float64()))
	}

	sphereResult := testing.Benchmark(func(b *testing.B) {
		count := 0
		for i := 0; i < b.N; i++ {
			if SphereSphere(spheres[i%len(spheres)], spheres[(i*7+3)%len(spheres)]) {
				count++
			}
		}
		_ = count
	})

	boxResult := testing.Benchmark(func(b *testing.B) {
		count := 0
		for i := 0; i < b.N; i++ {
			if OBB3OBB3(boxes[i%len(boxes)], boxes[(i*7+3)%len(boxes)]) {
				count++
			}
		}
		_ = count
	})

	t.Logf("sphere/sphere %v, obb/obb %v", sphereResult.NsPerOp(), boxResult.NsPerOp())

	assert.Less(t, sphereResult.NsPerOp(), boxResult.NsPerOp(),
		"a sphere test should be cheaper than an oriented box test")
}

// randomPointIn returns a point uniformly inside the sphere, by rejection against
// the bounding cube — unbiased, unlike scaling a random direction by a random
// radius.
func randomPointIn(rng *rand.Rand, sphere gomath.Sphere[float64]) gomath.Vector3[float64] {
	for {
		candidate := gomath.Vector3[float64]{
			X: (rng.Float64()*2 - 1) * sphere.Radius,
			Y: (rng.Float64()*2 - 1) * sphere.Radius,
			Z: (rng.Float64()*2 - 1) * sphere.Radius,
		}
		if candidate.LengthSq() <= sphere.Radius*sphere.Radius {
			return sphere.Center.Add(candidate)
		}
	}
}
