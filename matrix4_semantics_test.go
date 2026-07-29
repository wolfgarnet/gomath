package gomath

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file holds the property tests that pin down Matrix4's semantics, as
// opposed to matrix4_test.go's example-by-example checks: inverse really
// inverts, the projections really map the view volume onto the clip cube, and
// the rotation matrices really agree with the rotation code gomath already had.

// randomInvertibleMatrix4 builds a transform out of a translation, three axis
// rotations and a non-zero scale. Composed that way it is always invertible, so
// the inverse property test never has to skip a degenerate sample.
func randomInvertibleMatrix4(rng *rand.Rand) Matrix4[float64] {
	angle := func() float64 { return (rng.Float64()*2 - 1) * math.Pi }
	offset := func() float64 { return (rng.Float64()*2 - 1) * 100 }
	// Keep scales away from zero so the matrix stays well conditioned.
	scale := func() float64 {
		s := rng.Float64()*4 + 0.25
		if rng.Intn(2) == 0 {
			return -s
		}
		return s
	}

	return TranslationMatrix4(offset(), offset(), offset()).
		Mul(RotationMatrix4X(angle())).
		Mul(RotationMatrix4Y(angle())).
		Mul(RotationMatrix4Z(angle())).
		Mul(ScaleMatrix4(scale(), scale(), scale()))
}

func TestMatrix4Semantics_InverseUndoesTransform(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	identity := Identity4[float64]()

	for i := 0; i < 200; i++ {
		m := randomInvertibleMatrix4(rng)

		inverse, ok := m.Inverse()
		require.True(t, ok, "composed transform must be invertible")

		// Both orders must return the identity.
		assert.True(t, m.Mul(inverse).ApproxEquals(identity, 1e-9),
			"m*inv should be identity, got %v", m.Mul(inverse))
		assert.True(t, inverse.Mul(m).ApproxEquals(identity, 1e-9),
			"inv*m should be identity, got %v", inverse.Mul(m))
	}
}

func TestMatrix4Semantics_InverseRoundTripsPoints(t *testing.T) {
	rng := rand.New(rand.NewSource(2))

	for i := 0; i < 200; i++ {
		m := randomInvertibleMatrix4(rng)
		inverse, ok := m.Inverse()
		require.True(t, ok)

		point := Vector3[float64]{rng.Float64() * 50, rng.Float64() * 50, rng.Float64() * 50}
		got := inverse.TransformPoint(m.TransformPoint(point))

		assert.InDelta(t, point.X, got.X, 1e-6)
		assert.InDelta(t, point.Y, got.Y, 1e-6)
		assert.InDelta(t, point.Z, got.Z, 1e-6)
	}
}

func TestMatrix4Semantics_MulIsAssociative(t *testing.T) {
	rng := rand.New(rand.NewSource(3))

	for i := 0; i < 100; i++ {
		a := randomInvertibleMatrix4(rng)
		b := randomInvertibleMatrix4(rng)
		c := randomInvertibleMatrix4(rng)

		assert.True(t, a.Mul(b).Mul(c).ApproxEquals(a.Mul(b.Mul(c)), 1e-6))
	}
}

func TestMatrix4Semantics_DeterminantOfProductIsProductOfDeterminants(t *testing.T) {
	rng := rand.New(rand.NewSource(4))

	for i := 0; i < 100; i++ {
		a := randomInvertibleMatrix4(rng)
		b := randomInvertibleMatrix4(rng)

		want := a.Determinant() * b.Determinant()
		got := a.Mul(b).Determinant()

		assert.InDelta(t, want, got, math.Abs(want)*1e-9+1e-9)
	}
}

func TestMatrix4Semantics_RotationAxisAgreesWithVector3RotateAxis(t *testing.T) {
	// RotationMatrix4Axis and Vector3.RotateAxis are independent implementations
	// of the same rotation, so they are a genuine cross-check on the handedness
	// and on the column-major layout.
	rng := rand.New(rand.NewSource(5))

	for i := 0; i < 200; i++ {
		axis := Vector3[float64]{
			X: rng.Float64()*2 - 1,
			Y: rng.Float64()*2 - 1,
			Z: rng.Float64()*2 - 1,
		}.Normalize()
		if axis.IsZero() {
			continue
		}

		phi := (rng.Float64()*2 - 1) * math.Pi
		v := Vector3[float64]{rng.Float64() * 10, rng.Float64() * 10, rng.Float64() * 10}

		want := v.RotateAxis(axis, phi)
		rotation := RotationMatrix4Axis(axis, phi)
		got := rotation.TransformDirection(v)

		assert.InDelta(t, want.X, got.X, 1e-9)
		assert.InDelta(t, want.Y, got.Y, 1e-9)
		assert.InDelta(t, want.Z, got.Z, 1e-9)
	}
}

