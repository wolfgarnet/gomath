package gomath

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSphereContainsPoint(t *testing.T) {
	sphere := NewSphere(Vector3[float64]{X: 1, Y: 2, Z: 3}, 2)

	assert.True(t, sphere.ContainsPoint(sphere.Center), "the centre is inside")
	assert.True(t, sphere.ContainsPoint(Vector3[float64]{X: 3, Y: 2, Z: 3}),
		"a point exactly on the surface counts as contained")
	assert.False(t, sphere.ContainsPoint(Vector3[float64]{X: 3.001, Y: 2, Z: 3}))

	// The distinguishing case: a point that is within the radius on every axis
	// separately but outside the ball. An axis-wise test would get this wrong.
	diagonal := sphere.Center.Add(Vector3[float64]{X: 1.5, Y: 1.5, Z: 0})
	require.Less(t, 1.5, sphere.Radius, "setup: each component must be inside the radius")
	assert.False(t, sphere.ContainsPoint(diagonal),
		"a point past the surface diagonally is outside even though no component is")
}

func TestSphereContainsSphere(t *testing.T) {
	outer := NewSphere(Vector3[float64]{}, 3)

	assert.True(t, outer.ContainsSphere(NewSphere(Vector3[float64]{}, 3)),
		"a sphere contains itself")
	assert.True(t, outer.ContainsSphere(NewSphere(Vector3[float64]{X: 2}, 1)),
		"internally tangent counts as contained")
	assert.False(t, outer.ContainsSphere(NewSphere(Vector3[float64]{X: 2.5}, 1)),
		"poking out of the far side is not contained")
	assert.False(t, outer.ContainsSphere(NewSphere(Vector3[float64]{}, 3.5)),
		"a larger sphere is never contained")

	// Containment must agree with point containment: every point of a contained
	// sphere is a point of the container.
	inner := NewSphere(Vector3[float64]{X: 1, Y: -0.5}, 1.2)
	require.True(t, inner.Radius > 0)
	require.True(t, outer.ContainsSphere(inner), "setup")

	rng := rand.New(rand.NewSource(311))
	for i := 0; i < 5000; i++ {
		point := randomPointInSphere(rng, inner)
		require.True(t, outer.ContainsPoint(point),
			"%v is in the contained sphere but not in the container", point)
	}
}

func TestSphereClosestPoint(t *testing.T) {
	sphere := NewSphere(Vector3[float64]{X: 1}, 2)

	// Inside points are returned unchanged, which is what makes the function a
	// projection onto the SOLID ball rather than onto its surface.
	inside := Vector3[float64]{X: 1.5, Y: 0.5}
	assert.Equal(t, inside, sphere.ClosestPoint(inside))

	outside := Vector3[float64]{X: 10}
	closest := sphere.ClosestPoint(outside)
	assert.InDelta(t, 3, closest.X, 1e-12, "got %v", closest)
	assert.InDelta(t, 0, closest.Y, 1e-12, "got %v", closest)
	assert.InDelta(t, 0, closest.Z, 1e-12, "got %v", closest)

	// For any outside point, the result must be on the surface and no other point
	// of the sphere may be nearer. Checking against random surface points is an
	// independent statement of "closest".
	rng := rand.New(rand.NewSource(312))
	for i := 0; i < 2000; i++ {
		query := Vector3[float64]{
			X: (rng.Float64()*2 - 1) * 8,
			Y: (rng.Float64()*2 - 1) * 8,
			Z: (rng.Float64()*2 - 1) * 8,
		}
		if sphere.ContainsPoint(query) {
			continue
		}

		result := sphere.ClosestPoint(query)
		require.InDelta(t, sphere.Radius, result.Sub(sphere.Center).Length(), 1e-12,
			"the closest point to an outside query must be on the surface")

		best := result.Sub(query).Length()
		for j := 0; j < 50; j++ {
			candidate := randomPointInSphere(rng, sphere)
			require.LessOrEqual(t, best, candidate.Sub(query).Length()+1e-12,
				"found a nearer point %v than the reported closest %v", candidate, result)
		}
	}
}

