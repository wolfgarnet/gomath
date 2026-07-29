package gomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuaternion_Identity(t *testing.T) {
	identity := IdentityQuaternion[float64]()

	assert.Equal(t, Quaternion[float64]{0, 0, 0, 1}, identity)
	assert.Equal(t, float64(1), identity.Length())

	// The identity leaves a vector alone.
	v := Vector3[float64]{1, 2, 3}
	assert.Equal(t, v, identity.RotateVector3(v))
}

func TestQuaternion_FromAxisAngleQuarterTurns(t *testing.T) {
	quarter := math.Pi / 2

	// Same expectations as the matrix rotations: +Y about X goes to +Z, and so on.
	got := QuaternionFromAxisAngle(Vector3[float64]{X: 1}, quarter).
		RotateVector3(Vector3[float64]{0, 1, 0})
	assert.InDelta(t, 0, got.X, 1e-12)
	assert.InDelta(t, 0, got.Y, 1e-12)
	assert.InDelta(t, 1, got.Z, 1e-12)

	got = QuaternionFromAxisAngle(Vector3[float64]{Y: 1}, quarter).
		RotateVector3(Vector3[float64]{0, 0, 1})
	assert.InDelta(t, 1, got.X, 1e-12)
	assert.InDelta(t, 0, got.Y, 1e-12)
	assert.InDelta(t, 0, got.Z, 1e-12)

	got = QuaternionFromAxisAngle(Vector3[float64]{Z: 1}, quarter).
		RotateVector3(Vector3[float64]{1, 0, 0})
	assert.InDelta(t, 0, got.X, 1e-12)
	assert.InDelta(t, 1, got.Y, 1e-12)
	assert.InDelta(t, 0, got.Z, 1e-12)
}

func TestQuaternion_FromAxisAngleDegenerateAxis(t *testing.T) {
	// A zero-length axis defines no rotation.
	assert.Equal(t, IdentityQuaternion[float64](),
		QuaternionFromAxisAngle(Vector3[float64]{}, 1.0))

	// The axis need not be unit length; it is normalised first.
	long := QuaternionFromAxisAngle(Vector3[float64]{Z: 17}, 0.9)
	unit := QuaternionFromAxisAngle(Vector3[float64]{Z: 1}, 0.9)
	assert.True(t, long.ApproxEquals(unit, 1e-12))
}

func TestQuaternion_ConjugateAndInverse(t *testing.T) {
	q := QuaternionFromAxisAngle(Vector3[float64]{1, 2, 3}.Normalize(), 0.8)

	// For a unit quaternion the inverse is the conjugate.
	assert.True(t, q.Inverse().ApproxEquals(q.Conjugate(), 1e-12))

	// q composed with its inverse is the identity rotation.
	assert.True(t, q.Mul(q.Inverse()).ApproxEqualRotation(IdentityQuaternion[float64](), 1e-12))

	// A zero quaternion has no inverse; the identity is returned instead.
	assert.Equal(t, IdentityQuaternion[float64](), Quaternion[float64]{}.Inverse())
}

func TestQuaternion_NormalizeOfZeroIsIdentity(t *testing.T) {
	assert.Equal(t, IdentityQuaternion[float64](), Quaternion[float64]{}.Normalize())
	assert.InDelta(t, 1, Quaternion[float64]{1, 2, 3, 4}.Normalize().Length(), 1e-12)
}

func TestQuaternion_SlerpEndpointsAreExact(t *testing.T) {
	a := QuaternionFromAxisAngle(Vector3[float64]{Y: 1}, 0.3)
	b := QuaternionFromAxisAngle(Vector3[float64]{Y: 1}, 2.1)

	// Exact, not merely close: callers rely on the endpoints being untouched.
	assert.Equal(t, a, a.Slerp(b, 0))
	assert.Equal(t, b, a.Slerp(b, 1))

	// Out-of-range t clamps to the endpoints.
	assert.Equal(t, a, a.Slerp(b, -0.5))
	assert.Equal(t, b, a.Slerp(b, 1.5))
}

func TestQuaternion_SlerpMidpointHalvesTheAngle(t *testing.T) {
	a := IdentityQuaternion[float64]()
	b := QuaternionFromAxisAngle(Vector3[float64]{Y: 1}, math.Pi/2)

	middle := a.Slerp(b, 0.5)
	want := QuaternionFromAxisAngle(Vector3[float64]{Y: 1}, math.Pi/4)

	assert.True(t, middle.ApproxEqualRotation(want, 1e-12),
		"halfway between 0 and 90 degrees should be 45, got %v want %v", middle, want)
}