func TestMatrix4Semantics_RotationZAgreesWith2DRotationMatrix(t *testing.T) {
	// The XY plane is shared with the 2D Matrix, so the two must agree there.
	for _, theta := range []float64{0, 0.3, 1.2, math.Pi / 2, -0.75} {
		m2 := RotationMatrix(theta)
		m4 := RotationMatrix4Z(theta)

		v := Vector2[float64]{3, -4}
		want := m2.TransformPoint(v)
		got := m4.TransformDirection(Vector3[float64]{v.X, v.Y, 0})

		assert.InDelta(t, want.X, got.X, 1e-12)
		assert.InDelta(t, want.Y, got.Y, 1e-12)
		assert.InDelta(t, 0, got.Z, 1e-12)
	}
}

func TestMatrix4Semantics_RotationsPreserveLengthAndOrthonormality(t *testing.T) {
	rng := rand.New(rand.NewSource(6))

	for i := 0; i < 100; i++ {
		axis := Vector3[float64]{
			X: rng.Float64()*2 - 1,
			Y: rng.Float64()*2 - 1,
			Z: rng.Float64()*2 - 1,
		}.Normalize()
		if axis.IsZero() {
			continue
		}

		m := RotationMatrix4Axis(axis, (rng.Float64()*2-1)*math.Pi)

		// A rotation is volume preserving.
		assert.InDelta(t, 1, m.Determinant(), 1e-9)

		// Its inverse is its transpose.
		inverse, ok := m.Inverse()
		require.True(t, ok)
		assert.True(t, inverse.ApproxEquals(m.Transpose(), 1e-9))

		// And it preserves the length of what it rotates.
		v := Vector3[float64]{1, 2, 3}
		assert.InDelta(t, v.Length(), m.TransformDirection(v).Length(), 1e-9)
	}
}

func TestMatrix4Semantics_PerspectiveMapsFrustumCornersToClipCube(t *testing.T) {
	// The defining property of the projection: the eight corners of the view
	// frustum land on the eight corners of the NDC cube.
	const (
		fovY   = math.Pi / 3
		aspect = 16.0 / 9.0
		near   = 0.1
		far    = 500.0
	)

	projection := PerspectiveMatrix4[float64](fovY, aspect, near, far)
	tangent := math.Tan(fovY / 2)

	for _, plane := range []struct {
		distance float64
		wantZ    float64
	}{
		{near, -1}, // the near plane maps to z = -1
		{far, +1},  // the far plane maps to z = +1
	} {
		halfHeight := plane.distance * tangent
		halfWidth := halfHeight * aspect

		for _, sx := range []float64{-1, 1} {
			for _, sy := range []float64{-1, 1} {
				// Eye space looks down -Z, so the corner is at negative z.
				corner := Vector3[float64]{
					X: sx * halfWidth,
					Y: sy * halfHeight,
					Z: -plane.distance,
				}

				ndc := projection.TransformVector4(corner.ToVector4(1)).PerspectiveDivide()

				assert.InDelta(t, sx, ndc.X, 1e-9)
				assert.InDelta(t, sy, ndc.Y, 1e-9)
				assert.InDelta(t, plane.wantZ, ndc.Z, 1e-9)
			}
		}
	}
}

func TestMatrix4Semantics_PerspectiveMatchesSymmetricFrustum(t *testing.T) {
	// PerspectiveMatrix4 is the symmetric special case of FrustumMatrix4, so the
	// two must produce the same matrix when given equivalent bounds.
	const (
		fovY   = math.Pi / 4
		aspect = 4.0 / 3.0
		near   = 0.5
		far    = 100.0
	)

	halfHeight := near * math.Tan(fovY/2)
	halfWidth := halfHeight * aspect

	perspective := PerspectiveMatrix4[float64](fovY, aspect, near, far)
	frustum := FrustumMatrix4[float64](-halfWidth, halfWidth, -halfHeight, halfHeight, near, far)

	assert.True(t, perspective.ApproxEquals(frustum, 1e-12),
		"perspective %v should equal symmetric frustum %v", perspective, frustum)
}

func TestMatrix4Semantics_OrthographicMapsBoxCornersToClipCube(t *testing.T) {
	const (
		left   = -8.0
		right  = 12.0
		bottom = -5.0
		top    = 15.0
		near   = 1.0
		far    = 250.0
	)

	projection := OrthographicMatrix4[float64](left, right, bottom, top, near, far)

	for _, corner := range []struct {
		point Vector3[float64]
		want  Vector3[float64]
	}{
		{Vector3[float64]{left, bottom, -near}, Vector3[float64]{-1, -1, -1}},
		{Vector3[float64]{right, top, -near}, Vector3[float64]{1, 1, -1}},
		{Vector3[float64]{left, bottom, -far}, Vector3[float64]{-1, -1, 1}},
		{Vector3[float64]{right, top, -far}, Vector3[float64]{1, 1, 1}},
	} {
		ndc := projection.TransformVector4(corner.point.ToVector4(1)).PerspectiveDivide()

		assert.InDelta(t, corner.want.X, ndc.X, 1e-12)
		assert.InDelta(t, corner.want.Y, ndc.Y, 1e-12)
		assert.InDelta(t, corner.want.Z, ndc.Z, 1e-12)
	}
}

