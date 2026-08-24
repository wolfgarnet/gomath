package collider

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wolfgarnet/gomath"
)

// The ground-picking case, which is what all of this is for: a camera above the
// world, a ray through the cursor, and the point on the map the player means.
//
// Both answers are arithmetic done here rather than taken from the function.
func TestRay3Plane3IsGroundPicking(t *testing.T) {
	ground := gomath.NewPlane3FromPointNormal(
		gomath.Vector3[float64]{}, gomath.Vector3[float64]{Y: 1})

	// Straight down from ten units up.
	straight := gomath.NewRay3(
		gomath.Vector3[float64]{Y: 10}, gomath.Vector3[float64]{Y: -1})

	distance, hit := Ray3Plane3(straight, ground)
	require.True(t, hit)
	assert.InDelta(t, 10, distance, 1e-12, "ten units up is ten units away")
	assert.InDelta(t, 0, straight.At(distance).Y, 1e-12, "the hit is on the plane")

	// At 45 degrees the hypotenuse is 10*sqrt(2), landing ten units out.
	angled := gomath.NewRay3(
		gomath.Vector3[float64]{Y: 10}, gomath.Vector3[float64]{X: 1, Y: -1})

	distance, hit = Ray3Plane3(angled, ground)
	require.True(t, hit)
	assert.InDelta(t, 10*math.Sqrt2, distance, 1e-12)

	where := angled.At(distance)
	assert.InDelta(t, 10, where.X, 1e-12, "45 degrees from 10 up lands 10 out")
	assert.InDelta(t, 0, where.Y, 1e-12)
}

// A ray pointing away from the plane meets it only behind its origin, and a ray
// is a half-line. Reporting that as a hit is what makes a camera pick the ground
// behind itself when the player looks at the sky.
func TestRay3Plane3RejectsHitsBehindTheOrigin(t *testing.T) {
	ground := gomath.NewPlane3FromPointNormal(
		gomath.Vector3[float64]{}, gomath.Vector3[float64]{Y: 1})

	skywards := gomath.NewRay3(
		gomath.Vector3[float64]{Y: 10}, gomath.Vector3[float64]{Y: 1})

	_, hit := Ray3Plane3(skywards, ground)
	assert.False(t, hit, "looking up must not pick the ground behind the camera")

	// The flipped ray does hit, which is how a caller asks the other question.
	distance, hit := Ray3Plane3(skywards.Flipped(), ground)
	require.True(t, hit)
	assert.InDelta(t, 10, distance, 1e-12)
}

func TestRay3Plane3ParallelAndOnPlane(t *testing.T) {
	ground := gomath.NewPlane3FromPointNormal(
		gomath.Vector3[float64]{}, gomath.Vector3[float64]{Y: 1})

	// Parallel and above: never meets.
	parallel := gomath.NewRay3(
		gomath.Vector3[float64]{Y: 5}, gomath.Vector3[float64]{X: 1})
	_, hit := Ray3Plane3(parallel, ground)
	assert.False(t, hit, "a ray parallel to the plane never reaches it")

	// Lying IN the plane: its origin is a point of the plane, so distance 0.
	within := gomath.NewRay3(gomath.Vector3[float64]{}, gomath.Vector3[float64]{X: 1})
	distance, hit := Ray3Plane3(within, ground)
	assert.True(t, hit, "a ray lying in the plane starts on it")
	assert.InDelta(t, 0, distance, 1e-12)

	// A degenerate ray picks nothing rather than everything.
	_, hit = Ray3Plane3(gomath.Ray3[float64]{Origin: gomath.Vector3[float64]{Y: 5}}, ground)
	assert.False(t, hit)
}

// The distance to a sphere straight ahead is the gap between them, which needs
// no trigonometry to predict.
func TestRay3SphereDistanceIsTheGap(t *testing.T) {
	sphere := gomath.Sphere[float64]{
		Center: gomath.Vector3[float64]{Z: -10},
		Radius: 2,
	}
	ray := gomath.NewRay3(gomath.Vector3[float64]{}, gomath.Vector3[float64]{Z: -1})

	distance, hit := Ray3Sphere(ray, sphere)
	require.True(t, hit)
	assert.InDelta(t, 8, distance, 1e-12, "ten away, radius two, so the surface is at eight")

	// And the hit point is on the surface.
	assert.InDelta(t, sphere.Radius, ray.At(distance).Sub(sphere.Center).Length(), 1e-12)
}

