package gomath

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Property tests pinning down Quaternion's semantics. The most valuable ones
// here cross-check against code gomath already had — Vector3.RotateAxis and the
// Phase 1 matrix builders — so agreement is real evidence about handedness and
// layout rather than the new code agreeing with itself.

func randomUnitAxis(rng *rand.Rand) (Vector3[float64], bool) {
	axis := Vector3[float64]{
		X: rng.Float64()*2 - 1,
		Y: rng.Float64()*2 - 1,
		Z: rng.Float64()*2 - 1,
	}.Normalize()

	return axis, !axis.IsZero()
}

func TestQuaternionSemantics_AgreesWithRotateAxisAndRotationMatrix(t *testing.T) {
	// The three-way handedness check: QuaternionFromAxisAngle,
	// Vector3.RotateAxis and RotationMatrix4Axis are three independent
	// implementations of the same rotation and must all agree.
	rng := rand.New(rand.NewSource(21))

	for i := 0; i < 300; i++ {
		axis, ok := randomUnitAxis(rng)
		if !ok {
			continue
		}

		phi := (rng.Float64()*2 - 1) * math.Pi
		v := Vector3[float64]{rng.Float64() * 10, rng.Float64() * 10, rng.Float64() * 10}

		viaVector := v.RotateAxis(axis, phi)

		rotationMatrix := RotationMatrix4Axis(axis, phi)
		viaMatrix := rotationMatrix.TransformDirection(v)

		viaQuaternion := QuaternionFromAxisAngle(axis, phi).RotateVector3(v)

		assert.InDelta(t, viaVector.X, viaQuaternion.X, 1e-9)
		assert.InDelta(t, viaVector.Y, viaQuaternion.Y, 1e-9)
		assert.InDelta(t, viaVector.Z, viaQuaternion.Z, 1e-9)

		assert.InDelta(t, viaMatrix.X, viaQuaternion.X, 1e-9)
		assert.InDelta(t, viaMatrix.Y, viaQuaternion.Y, 1e-9)
		assert.InDelta(t, viaMatrix.Z, viaQuaternion.Z, 1e-9)
	}
}

func TestQuaternionSemantics_ToMatrix4AgreesWithRotationMatrix4Axis(t *testing.T) {
	rng := rand.New(rand.NewSource(22))

	for i := 0; i < 200; i++ {
		axis, ok := randomUnitAxis(rng)
		if !ok {
			continue
		}

		phi := (rng.Float64()*2 - 1) * math.Pi

		fromQuaternion := QuaternionFromAxisAngle(axis, phi).ToMatrix4()
		fromBuilder := RotationMatrix4Axis(axis, phi)

		assert.True(t, fromQuaternion.ApproxEquals(fromBuilder, 1e-9),
			"quaternion matrix %v should equal builder %v", fromQuaternion, fromBuilder)
	}
}

func TestQuaternionSemantics_ToMatrix3MatchesToMatrix4Basis(t *testing.T) {
	rng := rand.New(rand.NewSource(23))

	for i := 0; i < 100; i++ {
		axis, ok := randomUnitAxis(rng)
		if !ok {
			continue
		}

		q := QuaternionFromAxisAngle(axis, (rng.Float64()*2-1)*math.Pi)

		asMatrix4 := q.ToMatrix4()

		assert.True(t, q.ToMatrix3().ApproxEquals(asMatrix4.Basis(), 1e-12))
	}
}

func TestQuaternionSemantics_MatrixRoundTrip(t *testing.T) {
	// Every branch of Shepperd's method must be exercised, which random
	// rotations do: the branch taken depends on which diagonal term is largest.
	rng := rand.New(rand.NewSource(24))

	for i := 0; i < 500; i++ {
		axis, ok := randomUnitAxis(rng)
		if !ok {
			continue
		}

		q := QuaternionFromAxisAngle(axis, (rng.Float64()*2-1)*math.Pi)

		recovered := QuaternionFromRotationMatrix(q.ToMatrix4())

		assert.True(t, recovered.ApproxEqualRotation(q, 1e-9),
			"round trip failed: got %v want %v", recovered, q)
	}
}