func TestQuaternion_SlerpTakesTheShortPath(t *testing.T) {
	a := QuaternionFromAxisAngle(Vector3[float64]{Y: 1}, 0)
	// 350 degrees the long way is 10 degrees the short way.
	b := QuaternionFromAxisAngle(Vector3[float64]{Y: 1}, DegToRad(350.0))

	middle := a.Slerp(b, 0.5)
	_, angle := middle.ToAxisAngle()

	// The midpoint must be 5 degrees from the start, not 175.
	assert.InDelta(t, DegToRad(5.0), angle, 1e-9)
}

func TestQuaternion_SlerpOfIdenticalRotations(t *testing.T) {
	// A degenerate arc must not divide by a vanishing sine.
	q := QuaternionFromAxisAngle(Vector3[float64]{Z: 1}, 0.4)

	middle := q.Slerp(q, 0.5)

	assert.False(t, math.IsNaN(middle.X))
	assert.True(t, middle.ApproxEqualRotation(q, 1e-12))
}

func TestQuaternion_ToAxisAngleRoundTrip(t *testing.T) {
	axis := Vector3[float64]{1, -2, 3}.Normalize()
	const angle = 1.1

	gotAxis, gotAngle := QuaternionFromAxisAngle(axis, angle).ToAxisAngle()

	assert.InDelta(t, angle, gotAngle, 1e-12)
	assert.InDelta(t, axis.X, gotAxis.X, 1e-12)
	assert.InDelta(t, axis.Y, gotAxis.Y, 1e-12)
	assert.InDelta(t, axis.Z, gotAxis.Z, 1e-12)

	// The identity has no meaningful axis, but must still return a usable one.
	identityAxis, identityAngle := IdentityQuaternion[float64]().ToAxisAngle()
	assert.Equal(t, float64(0), identityAngle)
	assert.InDelta(t, 1, identityAxis.Length(), 1e-12)
}

func TestQuaternion_FromUnitVectors(t *testing.T) {
	from := Vector3[float64]{1, 0, 0}
	to := Vector3[float64]{0, 1, 0}

	rotated := QuaternionFromUnitVectors(from, to).RotateVector3(from)

	assert.InDelta(t, to.X, rotated.X, 1e-12)
	assert.InDelta(t, to.Y, rotated.Y, 1e-12)
	assert.InDelta(t, to.Z, rotated.Z, 1e-12)
}

func TestQuaternion_FromUnitVectorsOpposite(t *testing.T) {
	// Opposite vectors have no unique shortest rotation, but any 180-degree
	// rotation about a perpendicular axis is correct.
	from := Vector3[float64]{1, 0, 0}
	to := Vector3[float64]{-1, 0, 0}

	rotated := QuaternionFromUnitVectors(from, to).RotateVector3(from)

	assert.InDelta(t, to.X, rotated.X, 1e-9)
	assert.InDelta(t, to.Y, rotated.Y, 1e-9)
	assert.InDelta(t, to.Z, rotated.Z, 1e-9)
}

func TestQuaternion_AngleTo(t *testing.T) {
	a := IdentityQuaternion[float64]()
	b := QuaternionFromAxisAngle(Vector3[float64]{Y: 1}, math.Pi/3)

	assert.InDelta(t, math.Pi/3, a.AngleTo(b), 1e-9)
	assert.InDelta(t, 0, a.AngleTo(a), 1e-9)
}

func TestQuaternion_EqualsVersusEqualRotation(t *testing.T) {
	q := QuaternionFromAxisAngle(Vector3[float64]{Z: 1}, 0.7)
	negated := q.MultiplyScalar(-1)

	// q and -q are the same rotation but not the same components. This is the
	// trap ApproxEqualRotation exists to avoid.
	assert.False(t, q.Equals(negated))
	assert.False(t, q.ApproxEquals(negated, 1e-9))
	assert.True(t, q.ApproxEqualRotation(negated, 1e-9))

	v := Vector3[float64]{1, 2, 3}
	rotatedByQ := q.RotateVector3(v)
	rotatedByNegated := negated.RotateVector3(v)
	assert.InDelta(t, rotatedByQ.X, rotatedByNegated.X, 1e-12)
	assert.InDelta(t, rotatedByQ.Y, rotatedByNegated.Y, 1e-12)
	assert.InDelta(t, rotatedByQ.Z, rotatedByNegated.Z, 1e-12)
}