func TestRay3SphereMissesAndEdgeCases(t *testing.T) {
	sphere := gomath.Sphere[float64]{Center: gomath.Vector3[float64]{Z: -10}, Radius: 2}

	// Passing beside it.
	beside := gomath.NewRay3(gomath.Vector3[float64]{X: 5}, gomath.Vector3[float64]{Z: -1})
	_, hit := Ray3Sphere(beside, sphere)
	assert.False(t, hit)

	// Pointing away from it.
	away := gomath.NewRay3(gomath.Vector3[float64]{}, gomath.Vector3[float64]{Z: 1})
	_, hit = Ray3Sphere(away, sphere)
	assert.False(t, hit, "a sphere behind the origin is not hit")

	// Starting inside reports 0, so a camera inside a bounding sphere does not
	// pick its far wall.
	inside := gomath.NewRay3(sphere.Center, gomath.Vector3[float64]{X: 1})
	distance, hit := Ray3Sphere(inside, sphere)
	require.True(t, hit)
	assert.InDelta(t, 0, distance, 1e-12)

	// Grazing exactly at the radius still counts.
	grazing := gomath.NewRay3(
		gomath.Vector3[float64]{X: 2}, gomath.Vector3[float64]{Z: -1})
	distance, hit = Ray3Sphere(grazing, sphere)
	assert.True(t, hit, "a tangent ray touches the sphere")
	assert.InDelta(t, 10, distance, 1e-6)
}

func TestRay3Box3EntersAtTheNearFace(t *testing.T) {
	box := gomath.Box3[float64]{
		Center:   gomath.Vector3[float64]{Z: -10},
		HalfSize: gomath.Vector3[float64]{X: 1, Y: 1, Z: 1},
	}
	ray := gomath.NewRay3(gomath.Vector3[float64]{}, gomath.Vector3[float64]{Z: -1})

	distance, hit := Ray3Box3(ray, box)
	require.True(t, hit)
	assert.InDelta(t, 9, distance, 1e-12, "the near face is one unit nearer than the centre")
}

func TestRay3Box3MissesAndEdgeCases(t *testing.T) {
	box := gomath.Box3[float64]{
		Center:   gomath.Vector3[float64]{},
		HalfSize: gomath.Vector3[float64]{X: 1, Y: 1, Z: 1},
	}

	// Beside it, parallel to an axis — the case the parallel-slab branch handles.
	beside := gomath.NewRay3(
		gomath.Vector3[float64]{X: 5, Y: 0, Z: -10}, gomath.Vector3[float64]{Z: 1})
	_, hit := Ray3Box3(beside, box)
	assert.False(t, hit, "a ray parallel to an axis and outside the slab must miss")

	// Through the middle, parallel to an axis: inside two slabs, crossing one.
	through := gomath.NewRay3(
		gomath.Vector3[float64]{Z: -10}, gomath.Vector3[float64]{Z: 1})
	distance, hit := Ray3Box3(through, box)
	require.True(t, hit)
	assert.InDelta(t, 9, distance, 1e-12)

	// Behind.
	behind := gomath.NewRay3(
		gomath.Vector3[float64]{Z: -10}, gomath.Vector3[float64]{Z: -1})
	_, hit = Ray3Box3(behind, box)
	assert.False(t, hit)

	// Starting inside.
	inside := gomath.NewRay3(gomath.Vector3[float64]{}, gomath.Vector3[float64]{X: 1})
	distance, hit = Ray3Box3(inside, box)
	require.True(t, hit)
	assert.InDelta(t, 0, distance, 1e-12)

	// A diagonal that passes the corner region but misses the box, which the
	// slab test only rejects because it intersects the intervals rather than
	// testing them independently.
	corner := gomath.NewRay3(
		gomath.Vector3[float64]{X: -5, Y: 1.5, Z: 0}, gomath.Vector3[float64]{X: 1})
	_, hit = Ray3Box3(corner, box)
	assert.False(t, hit, "passing above the box is a miss")
}

