package gomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatrix4_IdentityAndStorageOrder(t *testing.T) {
	identity := Identity4[float64]()

	assert.Equal(t, identity, NewMatrix4[float64]())
	assert.Equal(t, Vector3[float64]{3, 4, 5}, identity.TransformPoint(Vector3[float64]{3, 4, 5}))

	// NewMatrix4From takes its arguments row by row but stores column-major, so
	// element (row r, col c) must land at index c*4+r.
	m := NewMatrix4From[float64](
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	)

	assert.Equal(t, float64(1), m.At(0, 0))
	assert.Equal(t, float64(2), m.At(0, 1))
	assert.Equal(t, float64(5), m.At(1, 0))
	assert.Equal(t, float64(16), m.At(3, 3))

	// Column 0 is the first four stored elements.
	assert.Equal(t, [4]float64{1, 5, 9, 13}, [4]float64{m[0], m[1], m[2], m[3]})
}

func TestMatrix4_TranslationSitsInColumnThree(t *testing.T) {
	// This is the layout GL depends on: glUniformMatrix4fv with transpose=false
	// reads the translation from indices 12, 13, 14.
	m := TranslationMatrix4[float64](7, 8, 9)

	assert.Equal(t, float64(7), m[12])
	assert.Equal(t, float64(8), m[13])
	assert.Equal(t, float64(9), m[14])
	assert.Equal(t, Vector3[float64]{7, 8, 9}, m.GetPosition())
}

func TestMatrix4_TransformPointVersusDirection(t *testing.T) {
	m := TranslationMatrix4[float64](5, -3, 2)

	// A point picks up the translation...
	assert.Equal(t, Vector3[float64]{6, -1, 5}, m.TransformPoint(Vector3[float64]{1, 2, 3}))
	// ...a direction does not.
	assert.Equal(t, Vector3[float64]{1, 2, 3}, m.TransformDirection(Vector3[float64]{1, 2, 3}))
}

func TestMatrix4_TransformVector4MatchesHomogeneousPoint(t *testing.T) {
	m := TranslationMatrix4[float64](5, -3, 2).Mul(RotationMatrix4Y[float64](0.7)).Scaled(2, 2, 2)
	v := Vector3[float64]{1, 2, 3}

	// Transforming (v, 1) and dividing by W must agree with TransformPoint.
	got := m.TransformVector4(v.ToVector4(1)).PerspectiveDivide()
	want := m.TransformPoint(v)

	assert.InDelta(t, want.X, got.X, 1e-12)
	assert.InDelta(t, want.Y, got.Y, 1e-12)
	assert.InDelta(t, want.Z, got.Z, 1e-12)

	// Transforming (v, 0) must agree with TransformDirection.
	gotDir := m.TransformVector4(v.ToVector4(0)).ToVector3()
	wantDir := m.TransformDirection(v)

	assert.InDelta(t, wantDir.X, gotDir.X, 1e-12)
	assert.InDelta(t, wantDir.Y, gotDir.Y, 1e-12)
	assert.InDelta(t, wantDir.Z, gotDir.Z, 1e-12)
}

func TestMatrix4_MulComposesRightToLeft(t *testing.T) {
	translate := TranslationMatrix4[float64](10, 0, 0)
	scale := ScaleMatrix4[float64](2, 2, 2)

	// translate*scale scales first, then translates: (1,0,0) -> (2,0,0) -> (12,0,0).
	translateScale := translate.Mul(scale)
	assert.Equal(t, Vector3[float64]{12, 0, 0},
		translateScale.TransformPoint(Vector3[float64]{1, 0, 0}))

	// scale*translate translates first, then scales: (1,0,0) -> (11,0,0) -> (22,0,0).
	scaleTranslate := scale.Mul(translate)
	assert.Equal(t, Vector3[float64]{22, 0, 0},
		scaleTranslate.TransformPoint(Vector3[float64]{1, 0, 0}))
}

func TestMatrix4_TransposeRoundTrip(t *testing.T) {
	m := NewMatrix4From[float64](
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	)

	transposed := m.Transpose()

	assert.Equal(t, m, transposed.Transpose())
	assert.Equal(t, m.At(1, 2), transposed.At(2, 1))
}

func TestMatrix4_Determinant(t *testing.T) {
	assert.Equal(t, float64(1), Identity4[float64]().Determinant())

	// A pure scale has determinant sx*sy*sz.
	assert.InDelta(t, 24, ScaleMatrix4[float64](2, 3, 4).Determinant(), 1e-12)

	// Rotation and translation preserve volume, so they leave it unchanged.
	m := TranslationMatrix4[float64](5, 6, 7).
		Mul(RotationMatrix4Y[float64](0.9)).
		Mul(ScaleMatrix4[float64](2, 3, 4))
	assert.InDelta(t, 24, m.Determinant(), 1e-12)

	// A matrix that collapses an axis is singular.
	assert.Equal(t, float64(0), ScaleMatrix4[float64](1, 0, 1).Determinant())
}

func TestMatrix4_InverseOfSingularReportsFailure(t *testing.T) {
	inverse, ok := ScaleMatrix4[float64](1, 0, 1).Inverse()

	assert.False(t, ok)
	assert.Equal(t, Identity4[float64](), inverse)
}

func TestMatrix4_GetScale(t *testing.T) {
	// Scale is recovered from the basis column lengths, so it survives rotation.
	m := RotationMatrix4Z[float64](0.6).Mul(ScaleMatrix4[float64](2, 3, 4))
	scale := m.GetScale()

	assert.InDelta(t, 2, scale.X, 1e-12)
	assert.InDelta(t, 3, scale.Y, 1e-12)
	assert.InDelta(t, 4, scale.Z, 1e-12)
}

