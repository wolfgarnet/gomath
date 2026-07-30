package gomath

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The plane extraction is cross-checked against clip space itself.
//
// A point is inside the view volume exactly when its clip coordinates satisfy
// -w <= x,y,z <= w. That is the definition the extraction is derived FROM, but
// evaluating it is a completely different computation: push the point through
// the matrix and compare four numbers, versus six precomputed plane equations.
// Agreement therefore says the algebra was rearranged correctly, which is the
// only thing that can really go wrong here.

// insideClipVolume is the oracle: transform the point and test the clip cube
// directly, with no planes involved.
func insideClipVolume(viewProjection Matrix4[float64], point Vector3[float64]) bool {
	clip := viewProjection.TransformVector4(point.ToVector4(1))

	// Behind the eye: w <= 0 means the point is not in the visible volume at all.
	if clip.W <= 0 {
		return false
	}

	return clip.X >= -clip.W && clip.X <= clip.W &&
		clip.Y >= -clip.W && clip.Y <= clip.W &&
		clip.Z >= -clip.W && clip.Z <= clip.W
}

func testViewProjection() Matrix4[float64] {
	view := LookAtMatrix4(
		Vector3[float64]{X: 0, Y: 0, Z: 10},
		Vector3[float64]{},
		Vector3[float64]{Y: 1})
	projection := PerspectiveMatrix4[float64](DegToRad[float64](60), 16.0/9.0, 0.5, 100)

	return projection.Mul(view)
}

func TestFrustumAgreesWithClipSpace(t *testing.T) {
	viewProjection := testViewProjection()
	frustum := NewFrustumFromMatrix(viewProjection)

	rng := rand.New(rand.NewSource(81))

	inside, outside := 0, 0

	for i := 0; i < 20000; i++ {
		point := Vector3[float64]{
			X: (rng.Float64()*2 - 1) * 40,
			Y: (rng.Float64()*2 - 1) * 40,
			Z: (rng.Float64()*2 - 1) * 60,
		}

		oracle := insideClipVolume(viewProjection, point)
		planes := frustum.ContainsPoint(point)

		if oracle {
			inside++
		} else {
			outside++
		}

		require.Equal(t, oracle, planes,
			"disagreement at %v: clip space says %v, planes say %v", point, oracle, planes)
	}

	// A test where everything lands on one side of the answer proves nothing.
	assert.Greater(t, inside, 200, "the sample should contain visible points")
	assert.Greater(t, outside, 200, "the sample should contain culled points")
}

func TestFrustumAgreesWithClipSpaceForOrthographic(t *testing.T) {
	// The extraction is derived from clip space, not from camera parameters, so
	// it must work unchanged for a projection with no vanishing point.
	view := LookAtMatrix4(
		Vector3[float64]{X: 2, Y: 3, Z: 12},
		Vector3[float64]{},
		Vector3[float64]{Y: 1})
	projection := OrthographicMatrix4[float64](-8, 8, -5, 5, 0.5, 60)
	viewProjection := projection.Mul(view)

	frustum := NewFrustumFromMatrix(viewProjection)

	rng := rand.New(rand.NewSource(82))
	inside := 0

	for i := 0; i < 20000; i++ {
		point := Vector3[float64]{
			X: (rng.Float64()*2 - 1) * 20,
			Y: (rng.Float64()*2 - 1) * 20,
			Z: (rng.Float64()*2 - 1) * 40,
		}

		// An orthographic w is always 1, so the clip test needs no w guard, but
		// the oracle handles it uniformly anyway.
		oracle := insideClipVolume(viewProjection, point)
		if oracle {
			inside++
		}

		require.Equal(t, oracle, frustum.ContainsPoint(point), "at %v", point)
	}

	assert.Greater(t, inside, 200)
}

func TestFrustumPlaneNormalsPointInward(t *testing.T) {
	// The inward convention is what makes every test "negative on some plane
	// means outside". If a normal were flipped, that plane would cull everything
	// visible and pass everything hidden.
	frustum := NewFrustumFromMatrix(testViewProjection())

	// A point clearly inside the volume: on the view axis, between the planes.
	centre := Vector3[float64]{X: 0, Y: 0, Z: 0}
	require.True(t, frustum.ContainsPoint(centre), "setup: the origin should be visible")

	for index, plane := range frustum.Planes {
		assert.Greater(t, plane.SignedDistance(centre), 0.0,
			"plane %d should report a visible point as in front, so its normal points inward", index)

		// And the normals must be unit length, or distances are not metric.
		assert.InDelta(t, 1, plane.Normal.Length(), 1e-9, "plane %d normal", index)
	}
}

func TestFrustumNearAndFarPlanes(t *testing.T) {
	// A camera at +Z looking at the origin with near 0.5 and far 100: the near
	// plane sits at z = 9.5 and the far at z = -90.
	frustum := NewFrustumFromMatrix(testViewProjection())

	assert.True(t, frustum.ContainsPoint(Vector3[float64]{Z: 9.4}), "just inside near")
	assert.False(t, frustum.ContainsPoint(Vector3[float64]{Z: 9.6}), "just outside near")

	assert.True(t, frustum.ContainsPoint(Vector3[float64]{Z: -89}), "just inside far")
	assert.False(t, frustum.ContainsPoint(Vector3[float64]{Z: -91}), "beyond far")

	// Behind the camera entirely.
	assert.False(t, frustum.ContainsPoint(Vector3[float64]{Z: 20}))
}