// An oriented box with no rotation must agree exactly with the axis-aligned
// test, or the two have drifted apart.
func TestRay3OBB3AgreesWithBox3WhenUnrotated(t *testing.T) {
	box := gomath.Box3[float64]{
		Center:   gomath.Vector3[float64]{X: 1, Y: 2, Z: -6},
		HalfSize: gomath.Vector3[float64]{X: 1, Y: 2, Z: 3},
	}
	oriented := gomath.NewOBB3FromBox3(box)

	random := rand.New(rand.NewSource(7))
	for range 200 {
		ray := randomRay(random)

		alignedDistance, alignedHit := Ray3Box3(ray, box)
		orientedDistance, orientedHit := Ray3OBB3(ray, oriented)

		require.Equal(t, alignedHit, orientedHit, "ray %v disagreed on hit", ray)
		if alignedHit {
			assert.InDelta(t, alignedDistance, orientedDistance, 1e-9)
		}
	}
}

// The discriminating case for orientation: a ray that misses the box's
// axis-aligned extent but hits the box itself once it is turned. If the rotation
// were ignored, this would report a miss.
func TestRay3OBB3RespectsTheRotation(t *testing.T) {
	// A long thin box turned 45 degrees about Y. Its long axis then runs from
	// about (-2.83, 0, 2.83) to (2.83, 0, -2.83).
	halfSize := gomath.Vector3[float64]{X: 4, Y: 0.5, Z: 0.5}
	rotation := gomath.QuaternionFromAxisAngle(
		gomath.Vector3[float64]{Y: 1}, math.Pi/4).ToMatrix3()

	oriented := gomath.NewOBB3(gomath.Vector3[float64]{}, halfSize, rotation)
	unrotated := gomath.NewOBB3(gomath.Vector3[float64]{}, halfSize, gomath.Identity3[float64]())

	// Straight down onto a point that the TURNED box covers and the unturned one
	// does not: (2.5, ·, -2.5) is 3.54 along the rotated long axis — inside its
	// 4 — but its z of -2.5 is far outside the unrotated box's half-extent of
	// 0.5. That difference is the whole point, and it is what a rotation-ignoring
	// implementation gets wrong.
	overhead := gomath.NewRay3(
		gomath.Vector3[float64]{X: 2.5, Y: 10, Z: -2.5},
		gomath.Vector3[float64]{Y: -1})

	distance, hit := Ray3OBB3(overhead, oriented)
	require.True(t, hit, "the turned box lies under this ray")
	assert.InDelta(t, 9.5, distance, 1e-9, "it enters through the top face at y = 0.5")

	// The hit is on the surface: in the box's own frame the vertical coordinate
	// is exactly a half-extent.
	local := overhead.At(distance)
	assert.InDelta(t, 0.5, math.Abs(local.Dot(oriented.Axis(1))), 1e-9)

	// The control. Without it, an implementation that ignored the rotation
	// entirely could still pass the assertion above.
	_, unrotatedHit := Ray3OBB3(overhead, unrotated)
	require.False(t, unrotatedHit,
		"setup: the UNrotated box must not be under this ray, or the test proves nothing")
}

// A randomised cross-check against marching along the ray.
//
// The analytic tests above each pin one predicted number; this pins the SHAPE of
// the answer over many configurations, by finding the first sampled point inside
// the volume and requiring the reported distance to agree. It is the same style
// as the ellipsoid oracle already here.
func TestRay3SphereAgreesWithMarching(t *testing.T) {
	random := rand.New(rand.NewSource(11))

	const (
		samples = 4000
		reach   = 30.0
		step    = reach / samples
	)

	checked := 0
	for range 400 {
		sphere := gomath.Sphere[float64]{
			Center: randomPoint(random, 8),
			Radius: 0.5 + random.Float64()*3,
		}
		// Aimed near the sphere rather than at random, so a useful fraction
		// actually hit. A purely random ray misses almost always, and an oracle
		// that never finds a hit confirms nothing.
		ray := randomRayNear(random, sphere.Center, sphere.Radius*1.6)

		distance, hit := Ray3Sphere(ray, sphere)

		// March until the first point inside.
		marched := -1.0
		for sample := range samples {
			at := float64(sample) * step
			if ray.At(at).Sub(sphere.Center).Length() <= sphere.Radius {
				marched = at

				break
			}
		}

		if marched < 0 {
			// Marching found nothing. The analytic test may still find a graze
			// between samples, so only a confident miss is required.
			if hit && distance < reach-step {
				require.InDelta(t, marched, distance, step*2,
					"reported a hit at %v that marching never found", distance)
			}

			continue
		}

		require.True(t, hit, "marching hit at %v but the test reported a miss", marched)
		// Marching overshoots by up to one step, never undershoots.
		assert.LessOrEqual(t, distance, marched+step)
		assert.GreaterOrEqual(t, distance, marched-step)
		checked++
	}

	require.Positive(t, checked, "the oracle never found a hit, so it proved nothing")
	t.Logf("cross-checked %d ray/sphere hits against marching", checked)
}