func TestMatrix4_WithPosition(t *testing.T) {
	m := ScaleMatrix4[float64](2, 3, 4).WithPosition(Vector3[float64]{1, 2, 3})

	assert.Equal(t, Vector3[float64]{1, 2, 3}, m.GetPosition())
	// The linear part is untouched.
	assert.Equal(t, Vector3[float64]{2, 3, 4}, m.GetScale())
}

func TestMatrix4_Floats32PreservesColumnMajorOrder(t *testing.T) {
	translation := TranslationMatrix4[float64](7, 8, 9)
	floats := translation.Floats32()

	assert.Equal(t, float32(7), floats[12])
	assert.Equal(t, float32(8), floats[13])
	assert.Equal(t, float32(9), floats[14])
	assert.Equal(t, float32(1), floats[15])
}

func TestMatrix4_ApproxEquals(t *testing.T) {
	a := Identity4[float64]()
	b := a
	b[5] += 1e-9

	assert.False(t, a.Equals(b))
	assert.True(t, a.ApproxEquals(b, 1e-6))
	assert.False(t, a.ApproxEquals(b, 1e-12))
}

func TestMatrix4_RotationsAboutEachAxis(t *testing.T) {
	quarter := math.Pi / 2

	// Rotating +Y a quarter turn about X takes it to +Z (counter-clockwise when
	// looking down the axis towards the origin).
	rotationX := RotationMatrix4X[float64](quarter)
	got := rotationX.TransformDirection(Vector3[float64]{0, 1, 0})
	assert.InDelta(t, 0, got.X, 1e-12)
	assert.InDelta(t, 0, got.Y, 1e-12)
	assert.InDelta(t, 1, got.Z, 1e-12)

	// Rotating +Z a quarter turn about Y takes it to +X.
	rotationY := RotationMatrix4Y[float64](quarter)
	got = rotationY.TransformDirection(Vector3[float64]{0, 0, 1})
	assert.InDelta(t, 1, got.X, 1e-12)
	assert.InDelta(t, 0, got.Y, 1e-12)
	assert.InDelta(t, 0, got.Z, 1e-12)

	// Rotating +X a quarter turn about Z takes it to +Y.
	rotationZ := RotationMatrix4Z[float64](quarter)
	got = rotationZ.TransformDirection(Vector3[float64]{1, 0, 0})
	assert.InDelta(t, 0, got.X, 1e-12)
	assert.InDelta(t, 1, got.Y, 1e-12)
	assert.InDelta(t, 0, got.Z, 1e-12)
}

func TestMatrix4_RotationAxisDegenerateAxis(t *testing.T) {
	// A zero-length axis defines no rotation, so the identity is returned rather
	// than a matrix of NaNs.
	assert.Equal(t, Identity4[float64](), RotationMatrix4Axis(Vector3[float64]{}, 1.0))
}

func TestMatrix4_TRS(t *testing.T) {
	// TRS scales first, then rotates, then translates.
	m := TRSMatrix4(
		Vector3[float64]{10, 0, 0},
		RotationMatrix4Z[float64](math.Pi/2),
		Vector3[float64]{2, 2, 2},
	)

	// (1,0,0) scaled to (2,0,0), rotated to (0,2,0), translated to (10,2,0).
	got := m.TransformPoint(Vector3[float64]{1, 0, 0})
	assert.InDelta(t, 10, got.X, 1e-12)
	assert.InDelta(t, 2, got.Y, 1e-12)
	assert.InDelta(t, 0, got.Z, 1e-12)
}

func TestMatrix4_LocalSpaceCombinators(t *testing.T) {
	// Translated/Scaled/RotatedAxis right-multiply, so the composed transform is
	// applied in the receiver's local space rather than the world's.
	base := TranslationMatrix4[float64](10, 0, 0)

	// Translating locally then transforming the origin gives 10+5 along X.
	translated := base.Translated(5, 0, 0)
	assert.Equal(t, Vector3[float64]{15, 0, 0}, translated.TransformPoint(Vector3[float64]{}))
	assert.True(t, translated.ApproxEquals(base.Mul(TranslationMatrix4[float64](5, 0, 0)), 1e-12))

	scaled := base.Scaled(2, 3, 4)
	assert.True(t, scaled.ApproxEquals(base.Mul(ScaleMatrix4[float64](2, 3, 4)), 1e-12))
	assert.Equal(t, Vector3[float64]{12, 3, 4}, scaled.TransformPoint(Vector3[float64]{1, 1, 1}))

	axis := Vector3[float64]{0, 1, 0}
	rotated := base.RotatedAxis(axis, math.Pi/2)
	assert.True(t, rotated.ApproxEquals(base.Mul(RotationMatrix4Axis(axis, math.Pi/2)), 1e-12))

	// Rotating about Y in local space then moving along local +Z: +Z maps to +X.
	got := rotated.TransformPoint(Vector3[float64]{0, 0, 1})
	assert.InDelta(t, 11, got.X, 1e-12)
	assert.InDelta(t, 0, got.Y, 1e-12)
	assert.InDelta(t, 0, got.Z, 1e-12)
}

func TestMatrix4_MultiplyScalarAndString(t *testing.T) {
	doubled := Identity4[float64]().MultiplyScalar(2)
	assert.Equal(t, float64(2), doubled[0])
	assert.Equal(t, float64(2), doubled[15])

	// String renders row-major, so it reads the way the matrix is written.
	assert.Equal(t,
		"Matrix4[[1 0 0 7] [0 1 0 8] [0 0 1 9] [0 0 0 1]]",
		TranslationMatrix4[float64](7, 8, 9).String())
}
