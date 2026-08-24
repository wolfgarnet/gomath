package gomath

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewCapsule3Upright's height is the whole character, caps included, so the
// bounding box it produces is exactly that tall. Getting this wrong by two radii
// is the kind of mistake that only shows up as a character standing slightly
// inside the floor.
func TestNewCapsule3UprightTotalHeight(t *testing.T) {
	feet := Vector3[float64]{X: 3, Y: 10, Z: -4}
	capsule := NewCapsule3Upright(feet, 1.8, 0.3)

	assert.Equal(t, Vector3[float64]{X: 3, Y: 10.3, Z: -4}, capsule.Segment.P1)
	assert.Equal(t, Vector3[float64]{X: 3, Y: 11.5, Z: -4}, capsule.Segment.P2)

	assert.InDelta(t, 1.8, capsule.Height(), 1e-12)

	bounds := capsule.BoundingBox()
	assert.InDelta(t, 10, bounds.Min().Y, 1e-12, "the soles are at the base")
	assert.InDelta(t, 11.8, bounds.Max().Y, 1e-12, "the crown is a full height above it")
	assert.InDelta(t, 1.8, bounds.Size().Y, 1e-12)
}

// Asking for a capsule shorter than its own caps has no answer, and the
// documented behaviour is to collapse to the sphere of that radius rather than to
// fail mid-frame.
func TestNewCapsule3UprightBelowTwiceTheRadiusIsASphere(t *testing.T) {
	capsule := NewCapsule3Upright(Vector3[float64]{}, 0.4, 0.5)

	require.True(t, capsule.IsSphere())

	sphere, ok := capsule.Sphere()
	require.True(t, ok)
	assert.Equal(t, Vector3[float64]{Y: 0.2}, sphere.Center, "centred at mid-height")
	assert.Equal(t, 0.5, sphere.Radius, "never smaller than asked for")
}

// The reduction the whole feature leans on: a degenerate capsule is a sphere, and
// every self-contained helper must agree with Sphere's.
func TestCapsule3DegenerateAgreesWithSphere(t *testing.T) {
	sphere := Sphere[float64]{Center: Vector3[float64]{X: 1, Y: -2, Z: 3}, Radius: 0.75}
	capsule := NewCapsule3FromSphere(sphere)

	require.True(t, capsule.IsSphere())

	assert.Equal(t, sphere.Center, capsule.Center())
	assert.InDelta(t, 2*sphere.Radius, capsule.Height(), 1e-12)
	assert.InDelta(t, sphere.Volume(), capsule.Volume(), 1e-12)
	assert.InDelta(t, sphere.SurfaceArea(), capsule.SurfaceArea(), 1e-12)
	assert.Equal(t, sphere.BoundingBox(), capsule.BoundingBox())

	random := rand.New(rand.NewSource(20260824))
	for range 200 {
		point := Vector3[float64]{
			X: random.Float64()*6 - 2,
			Y: random.Float64()*6 - 5,
			Z: random.Float64() * 6,
		}

		assert.Equal(t, sphere.ContainsPoint(point), capsule.ContainsPoint(point), "at %v", point)

		wanted := sphere.ClosestPoint(point)
		assert.InDelta(t, 0, wanted.DistanceTo(capsule.ClosestPoint(point)), 1e-12, "at %v", point)
	}
}

func TestCapsule3ContainsPoint(t *testing.T) {
	// A metre of segment along X, half a metre of radius.
	capsule := NewCapsule3(Vector3[float64]{}, Vector3[float64]{X: 1}, 0.5)

	tests := []struct {
		name  string
		point Vector3[float64]
		want  bool
	}{
		{"on the segment", Vector3[float64]{X: 0.5}, true},
		{"beside the body, inside", Vector3[float64]{X: 0.5, Y: 0.49}, true},
		{"beside the body, outside", Vector3[float64]{X: 0.5, Y: 0.51}, false},
		{"exactly on the surface counts", Vector3[float64]{X: 0.5, Y: 0.5}, true},
		{"inside the far cap", Vector3[float64]{X: 1.4}, true},
		{"past the far cap", Vector3[float64]{X: 1.6}, false},
		{"diagonally past the near cap", Vector3[float64]{X: -0.4, Y: 0.4}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, capsule.ContainsPoint(test.point))
		})
	}
}