func TestQuaternionSemantics_MatrixRoundTripAtHalfTurns(t *testing.T) {
	// A rotation of exactly pi puts the trace at its minimum and forces
	// Shepperd's non-trace branches, which is where a naive conversion divides
	// by zero. Check each principal axis explicitly rather than hoping the
	// random sampler lands there.
	for _, axis := range []Vector3[float64]{
		{X: 1}, {Y: 1}, {Z: 1},
		{X: 1, Y: 1}, {Y: 1, Z: 1}, {X: 1, Z: 1},
	} {
		q := QuaternionFromAxisAngle(axis.Normalize(), math.Pi)

		recovered := QuaternionFromRotationMatrix(q.ToMatrix4())

		assert.True(t, recovered.ApproxEqualRotation(q, 1e-9),
			"half turn about %v: got %v want %v", axis, recovered, q)
	}
}

func TestQuaternionSemantics_MulComposesLikeMatrixMul(t *testing.T) {
	// q.Mul(b) must mean the same thing as q's matrix times b's matrix, or
	// composition order silently disagrees between the two representations.
	rng := rand.New(rand.NewSource(25))

	for i := 0; i < 200; i++ {
		axisA, okA := randomUnitAxis(rng)
		axisB, okB := randomUnitAxis(rng)
		if !okA || !okB {
			continue
		}

		a := QuaternionFromAxisAngle(axisA, (rng.Float64()*2-1)*math.Pi)
		b := QuaternionFromAxisAngle(axisB, (rng.Float64()*2-1)*math.Pi)

		viaQuaternion := a.Mul(b).ToMatrix4()
		viaMatrix := a.ToMatrix4().Mul(b.ToMatrix4())

		assert.True(t, viaQuaternion.ApproxEquals(viaMatrix, 1e-9))

		// And applying the composition must equal applying b then a.
		v := Vector3[float64]{1, -2, 3}
		assert.InDelta(t,
			a.RotateVector3(b.RotateVector3(v)).X,
			a.Mul(b).RotateVector3(v).X,
			1e-9)
	}
}

func TestQuaternionSemantics_RotationPreservesLengthAndAngles(t *testing.T) {
	rng := rand.New(rand.NewSource(26))

	for i := 0; i < 200; i++ {
		axis, ok := randomUnitAxis(rng)
		if !ok {
			continue
		}

		q := QuaternionFromAxisAngle(axis, (rng.Float64()*2-1)*math.Pi)

		a := Vector3[float64]{1, 2, 3}
		b := Vector3[float64]{-4, 5, 0.5}

		assert.InDelta(t, a.Length(), q.RotateVector3(a).Length(), 1e-9)
		// A rotation is rigid, so it preserves the angle between two vectors.
		assert.InDelta(t, a.Dot(b), q.RotateVector3(a).Dot(q.RotateVector3(b)), 1e-9)
	}
}

func TestQuaternionSemantics_EulerXYZMatchesMatrixComposition(t *testing.T) {
	// The order is in the name; this is what pins it down.
	rng := rand.New(rand.NewSource(27))

	for i := 0; i < 200; i++ {
		x := (rng.Float64()*2 - 1) * math.Pi
		y := (rng.Float64()*2 - 1) * math.Pi
		z := (rng.Float64()*2 - 1) * math.Pi

		fromEuler := QuaternionFromEulerXYZ(x, y, z).ToMatrix4()
		fromMatrices := RotationMatrix4X(x).
			Mul(RotationMatrix4Y(y)).
			Mul(RotationMatrix4Z(z))

		assert.True(t, fromEuler.ApproxEquals(fromMatrices, 1e-9),
			"euler %v %v %v: got %v want %v", x, y, z, fromEuler, fromMatrices)
	}
}

func TestQuaternionSemantics_SlerpStaysOnTheUnitSphere(t *testing.T) {
	// Interpolation must produce rotations at every step. This is precisely
	// what interpolating matrix elements fails to do.
	rng := rand.New(rand.NewSource(28))

	for i := 0; i < 100; i++ {
		axisA, okA := randomUnitAxis(rng)
		axisB, okB := randomUnitAxis(rng)
		if !okA || !okB {
			continue
		}

		a := QuaternionFromAxisAngle(axisA, (rng.Float64()*2-1)*math.Pi)
		b := QuaternionFromAxisAngle(axisB, (rng.Float64()*2-1)*math.Pi)

		for step := 0; step <= 10; step++ {
			t2 := float64(step) / 10
			interpolated := a.Slerp(b, t2)

			assert.InDelta(t, 1, interpolated.Length(), 1e-9,
				"slerp at t=%v left the unit sphere", t2)
		}
	}
}