func TestFrustumBoxTestIsConsistentWithPointTest(t *testing.T) {
	// A box containing a visible point must be reported visible. The converse
	// does not hold, since the box test is deliberately conservative.
	frustum := NewFrustumFromMatrix(testViewProjection())
	rng := rand.New(rand.NewSource(83))

	visibleBoxes := 0

	for i := 0; i < 5000; i++ {
		centre := Vector3[float64]{
			X: (rng.Float64()*2 - 1) * 20,
			Y: (rng.Float64()*2 - 1) * 20,
			Z: (rng.Float64()*2 - 1) * 40,
		}
		box := NewBox3(centre, Vector3[float64]{
			X: rng.Float64() + 0.1,
			Y: rng.Float64() + 0.1,
			Z: rng.Float64() + 0.1,
		})

		if frustum.IntersectsBox3(box) {
			visibleBoxes++

			continue
		}

		// Reported invisible, so NO corner and not the centre may be visible.
		assert.False(t, frustum.ContainsPoint(centre),
			"a culled box contained a visible centre: %v", box)
		for _, corner := range box.Corners() {
			assert.False(t, frustum.ContainsPoint(corner),
				"a culled box contained a visible corner %v: %v", corner, box)
		}
	}

	assert.Greater(t, visibleBoxes, 100)
}

func TestFrustumContainsBoxImpliesIntersects(t *testing.T) {
	frustum := NewFrustumFromMatrix(testViewProjection())
	rng := rand.New(rand.NewSource(84))

	fullyInside := 0

	for i := 0; i < 5000; i++ {
		box := NewBox3(
			Vector3[float64]{
				X: (rng.Float64()*2 - 1) * 10,
				Y: (rng.Float64()*2 - 1) * 10,
				Z: (rng.Float64()*2 - 1) * 30,
			},
			Vector3[float64]{X: rng.Float64(), Y: rng.Float64(), Z: rng.Float64()})

		if !frustum.ContainsBox3(box) {
			continue
		}
		fullyInside++

		assert.True(t, frustum.IntersectsBox3(box),
			"a box fully inside must also intersect")

		// Fully inside means every corner is visible.
		for _, corner := range box.Corners() {
			assert.True(t, frustum.ContainsPoint(corner),
				"a fully-contained box had a corner outside: %v", corner)
		}
	}

	assert.Greater(t, fullyInside, 50, "the sample should contain fully-inside boxes")
}

func TestFrustumOBBIsTighterThanItsBounds(t *testing.T) {
	// An oriented box's axis-aligned bounds are larger than the box, so anything
	// the OBB test culls the bounds test must also cull — never the reverse.
	frustum := NewFrustumFromMatrix(testViewProjection())
	rng := rand.New(rand.NewSource(85))

	tighter := 0

	for i := 0; i < 5000; i++ {
		axis := Vector3[float64]{
			X: rng.Float64()*2 - 1,
			Y: rng.Float64()*2 - 1,
			Z: rng.Float64()*2 - 1,
		}.Normalize()
		if axis.IsZero() {
			continue
		}

		oriented := NewOBB3FromQuaternion(
			Vector3[float64]{
				X: (rng.Float64()*2 - 1) * 18,
				Y: (rng.Float64()*2 - 1) * 18,
				Z: (rng.Float64()*2 - 1) * 40,
			},
			Vector3[float64]{X: rng.Float64()*3 + 0.2, Y: rng.Float64()*0.3 + 0.1, Z: rng.Float64()*0.3 + 0.1},
			QuaternionFromAxisAngle(axis, (rng.Float64()*2-1)*math.Pi))

		orientedVisible := frustum.IntersectsOBB3(oriented)
		boundsVisible := frustum.IntersectsBox3(oriented.BoundingBox())

		if orientedVisible {
			require.True(t, boundsVisible,
				"the loose bounds culled something the tight test kept, which is backwards")
		}
		if boundsVisible && !orientedVisible {
			tighter++
		}
	}

	// If the OBB test never wins, it is not actually tighter and there is no
	// reason for it to exist.
	assert.Greater(t, tighter, 10,
		"the oriented test should cull strictly more than the axis-aligned bounds")
}

func TestFrustumSphere(t *testing.T) {
	frustum := NewFrustumFromMatrix(testViewProjection())

	// A tiny sphere at the origin is visible; the same sphere far behind is not.
	assert.True(t, frustum.IntersectsSphere(NewSphere(Vector3[float64]{}, 0.1)))
	assert.False(t, frustum.IntersectsSphere(NewSphere(Vector3[float64]{Z: 200}, 0.1)))

	// A sphere huge enough to reach into the volume is visible even with its
	// centre far outside.
	assert.True(t, frustum.IntersectsSphere(NewSphere(Vector3[float64]{X: 100}, 200)))

	// A sphere fully inside must also intersect, and its own bounding box must
	// be reported visible too — the box is larger, so anything else would be
	// inconsistent.
	inside := NewSphere(Vector3[float64]{}, 0.5)
	assert.True(t, frustum.ContainsSphere(inside))
	assert.True(t, frustum.IntersectsSphere(inside))
	assert.True(t, frustum.IntersectsBox3(inside.BoundingBox()))
}

func TestFrustumIndicesMatchThePlanesTheyName(t *testing.T) {
	frustum := NewFrustumFromMatrix(testViewProjection())

	// The camera looks down -Z from +Z, so the near plane's inward normal points
	// away from the camera and the far plane's points back towards it.
	assert.InDelta(t, -1, frustum.Planes[FrustumNear].Normal.Z, 1e-9)
	assert.InDelta(t, 1, frustum.Planes[FrustumFar].Normal.Z, 1e-9)

	// Left and right planes lean inward from either side.
	assert.Greater(t, frustum.Planes[FrustumLeft].Normal.X, 0.0)
	assert.Less(t, frustum.Planes[FrustumRight].Normal.X, 0.0)
	assert.Greater(t, frustum.Planes[FrustumBottom].Normal.Y, 0.0)
	assert.Less(t, frustum.Planes[FrustumTop].Normal.Y, 0.0)
}