func TestSphereBoundingBoxAndBoxRoundTrip(t *testing.T) {
	sphere := NewSphere(Vector3[float64]{X: 2, Y: -1, Z: 4}, 1.5)

	box := sphere.BoundingBox()
	assert.Equal(t, sphere.Center, box.Center)

	// The bounds must contain the sphere: every surface point is in the box.
	rng := rand.New(rand.NewSource(313))
	for i := 0; i < 3000; i++ {
		require.True(t, box.ContainsPoint(randomPointInSphere(rng, sphere)))
	}

	// And the circumscribing sphere of those bounds must contain the sphere back,
	// larger by the corner-to-face ratio sqrt(3).
	circumscribed := NewSphereFromBox3(box)
	assert.True(t, circumscribed.ContainsSphere(sphere))
	// Bar the rounding slack the circumscribing radius is the box's half-diagonal.
	assert.InEpsilon(t, sphere.Radius*math.Sqrt(3), circumscribed.Radius, 1e-5)
}

func TestNewSphereFromPointsContainsThemAll(t *testing.T) {
	_, ok := NewSphereFromPoints([]Vector3[float64]{})
	assert.False(t, ok, "no points means no sphere")

	rng := rand.New(rand.NewSource(314))

	for trial := 0; trial < 200; trial++ {
		points := make([]Vector3[float64], 1+rng.Intn(40))
		for i := range points {
			points[i] = Vector3[float64]{
				X: (rng.Float64()*2 - 1) * 10,
				Y: (rng.Float64()*2 - 1) * 10,
				Z: (rng.Float64()*2 - 1) * 10,
			}
		}

		sphere, ok := NewSphereFromPoints(points)
		require.True(t, ok)

		// The only contract is containment. It is documented as non-minimal, so
		// tightness is deliberately not asserted.
		for _, point := range points {
			require.True(t, sphere.ContainsPoint(point),
				"point %v was left outside %v", point, sphere)
		}
	}

	// The slack that makes containment unconditional must stay invisibly small:
	// without this, growing it to paper over a real bug would go unnoticed.
	corner := []Vector3[float64]{{X: -1, Y: -1, Z: -1}, {X: 1, Y: 1, Z: 1}}
	tight, ok := NewSphereFromPoints(corner)
	require.True(t, ok)
	assert.InEpsilon(t, math.Sqrt(3), tight.Radius, 1e-5,
		"the radius should be the distance to the furthest point, bar rounding slack")

	// A single point gives a degenerate sphere at that point, not an empty one.
	single, ok := NewSphereFromPoints([]Vector3[float64]{{X: 5}})
	require.True(t, ok)
	assert.Equal(t, Vector3[float64]{X: 5}, single.Center)
	assert.Zero(t, single.Radius)
}

func TestSphereUnionContainsBoth(t *testing.T) {
	rng := rand.New(rand.NewSource(315))

	grewBoth := 0

	for i := 0; i < 2000; i++ {
		a := NewSphere(Vector3[float64]{
			X: (rng.Float64()*2 - 1) * 5,
			Y: (rng.Float64()*2 - 1) * 5,
			Z: (rng.Float64()*2 - 1) * 5,
		}, rng.Float64()*3+0.1)
		b := NewSphere(Vector3[float64]{
			X: (rng.Float64()*2 - 1) * 5,
			Y: (rng.Float64()*2 - 1) * 5,
			Z: (rng.Float64()*2 - 1) * 5,
		}, rng.Float64()*3+0.1)

		union := a.Union(b)

		// Expanded by a relative epsilon, because Union's centre is displaced by a
		// division and containment therefore holds to rounding rather than exactly.
		// See the note in Union.
		grown := union.Expanded(union.Radius * 1e-12)
		require.True(t, grown.ContainsSphere(a),
			"union %v does not contain %v", union, a)
		require.True(t, grown.ContainsSphere(b),
			"union %v does not contain %v", union, b)

		if union.Radius > a.Radius && union.Radius > b.Radius {
			grewBoth++
		}

		// The union must not be wastefully large: it can never exceed the span
		// from one far side to the other.
		span := (b.Center.Sub(a.Center).Length() + a.Radius + b.Radius) / 2
		require.LessOrEqual(t, union.Radius,
			math.Max(span, math.Max(a.Radius, b.Radius))*(1+1e-12))
	}

	assert.Greater(t, grewBoth, 100, "the sample should include genuinely disjoint pairs")
}

func TestSphereUnionOfConcentricKeepsTheCentre(t *testing.T) {
	// Concentric spheres are the case with no direction to move the centre along,
	// which is a division by zero if handled carelessly.
	small := NewSphere(Vector3[float64]{X: 1, Y: 1}, 1)
	large := NewSphere(Vector3[float64]{X: 1, Y: 1}, 4)

	union := small.Union(large)
	assert.Equal(t, large, union)
	assert.Equal(t, large, large.Union(small))
}