func TestQuaternionSemantics_SlerpAngleAdvancesUniformly(t *testing.T) {
	// The defining property of slerp over nlerp: equal steps in t produce equal
	// steps in angle.
	a := IdentityQuaternion[float64]()
	b := QuaternionFromAxisAngle(Vector3[float64]{Y: 1}, DegToRad(120.0))

	previous := 0.0
	for step := 1; step <= 10; step++ {
		t2 := float64(step) / 10
		angle := a.AngleTo(a.Slerp(b, t2))

		assert.InDelta(t, DegToRad(12.0), angle-previous, 1e-6,
			"step %d should advance the angle by a constant 12 degrees", step)
		previous = angle
	}
}

func TestQuaternionSemantics_ComposeDecomposeRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(29))

	for i := 0; i < 300; i++ {
		axis, ok := randomUnitAxis(rng)
		if !ok {
			continue
		}

		translation := Vector3[float64]{
			X: (rng.Float64()*2 - 1) * 100,
			Y: (rng.Float64()*2 - 1) * 100,
			Z: (rng.Float64()*2 - 1) * 100,
		}
		rotation := QuaternionFromAxisAngle(axis, (rng.Float64()*2-1)*math.Pi)
		scale := Vector3[float64]{
			X: rng.Float64()*4 + 0.25,
			Y: rng.Float64()*4 + 0.25,
			Z: rng.Float64()*4 + 0.25,
		}

		m := ComposeMatrix4(translation, rotation, scale)

		gotTranslation, gotRotation, gotScale, ok := m.Decompose()
		require.True(t, ok)

		assert.InDelta(t, translation.X, gotTranslation.X, 1e-9)
		assert.InDelta(t, translation.Y, gotTranslation.Y, 1e-9)
		assert.InDelta(t, translation.Z, gotTranslation.Z, 1e-9)

		assert.InDelta(t, scale.X, gotScale.X, 1e-9)
		assert.InDelta(t, scale.Y, gotScale.Y, 1e-9)
		assert.InDelta(t, scale.Z, gotScale.Z, 1e-9)

		assert.True(t, gotRotation.ApproxEqualRotation(rotation, 1e-8),
			"rotation round trip: got %v want %v", gotRotation, rotation)

		// The strongest form: recomposing must reproduce the matrix.
		assert.True(t, ComposeMatrix4(gotTranslation, gotRotation, gotScale).
			ApproxEquals(m, 1e-9))
	}
}

func TestQuaternionSemantics_ComposeMatchesTransformingByHand(t *testing.T) {
	// Compose is written out rather than built from three matrix products, so
	// check it against the long way round on random inputs.
	rng := rand.New(rand.NewSource(30))

	for i := 0; i < 200; i++ {
		axis, ok := randomUnitAxis(rng)
		if !ok {
			continue
		}

		translation := Vector3[float64]{rng.Float64() * 10, rng.Float64() * 10, rng.Float64() * 10}
		rotation := QuaternionFromAxisAngle(axis, (rng.Float64()*2-1)*math.Pi)
		scale := Vector3[float64]{rng.Float64()*3 + 0.5, rng.Float64()*3 + 0.5, rng.Float64()*3 + 0.5}

		composed := ComposeMatrix4(translation, rotation, scale)

		point := Vector3[float64]{1, -2, 3}

		// Scale, then rotate, then translate — by hand.
		byHand := rotation.RotateVector3(Vector3[float64]{
			X: point.X * scale.X,
			Y: point.Y * scale.Y,
			Z: point.Z * scale.Z,
		}).Add(translation)

		got := composed.TransformPoint(point)

		assert.InDelta(t, byHand.X, got.X, 1e-9)
		assert.InDelta(t, byHand.Y, got.Y, 1e-9)
		assert.InDelta(t, byHand.Z, got.Z, 1e-9)
	}
}

func TestQuaternionSemantics_NormalMatrixOfPureRotationIsTheRotation(t *testing.T) {
	// A rotation basis is orthonormal, so its inverse-transpose is itself. This
	// ties Phase 1's NormalMatrix to Phase 2's quaternions.
	rng := rand.New(rand.NewSource(31))

	for i := 0; i < 100; i++ {
		axis, ok := randomUnitAxis(rng)
		if !ok {
			continue
		}

		rotation := QuaternionFromAxisAngle(axis, (rng.Float64()*2-1)*math.Pi)
		model := rotation.ToMatrix4()

		normalMatrix, ok := model.NormalMatrix()
		require.True(t, ok)

		assert.True(t, normalMatrix.ApproxEquals(rotation.ToMatrix3(), 1e-9))
	}
}
