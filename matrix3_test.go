package gomath

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatrix3_IdentityAndStorageOrder(t *testing.T) {
	identity := Identity3[float64]()

	assert.Equal(t, identity, NewMatrix3[float64]())
	assert.Equal(t, Vector3[float64]{3, 4, 5}, identity.TransformVector3(Vector3[float64]{3, 4, 5}))

	// NewMatrix3From takes arguments row by row but stores column-major, so
	// element (row r, col c) must land at index c*3+r.
	m := NewMatrix3From[float64](
		1, 2, 3,
		4, 5, 6,
		7, 8, 9,
	)

	assert.Equal(t, float64(1), m.At(0, 0))
	assert.Equal(t, float64(2), m.At(0, 1))
	assert.Equal(t, float64(4), m.At(1, 0))
	assert.Equal(t, float64(9), m.At(2, 2))

	// Column 0 is the first three stored elements.
	assert.Equal(t, [3]float64{1, 4, 7}, [3]float64{m[0], m[1], m[2]})
	assert.Equal(t, Vector3[float64]{1, 4, 7}, m.Column(0))
}

func TestMatrix3_Basis3(t *testing.T) {
	m := Basis3(
		Vector3[float64]{1, 0, 0},
		Vector3[float64]{0, 2, 0},
		Vector3[float64]{0, 0, 3},
	)

	assert.Equal(t, Vector3[float64]{2, 4, 9}, m.TransformVector3(Vector3[float64]{2, 2, 3}))
}

func TestMatrix3_TransposeAndDeterminant(t *testing.T) {
	m := NewMatrix3From[float64](
		1, 2, 3,
		4, 5, 6,
		7, 8, 10,
	)

	transposed := m.Transpose()
	assert.Equal(t, m, transposed.Transpose())
	assert.Equal(t, m.At(1, 2), transposed.At(2, 1))

	// Determinant is unchanged by transposition.
	assert.InDelta(t, m.Determinant(), transposed.Determinant(), 1e-12)
	assert.Equal(t, float64(1), Identity3[float64]().Determinant())

	// A collapsed axis is singular.
	assert.Equal(t, float64(0), Basis3(
		Vector3[float64]{1, 0, 0},
		Vector3[float64]{0, 0, 0},
		Vector3[float64]{0, 0, 1},
	).Determinant())
}

func TestMatrix3_InverseOfSingularReportsFailure(t *testing.T) {
	singular := Basis3(
		Vector3[float64]{1, 0, 0},
		Vector3[float64]{0, 0, 0},
		Vector3[float64]{0, 0, 1},
	)

	inverse, ok := singular.Inverse()

	assert.False(t, ok)
	assert.Equal(t, Identity3[float64](), inverse)
}

func TestMatrix3_InverseUndoesTransform(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	identity := Identity3[float64]()

	for i := 0; i < 200; i++ {
		rotation := RotationMatrix4Axis(Vector3[float64]{
			X: rng.Float64()*2 - 1,
			Y: rng.Float64()*2 - 1,
			Z: rng.Float64()*2 - 1,
		}.Normalize(), (rng.Float64()*2-1)*math.Pi)

		scale := ScaleMatrix4(rng.Float64()*3+0.5, rng.Float64()*3+0.5, rng.Float64()*3+0.5)
		composed := rotation.Mul(scale)
		m := composed.Basis()

		inverse, ok := m.Inverse()
		require.True(t, ok)

		assert.True(t, m.Mul(inverse).ApproxEquals(identity, 1e-9))
		assert.True(t, inverse.Mul(m).ApproxEquals(identity, 1e-9))
	}
}

func TestMatrix3_Floats32PreservesColumnMajorOrder(t *testing.T) {
	m := NewMatrix3From[float64](
		1, 2, 3,
		4, 5, 6,
		7, 8, 9,
	)

	floats := m.Floats32()

	// Column-major: index 1 is row 1 column 0, which is 4.
	assert.Equal(t, float32(1), floats[0])
	assert.Equal(t, float32(4), floats[1])
	assert.Equal(t, float32(7), floats[2])
	assert.Equal(t, float32(2), floats[3])
}

func TestMatrix3_ApproxEquals(t *testing.T) {
	a := Identity3[float64]()
	b := a
	b[4] += 1e-9

	assert.False(t, a.Equals(b))
	assert.True(t, a.ApproxEquals(b, 1e-6))
	assert.False(t, a.ApproxEquals(b, 1e-12))
}