func TestSphereVolumeAndSurfaceArea(t *testing.T) {
	sphere := NewSphere(Vector3[float64]{}, 2)

	assert.InDelta(t, 4.0/3.0*math.Pi*8, sphere.Volume(), 1e-12)
	assert.InDelta(t, 4*math.Pi*4, sphere.SurfaceArea(), 1e-12)

	// Cross-check the volume by Monte-Carlo integration over the bounding box,
	// which never touches the closed form.
	rng := rand.New(rand.NewSource(316))
	box := sphere.BoundingBox()
	hits := 0
	const samples = 200000
	for i := 0; i < samples; i++ {
		point := Vector3[float64]{
			X: box.Center.X + (rng.Float64()*2-1)*box.HalfSize.X,
			Y: box.Center.Y + (rng.Float64()*2-1)*box.HalfSize.Y,
			Z: box.Center.Z + (rng.Float64()*2-1)*box.HalfSize.Z,
		}
		if sphere.ContainsPoint(point) {
			hits++
		}
	}

	boxVolume := 8 * box.HalfSize.X * box.HalfSize.Y * box.HalfSize.Z
	estimate := boxVolume * float64(hits) / samples
	assert.InDelta(t, sphere.Volume(), estimate, sphere.Volume()*0.02,
		"sampled volume %v should match the closed form %v", estimate, sphere.Volume())
}

func TestSphereIsEmptyAndTransforms(t *testing.T) {
	assert.True(t, NewSphere(Vector3[float64]{}, 0).IsEmpty(), "zero radius encloses nothing")
	assert.True(t, NewSphere(Vector3[float64]{}, -1).IsEmpty())
	assert.False(t, NewSphere(Vector3[float64]{}, 0.001).IsEmpty())

	sphere := NewSphere(Vector3[float64]{X: 1}, 2)

	moved := sphere.Translate(Vector3[float64]{Y: 3})
	assert.Equal(t, Vector3[float64]{X: 1, Y: 3}, moved.Center)
	assert.Equal(t, sphere.Radius, moved.Radius, "translation must not change the radius")

	// The receivers are values, so the original must be untouched by both.
	assert.Equal(t, Vector3[float64]{X: 1}, sphere.Center)
	assert.Equal(t, 2.0, sphere.Radius)

	assert.Equal(t, 2.5, sphere.Expanded(0.5).Radius)
}

func TestSphereEquality(t *testing.T) {
	a := NewSphere(Vector3[float64]{X: 1}, 2)

	assert.True(t, a.Equals(NewSphere(Vector3[float64]{X: 1}, 2)))
	assert.False(t, a.Equals(NewSphere(Vector3[float64]{X: 1}, 2.0000001)))
	assert.True(t, a.ApproxEquals(NewSphere(Vector3[float64]{X: 1}, 2.0000001), 1e-6))
	assert.False(t, a.ApproxEquals(NewSphere(Vector3[float64]{X: 1.1}, 2), 1e-6))
}

func TestSphereWorksWithIntegerScalars(t *testing.T) {
	// Sphere is constrained to Number, not Float, so it must be usable with an
	// integer scalar. The measures truncate, which is the expected consequence of
	// asking for an integer answer.
	// Radius 4 rather than 3, because at radius 3 the volume and the surface area
	// are both 36*pi and a swapped formula would pass either assertion.
	sphere := NewSphere(Vector3[int]{X: 2}, 4)

	assert.True(t, sphere.ContainsPoint(Vector3[int]{X: 5}))
	assert.False(t, sphere.ContainsPoint(Vector3[int]{X: 7}))
	assert.Equal(t, 268, sphere.Volume(), "4/3*pi*64 truncated")
	assert.Equal(t, 201, sphere.SurfaceArea(), "4*pi*16 truncated")
}

// randomPointInSphere returns a point uniformly inside the sphere, by rejection
// against the bounding box — simple, and unbiased in a way that scaling a random
// direction by a random radius is not.
func randomPointInSphere(rng *rand.Rand, sphere Sphere[float64]) Vector3[float64] {
	for {
		candidate := Vector3[float64]{
			X: (rng.Float64()*2 - 1) * sphere.Radius,
			Y: (rng.Float64()*2 - 1) * sphere.Radius,
			Z: (rng.Float64()*2 - 1) * sphere.Radius,
		}
		if candidate.LengthSq() <= sphere.Radius*sphere.Radius {
			return sphere.Center.Add(candidate)
		}
	}
}