func TestRay3BoxAgreesWithMarching(t *testing.T) {
	random := rand.New(rand.NewSource(13))

	const (
		samples = 4000
		reach   = 30.0
		step    = reach / samples
	)

	checked := 0
	for range 400 {
		box := gomath.Box3[float64]{
			Center: randomPoint(random, 8),
			HalfSize: gomath.Vector3[float64]{
				X: 0.5 + random.Float64()*2,
				Y: 0.5 + random.Float64()*2,
				Z: 0.5 + random.Float64()*2,
			},
		}
		ray := randomRayNear(random, box.Center, box.HalfSize.Length()*1.6)

		distance, hit := Ray3Box3(ray, box)

		marched := -1.0
		for sample := range samples {
			at := float64(sample) * step
			if box.ContainsPoint(ray.At(at)) {
				marched = at

				break
			}
		}

		if marched < 0 {
			continue
		}

		require.True(t, hit, "marching entered the box at %v but the test reported a miss", marched)
		assert.LessOrEqual(t, distance, marched+step)
		assert.GreaterOrEqual(t, distance, marched-step)
		checked++
	}

	require.Positive(t, checked, "the oracle never found a hit, so it proved nothing")
	t.Logf("cross-checked %d ray/box hits against marching", checked)
}

func randomPoint(random *rand.Rand, spread float64) gomath.Vector3[float64] {
	return gomath.Vector3[float64]{
		X: (random.Float64()*2 - 1) * spread,
		Y: (random.Float64()*2 - 1) * spread,
		Z: (random.Float64()*2 - 1) * spread,
	}
}

// randomRayNear aims a ray from somewhere outside towards a point scattered
// around a target, giving a mix of hits and near-misses rather than the
// overwhelming misses a uniformly random direction produces.
func randomRayNear(random *rand.Rand, target gomath.Vector3[float64], spread float64) gomath.Ray3[float64] {
	origin := target.Add(randomPoint(random, 1).Normalize().MultiplyScalar(12 + random.Float64()*6))
	aim := target.Add(randomPoint(random, spread))

	ray := gomath.NewRay3Between(origin, aim)
	if !ray.IsValid() {
		return gomath.NewRay3(origin, gomath.Vector3[float64]{Z: -1})
	}

	return ray
}

func randomRay(random *rand.Rand) gomath.Ray3[float64] {
	direction := randomPoint(random, 1)
	if direction.IsZero() {
		direction = gomath.Vector3[float64]{Z: -1}
	}

	return gomath.NewRay3(randomPoint(random, 12), direction)
}