func TestMatrix4_BasisDropsTranslation(t *testing.T) {
	m := TranslationMatrix4[float64](100, 200, 300).Mul(ScaleMatrix4[float64](2, 3, 4))

	basis := m.Basis()

	// The scale survives, the translation does not.
	assert.Equal(t, Vector3[float64]{2, 6, 12}, basis.TransformVector3(Vector3[float64]{1, 2, 3}))
}

func TestMatrix4_NormalMatrixUnderUniformScaleIsJustTheRotation(t *testing.T) {
	// Under rotation and uniform scale the inverse-transpose is the rotation
	// again (up to the uniform factor), so normals could have used the transform
	// directly. This is the case that hides the bug.
	rotation := RotationMatrix4Y[float64](0.7)

	normalMatrix, ok := rotation.NormalMatrix()
	require.True(t, ok)

	normal := Vector3[float64]{0, 0, 1}
	viaNormalMatrix := normalMatrix.TransformVector3(normal)
	viaTransform := rotation.TransformDirection(normal)

	assert.InDelta(t, viaTransform.X, viaNormalMatrix.X, 1e-12)
	assert.InDelta(t, viaTransform.Y, viaNormalMatrix.Y, 1e-12)
	assert.InDelta(t, viaTransform.Z, viaNormalMatrix.Z, 1e-12)
}

func TestMatrix4_NormalMatrixKeepsNormalsPerpendicularUnderNonUniformScale(t *testing.T) {
	// The property that justifies the whole type: after a non-uniform scale, a
	// normal transformed by the normal matrix stays perpendicular to a tangent
	// transformed by the model matrix. Transforming the normal by the model
	// matrix instead does NOT, which is checked at the end.
	rng := rand.New(rand.NewSource(12))

	for i := 0; i < 200; i++ {
		model := RotationMatrix4Axis(Vector3[float64]{
			X: rng.Float64()*2 - 1,
			Y: rng.Float64()*2 - 1,
			Z: rng.Float64()*2 - 1,
		}.Normalize(), (rng.Float64()*2-1)*math.Pi).
			Mul(ScaleMatrix4(rng.Float64()*4+0.5, rng.Float64()*4+0.5, rng.Float64()*4+0.5))

		normalMatrix, ok := model.NormalMatrix()
		require.True(t, ok)

		// Build a tangent and a normal that start out perpendicular.
		normal := Vector3[float64]{
			X: rng.Float64()*2 - 1,
			Y: rng.Float64()*2 - 1,
			Z: rng.Float64()*2 - 1,
		}.Normalize()
		if normal.IsZero() {
			continue
		}

		arbitrary := Vector3[float64]{1, 0, 0}
		if math.Abs(normal.Dot(arbitrary)) > 0.9 {
			arbitrary = Vector3[float64]{0, 1, 0}
		}
		tangent := normal.Cross(arbitrary).Normalize()
		require.InDelta(t, 0, normal.Dot(tangent), 1e-12, "setup: must start perpendicular")

		transformedTangent := model.TransformDirection(tangent)
		transformedNormal := normalMatrix.TransformVector3(normal)

		assert.InDelta(t, 0,
			transformedNormal.Normalize().Dot(transformedTangent.Normalize()),
			1e-9,
			"normal matrix must keep the normal perpendicular to the surface")
	}
}

func TestMatrix4_NormalMatrixIsNecessary(t *testing.T) {
	// Demonstrates that the normal matrix is not busywork: under a strong
	// non-uniform scale, using the model matrix on a normal tilts it away from
	// perpendicular, while the normal matrix keeps it correct.
	model := ScaleMatrix4[float64](1, 8, 1)

	normal := Vector3[float64]{1, 1, 0}.Normalize()
	tangent := Vector3[float64]{1, -1, 0}.Normalize() // perpendicular to normal
	require.InDelta(t, 0, normal.Dot(tangent), 1e-12)

	transformedTangent := model.TransformDirection(tangent)

	normalMatrix, ok := model.NormalMatrix()
	require.True(t, ok)

	correct := normalMatrix.TransformVector3(normal).Normalize()
	naive := model.TransformDirection(normal).Normalize()

	assert.InDelta(t, 0, correct.Dot(transformedTangent.Normalize()), 1e-9,
		"normal matrix keeps it perpendicular")
	assert.Greater(t, math.Abs(naive.Dot(transformedTangent.Normalize())), 0.1,
		"using the model matrix on a normal should visibly break perpendicularity")
}

func TestMatrix4_NormalMatrixOfSingularBasisReportsFailure(t *testing.T) {
	collapsed := ScaleMatrix4[float64](1, 0, 1)
	normalMatrix, ok := collapsed.NormalMatrix()

	assert.False(t, ok)
	assert.Equal(t, Identity3[float64](), normalMatrix)
}
