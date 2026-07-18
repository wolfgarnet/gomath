package gomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatrix_Constructors(t *testing.T) {
	// TranslationMatrix moves a point.
	tm := TranslationMatrix[float64](5, -3)
	assert.Equal(t, Vector2[float64]{6, -1}, tm.TransformPoint(Vector2[float64]{1, 2}))
	// ...but leaves a direction (no translation) alone.
	assert.Equal(t, Vector2[float64]{1, 2}, tm.TransformDirection(Vector2[float64]{1, 2}))

	// RotationMatrix rotates counter-clockwise (matches Vector2.Rotate).
	rm := RotationMatrix[float64](math.Pi / 2)
	got := rm.TransformPoint(Vector2[float64]{1, 0})
	assert.InDelta(t, 0, got.X, 1e-9)
	assert.InDelta(t, 1, got.Y, 1e-9)

	// ScaleMatrix scales each axis.
	sm := ScaleMatrix[float64](2, 3)
	assert.Equal(t, Vector2[float64]{4, 9}, sm.TransformPoint(Vector2[float64]{2, 3}))

	// ShearMatrix shears.
	sh := ShearMatrix[float64](1, 0) // x += y
	assert.Equal(t, Vector2[float64]{5, 3}, sh.TransformPoint(Vector2[float64]{2, 3}))
}

func TestMatrix_TransformPointMatchesApplyMatrix(t *testing.T) {
	m := TranslationMatrix[float64](3, -2).Mul(RotationMatrix[float64](0.6)).Mul(ScaleMatrix[float64](2, 1.5))
	v := Vector2[float64]{4, -1}
	assert.Equal(t, v.ApplyMatrix(m), m.TransformPoint(v))
}

func TestMatrix_MulMatchesMatrixMultiply(t *testing.T) {
	a := RotationMatrix[float64](math.Pi / 3)
	b := TranslationMatrix[float64](2, 5)
	want := MatrixMultiply(&a, &b)
	assert.Equal(t, want, a.Mul(b))
}

func TestMatrix_TRS(t *testing.T) {
	m := TRSMatrix(Vector2[float64]{5, 6}, math.Pi/2, Vector2[float64]{2, 3})

	// Origin maps to the translation.
	origin := m.TransformPoint(Vector2[float64]{0, 0})
	assert.InDelta(t, 5, origin.X, 1e-9)
	assert.InDelta(t, 6, origin.Y, 1e-9)

	// (1,0): scale by 2 -> (2,0), rotate 90 -> (0,2), translate -> (5,8).
	p := m.TransformPoint(Vector2[float64]{1, 0})
	assert.InDelta(t, 5, p.X, 1e-9)
	assert.InDelta(t, 8, p.Y, 1e-9)
}

func TestMatrix_Inverse(t *testing.T) {
	m := TRSMatrix(Vector2[float64]{3, -4}, 0.7, Vector2[float64]{2, 1.5})
	inv, ok := m.Inverse()
	assert.True(t, ok)

	// m * inv == identity.
	assertMatrixInDelta(t, NewMatrix[float64](), m.Mul(inv))

	// Round-trip a point.
	v := Vector2[float64]{7, -2}
	assert.InDelta(t, v.X, inv.TransformPoint(m.TransformPoint(v)).X, 1e-9)
	assert.InDelta(t, v.Y, inv.TransformPoint(m.TransformPoint(v)).Y, 1e-9)

	// Singular matrix reports ok == false.
	_, ok = ScaleMatrix[float64](0, 0).Inverse()
	assert.False(t, ok)
}

func TestMatrix_GetScale(t *testing.T) {
	assert.InDelta(t, 2, ScaleMatrix[float64](2, 3).GetScale().X, 1e-9)
	assert.InDelta(t, 3, ScaleMatrix[float64](2, 3).GetScale().Y, 1e-9)

	// Rotation does not change the extracted scale magnitudes.
	rs := RotationMatrix[float64](0.9).Mul(ScaleMatrix[float64](2, 3))
	assert.InDelta(t, 2, rs.GetScale().X, 1e-9)
	assert.InDelta(t, 3, rs.GetScale().Y, 1e-9)
}

func TestMatrix_EqualsApprox(t *testing.T) {
	a := RotationMatrix[float64](0.5)
	b := RotationMatrix[float64](0.5)
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(RotationMatrix[float64](0.6)))

	c := a
	c[0] += 1e-10
	assert.False(t, a.Equals(c))
	assert.True(t, a.ApproxEquals(c, 1e-6))
}

func TestMatrix_FluentBuilders(t *testing.T) {
	// Fluent builders right-multiply: m.Translated(x,y) == m * Translation(x,y),
	// i.e. the new transform is applied in the matrix's local space (first).
	base := RotationMatrix[float64](math.Pi / 2)

	assert.Equal(t, base.Mul(TranslationMatrix[float64](2, 3)), base.Translated(2, 3))
	assert.Equal(t, base.Mul(RotationMatrix[float64](0.4)), base.Rotated(0.4))
	assert.Equal(t, base.Mul(ScaleMatrix[float64](2, 3)), base.Scaled(2, 3))

	// Local-space translation then rotation: translate (1,0) in local X, then the
	// 90-degree rotation carries it to +Y.
	p := NewMatrix[float64]()
	p = p.Rotated(math.Pi / 2).Translated(1, 0)
	got := p.TransformPoint(Vector2[float64]{0, 0})
	assert.InDelta(t, 0, got.X, 1e-9)
	assert.InDelta(t, 1, got.Y, 1e-9)
}

func TestMatrix_String(t *testing.T) {
	assert.NotEmpty(t, NewMatrix[float64]().String())
}