func TestRay3Capsule3(t *testing.T) {
	// A capsule standing on the Y axis: segment from y=-1 to y=1, radius 0.5, so
	// the body spans that stretch and the caps reach y=±1.5.
	capsule := gomath.NewCapsule3(
		gomath.Vector3[float64]{Y: -1},
		gomath.Vector3[float64]{Y: 1},
		0.5,
	)

	tests := []struct {
		name     string
		ray      gomath.Ray3[float64]
		want     float64
		wantHit  bool
		tolerant bool
	}{
		{
			name:    "straight at the body from the side",
			ray:     gomath.NewRay3(gomath.Vector3[float64]{X: -5}, gomath.Vector3[float64]{X: 1}),
			want:    4.5,
			wantHit: true,
		},
		{
			name:    "down the axis into the top cap",
			ray:     gomath.NewRay3(gomath.Vector3[float64]{Y: 5}, gomath.Vector3[float64]{Y: -1}),
			want:    3.5,
			wantHit: true,
		},
		{
			name:    "up the axis into the bottom cap",
			ray:     gomath.NewRay3(gomath.Vector3[float64]{Y: -5}, gomath.Vector3[float64]{Y: 1}),
			want:    3.5,
			wantHit: true,
		},
		{
			name:    "at the side of a cap, above the body",
			ray:     gomath.NewRay3(gomath.Vector3[float64]{X: -5, Y: 1.2}, gomath.Vector3[float64]{X: 1}),
			want:    5 - math.Sqrt(0.25-0.04),
			wantHit: true,
		},
		{
			name:    "starting inside reports zero",
			ray:     gomath.NewRay3(gomath.Vector3[float64]{Y: 0.3}, gomath.Vector3[float64]{X: 1}),
			want:    0,
			wantHit: true,
		},
		{
			name:    "starting inside a cap reports zero",
			ray:     gomath.NewRay3(gomath.Vector3[float64]{Y: 1.3}, gomath.Vector3[float64]{Z: 1}),
			want:    0,
			wantHit: true,
		},
		{
			name:    "pointing away is not a hit",
			ray:     gomath.NewRay3(gomath.Vector3[float64]{X: -5}, gomath.Vector3[float64]{X: -1}),
			wantHit: false,
		},
		{
			name:    "past the end of a cap",
			ray:     gomath.NewRay3(gomath.Vector3[float64]{X: -5, Y: 1.6}, gomath.Vector3[float64]{X: 1}),
			wantHit: false,
		},
		{
			name:    "parallel to the axis and beside it, missing",
			ray:     gomath.NewRay3(gomath.Vector3[float64]{X: 0.7, Y: -5}, gomath.Vector3[float64]{Y: 1}),
			wantHit: false,
		},
		{
			// The case a body-first cascade gets wrong: the ray meets the INFINITE
			// cylinder well above the segment, and the real first surface is the
			// top cap.
			name:    "grazing above the body, caught by the cap",
			ray:     gomath.NewRay3Between(gomath.Vector3[float64]{X: -5, Y: 1.4}, gomath.Vector3[float64]{X: 0, Y: 1.4}),
			want:    5 - math.Sqrt(0.25-0.16),
			wantHit: true,
		},
		{
			name:    "an invalid ray never hits",
			ray:     gomath.Ray3[float64]{Origin: gomath.Vector3[float64]{X: -5}},
			wantHit: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			distance, hit := Ray3Capsule3(test.ray, capsule)

			require.Equal(t, test.wantHit, hit)

			if test.wantHit {
				assert.InDelta(t, test.want, distance, 1e-12)
			} else {
				assert.Zero(t, distance, "a miss must leave the distance at zero")
			}
		})
	}
}

// A degenerate capsule is a sphere, so its ray cast must be Ray3Sphere's — the
// same reduction the penetration functions are held to.
func TestRay3Capsule3ReducesToRay3Sphere(t *testing.T) {
	random := rand.New(rand.NewSource(20260824))

	for trial := range 500 {
		sphere := gomath.Sphere[float64]{Center: randomPoint(random, 4), Radius: 0.5 + random.Float64()*2}
		capsule := gomath.NewCapsule3FromSphere(sphere)
		ray := randomRayNear(random, sphere.Center, sphere.Radius*1.6)

		wantDistance, wantHit := Ray3Sphere(ray, sphere)
		gotDistance, gotHit := Ray3Capsule3(ray, capsule)

		require.Equalf(t, wantHit, gotHit, "trial %d", trial)
		assert.InDeltaf(t, wantDistance, gotDistance, 1e-12, "trial %d", trial)
	}
}

// The same marching oracle the sphere and box casts are held to: step along the
// ray until the first point inside, and require the analytic distance to agree
// within a step.
func TestRay3Capsule3AgainstMarching(t *testing.T) {
	random := rand.New(rand.NewSource(20260824))

	const (
		samples = 4000
		reach   = 30.0
		step    = reach / samples
	)

	checked := 0
	for range 400 {
		centre := randomPoint(random, 8)
		capsule := gomath.NewCapsule3(
			centre.Add(randomPoint(random, 2)),
			centre.Add(randomPoint(random, 2)),
			0.3+random.Float64()*1.5,
		)

		ray := randomRayNear(random, centre, capsule.Radius*2)

		distance, hit := Ray3Capsule3(ray, capsule)

		marched := -1.0
		for sample := range samples {
			at := float64(sample) * step
			if capsule.ContainsPoint(ray.At(at)) {
				marched = at

				break
			}
		}

		if marched < 0 {
			if hit && distance < reach-step {
				require.InDelta(t, marched, distance, step*2,
					"reported a hit at %v that marching never found", distance)
			}

			continue
		}

		require.True(t, hit, "marching hit at %v but the test reported a miss", marched)
		assert.LessOrEqual(t, distance, marched+step)
		assert.GreaterOrEqual(t, distance, marched-step)
		checked++
	}

	require.Positive(t, checked, "the oracle never found a hit, so it proved nothing")
	t.Logf("cross-checked %d ray/capsule hits against marching", checked)
}