func TestMatrix4Semantics_OrthographicKeepsParallelLinesParallel(t *testing.T) {
	// Unlike a perspective projection, an orthographic one has no vanishing
	// point: the same offset maps to the same screen offset at any depth.
	projection := OrthographicMatrix4[float64](-10, 10, -10, 10, 1, 100)

	nearPair := projection.TransformPoint(Vector3[float64]{5, 0, -2}).
		Sub(projection.TransformPoint(Vector3[float64]{0, 0, -2}))
	farPair := projection.TransformPoint(Vector3[float64]{5, 0, -90}).
		Sub(projection.TransformPoint(Vector3[float64]{0, 0, -90}))

	assert.InDelta(t, nearPair.X, farPair.X, 1e-12)
}

func TestMatrix4Semantics_LookAtPlacesCameraAtOriginLookingDownNegativeZ(t *testing.T) {
	eye := Vector3[float64]{10, 5, 20}
	target := Vector3[float64]{1, 2, 3}
	up := Vector3[float64]{0, 1, 0}

	view := LookAtMatrix4(eye, target, up)

	// The eye maps to the origin of eye space.
	atEye := view.TransformPoint(eye)
	assert.InDelta(t, 0, atEye.X, 1e-9)
	assert.InDelta(t, 0, atEye.Y, 1e-9)
	assert.InDelta(t, 0, atEye.Z, 1e-9)

	// The target maps onto the -Z axis, at the distance between the two.
	atTarget := view.TransformPoint(target)
	assert.InDelta(t, 0, atTarget.X, 1e-9)
	assert.InDelta(t, 0, atTarget.Y, 1e-9)
	assert.InDelta(t, -eye.Sub(target).Length(), atTarget.Z, 1e-9)
}

func TestMatrix4Semantics_LookAtBasisIsOrthonormal(t *testing.T) {
	rng := rand.New(rand.NewSource(7))

	for i := 0; i < 100; i++ {
		eye := Vector3[float64]{rng.Float64() * 50, rng.Float64() * 50, rng.Float64() * 50}
		target := Vector3[float64]{rng.Float64() * 50, rng.Float64() * 50, rng.Float64() * 50}
		if eye.Sub(target).IsZero() {
			continue
		}

		view := LookAtMatrix4(eye, target, Vector3[float64]{0, 1, 0})

		// A view matrix is a rigid transform: no scale, no shear.
		scale := view.GetScale()
		assert.InDelta(t, 1, scale.X, 1e-9)
		assert.InDelta(t, 1, scale.Y, 1e-9)
		assert.InDelta(t, 1, scale.Z, 1e-9)
		assert.InDelta(t, 1, view.Determinant(), 1e-9)

		// Distances are preserved.
		a := Vector3[float64]{1, 2, 3}
		b := Vector3[float64]{-4, 0, 6}
		assert.InDelta(t,
			a.Sub(b).Length(),
			view.TransformPoint(a).Sub(view.TransformPoint(b)).Length(),
			1e-9)
	}
}

func TestMatrix4Semantics_LookAtDegenerateUpReturnsIdentity(t *testing.T) {
	// Looking straight up with an up vector along the view direction leaves the
	// basis undefined; the identity is returned so the failure is visible rather
	// than a matrix full of NaNs silently poisoning the frame.
	view := LookAtMatrix4(
		Vector3[float64]{0, 10, 0},
		Vector3[float64]{0, 0, 0},
		Vector3[float64]{0, 1, 0},
	)

	assert.Equal(t, Identity4[float64](), view)

	// An eye sitting on the target is equally degenerate.
	assert.Equal(t, Identity4[float64](), LookAtMatrix4(
		Vector3[float64]{1, 1, 1},
		Vector3[float64]{1, 1, 1},
		Vector3[float64]{0, 1, 0},
	))
}

func TestMatrix4Semantics_ViewProjectionKeepsVisiblePointsInsideClipCube(t *testing.T) {
	// The end-to-end path a renderer actually uses: world -> eye -> clip.
	view := LookAtMatrix4(
		Vector3[float64]{0, 0, 10},
		Vector3[float64]{0, 0, 0},
		Vector3[float64]{0, 1, 0},
	)
	projection := PerspectiveMatrix4[float64](math.Pi/3, 1, 0.1, 100)
	viewProjection := projection.Mul(view)

	// A point at the world origin is dead ahead, so it lands at the centre of
	// the screen and strictly inside the depth range.
	centre := viewProjection.TransformVector4(Vector3[float64]{}.ToVector4(1)).PerspectiveDivide()
	assert.InDelta(t, 0, centre.X, 1e-9)
	assert.InDelta(t, 0, centre.Y, 1e-9)
	assert.Greater(t, centre.Z, -1.0)
	assert.Less(t, centre.Z, 1.0)

	// A point behind the camera projects outside the near plane.
	behind := viewProjection.TransformVector4(Vector3[float64]{0, 0, 20}.ToVector4(1))
	assert.LessOrEqual(t, behind.W, 0.0, "a point behind the eye must have non-positive W")
}