func TestQuaternion_String(t *testing.T) {
	assert.Equal(t, "Quaternion(0, 0, 0, 1)", IdentityQuaternion[float64]().String())
}

func TestScalarHelpers(t *testing.T) {
	assert.InDelta(t, math.Pi, DegToRad(180.0), 1e-12)
	assert.InDelta(t, 180, RadToDeg(math.Pi), 1e-12)
	assert.InDelta(t, math.Pi/6, Asin(0.5), 1e-12)
	assert.InDelta(t, math.Pi/4, Atan(1.0), 1e-12)

	assert.Equal(t, 5.0, Lerp(0.0, 10.0, 0.5))
	assert.Equal(t, 0.0, Lerp(0.0, 10.0, 0.0))
	assert.Equal(t, 10.0, Lerp(0.0, 10.0, 1.0))
	// t is not clamped, so it extrapolates.
	assert.Equal(t, 20.0, Lerp(0.0, 10.0, 2.0))

	assert.Equal(t, -1, Sign(-7))
	assert.Equal(t, 1, Sign(7))
	assert.Equal(t, 0, Sign(0))

	assert.InDelta(t, 1, Tan(math.Pi/4), 1e-12)
}

func TestQuaternion_ComposeAndDecompose(t *testing.T) {
	translation := Vector3[float64]{10, -5, 3}
	rotation := QuaternionFromAxisAngle(Vector3[float64]{1, 2, 3}.Normalize(), 0.9)
	scale := Vector3[float64]{2, 3, 4}

	m := ComposeMatrix4(translation, rotation, scale)

	// Compose must agree with the matrix-based TRS builder from Phase 1.
	viaTRS := TRSMatrix4(translation, rotation.ToMatrix4(), scale)
	assert.True(t, m.ApproxEquals(viaTRS, 1e-12),
		"ComposeMatrix4 should equal TRSMatrix4 with the same rotation")

	gotTranslation, gotRotation, gotScale, ok := m.Decompose()
	require.True(t, ok)

	assert.InDelta(t, translation.X, gotTranslation.X, 1e-12)
	assert.InDelta(t, translation.Y, gotTranslation.Y, 1e-12)
	assert.InDelta(t, translation.Z, gotTranslation.Z, 1e-12)

	assert.InDelta(t, scale.X, gotScale.X, 1e-12)
	assert.InDelta(t, scale.Y, gotScale.Y, 1e-12)
	assert.InDelta(t, scale.Z, gotScale.Z, 1e-12)

	assert.True(t, gotRotation.ApproxEqualRotation(rotation, 1e-9),
		"got %v want %v", gotRotation, rotation)
}

func TestQuaternion_DecomposeOfSingularBasisReportsFailure(t *testing.T) {
	collapsed := ScaleMatrix4[float64](1, 0, 1)

	_, rotation, _, ok := collapsed.Decompose()

	assert.False(t, ok)
	assert.Equal(t, IdentityQuaternion[float64](), rotation)
}

func TestQuaternion_DecomposeAttributesReflectionToX(t *testing.T) {
	// A negative scale is a reflection, which no rotation can express, so the
	// sign is folded into scale.X by convention.
	m := ComposeMatrix4(
		Vector3[float64]{},
		IdentityQuaternion[float64](),
		Vector3[float64]{-2, 3, 4},
	)

	_, _, scale, ok := m.Decompose()
	require.True(t, ok)

	assert.InDelta(t, -2, scale.X, 1e-12)
	assert.InDelta(t, 3, scale.Y, 1e-12)
	assert.InDelta(t, 4, scale.Z, 1e-12)
}

func TestQuaternion_ConstructorAndDegenerateUnitVectors(t *testing.T) {
	assert.Equal(t, Quaternion[float64]{1, 2, 3, 4}, NewQuaternion[float64](1, 2, 3, 4))

	// A zero vector has no direction, so no rotation can be derived from it.
	assert.Equal(t, IdentityQuaternion[float64](),
		QuaternionFromUnitVectors(Vector3[float64]{}, Vector3[float64]{Y: 1}))
	assert.Equal(t, IdentityQuaternion[float64](),
		QuaternionFromUnitVectors(Vector3[float64]{X: 1}, Vector3[float64]{}))

	// Identical vectors need no rotation.
	same := QuaternionFromUnitVectors(Vector3[float64]{X: 1}, Vector3[float64]{X: 1})
	assert.True(t, same.ApproxEqualRotation(IdentityQuaternion[float64](), 1e-12))
}