// ClosestPoint must land on the surface for anything outside, and the surface is
// the set of points exactly one radius from the segment.
func TestCapsule3ClosestPointLandsOnTheSurface(t *testing.T) {
	capsule := NewCapsule3(Vector3[float64]{Y: -1}, Vector3[float64]{Y: 1}, 0.6)

	random := rand.New(rand.NewSource(20260824))

	for range 500 {
		point := Vector3[float64]{
			X: random.Float64()*8 - 4,
			Y: random.Float64()*8 - 4,
			Z: random.Float64()*8 - 4,
		}

		closest := capsule.ClosestPoint(point)

		if capsule.ContainsPoint(point) {
			assert.Equal(t, point, closest, "an interior point is its own closest")

			continue
		}

		assert.InDelta(t, capsule.Radius, capsule.Segment.DistanceToPoint(closest), 1e-12,
			"the returned point must be on the surface, for %v", point)

		// And it must be the NEAREST surface point, which a sampled sweep of the
		// surface can only ever fail to beat.
		best := math.Inf(1)
		for i := range 200 {
			onAxis := capsule.Segment.PointAt(float64(i) / 199)
			best = math.Min(best, math.Abs(onAxis.DistanceTo(point)-capsule.Radius))
		}

		assert.LessOrEqual(t, closest.DistanceTo(point), best+1e-9, "for %v", point)
	}
}

func TestCapsule3BoundingBoxIsTight(t *testing.T) {
	capsule := NewCapsule3(Vector3[float64]{X: -1, Y: 2, Z: 0}, Vector3[float64]{X: 3, Y: 2, Z: 5}, 0.5)
	bounds := capsule.BoundingBox()

	assert.Equal(t, Vector3[float64]{X: -1.5, Y: 1.5, Z: -0.5}, bounds.Min())
	assert.Equal(t, Vector3[float64]{X: 3.5, Y: 2.5, Z: 5.5}, bounds.Max())

	// Tight means every face is touched: a sampled sweep of the surface must reach
	// each bound and never pass it.
	random := rand.New(rand.NewSource(20260824))
	for range 2000 {
		direction := Vector3[float64]{
			X: random.NormFloat64(),
			Y: random.NormFloat64(),
			Z: random.NormFloat64(),
		}.Normalize()

		for _, end := range []Vector3[float64]{capsule.Segment.P1, capsule.Segment.P2} {
			surface := end.Add(direction.MultiplyScalar(capsule.Radius))

			assert.True(t, bounds.ContainsPoint(surface), "surface point %v escapes the bounds", surface)
		}
	}
}

func TestCapsule3VolumeAndSurfaceArea(t *testing.T) {
	// A unit-radius capsule with a segment of length 2: a cylinder of volume 2*pi
	// plus a whole unit sphere.
	capsule := NewCapsule3(Vector3[float64]{}, Vector3[float64]{Z: 2}, 1)

	assert.InDelta(t, 2*math.Pi+4.0/3.0*math.Pi, capsule.Volume(), 1e-12)
	assert.InDelta(t, 4*math.Pi+4*math.Pi, capsule.SurfaceArea(), 1e-12)
}

func TestCapsule3Transforms(t *testing.T) {
	capsule := NewCapsule3(Vector3[float64]{}, Vector3[float64]{Y: 2}, 0.5)

	moved := capsule.Translate(Vector3[float64]{X: 5})
	assert.Equal(t, Vector3[float64]{X: 5}, moved.Segment.P1)
	assert.Equal(t, Vector3[float64]{}, capsule.Segment.P1, "the original is untouched")

	// Expanding is the Minkowski sum with a ball, so the total height grows by
	// twice the amount, not once.
	fatter := capsule.Expanded(0.25)
	assert.InDelta(t, 0.75, fatter.Radius, 1e-12)
	assert.InDelta(t, capsule.Height()+0.5, fatter.Height(), 1e-12)

	assert.True(t, capsule.Equals(NewCapsule3(Vector3[float64]{}, Vector3[float64]{Y: 2}, 0.5)))
	assert.False(t, capsule.Equals(fatter))
	assert.True(t, capsule.ApproxEquals(NewCapsule3(Vector3[float64]{X: 1e-9}, Vector3[float64]{Y: 2}, 0.5), 1e-6))
	assert.True(t, capsule.IsEmpty() == false)
	assert.True(t, NewCapsule3(Vector3[float64]{}, Vector3[float64]{Y: 2}, 0).IsEmpty())
}
