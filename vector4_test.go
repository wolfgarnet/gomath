package gomath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVector4_Arithmetic(t *testing.T) {
	a := NewVector4[float64](1, 2, 3, 4)
	b := NewVector4[float64](10, 20, 30, 40)

	assert.Equal(t, Vector4[float64]{11, 22, 33, 44}, a.Add(b))
	assert.Equal(t, Vector4[float64]{9, 18, 27, 36}, b.Sub(a))
	assert.Equal(t, Vector4[float64]{10, 40, 90, 160}, a.Mul(b))
	assert.Equal(t, Vector4[float64]{2, 4, 6, 8}, a.MultiplyScalar(2))
	assert.Equal(t, Vector4[float64]{10, 10, 10, 10}, b.Div(a))
	assert.Equal(t, Vector4[float64]{2, 3, 4, 5}, a.AddScalar(1))
	assert.Equal(t, Vector4[float64]{0, 1, 2, 3}, a.SubScalar(1))
	assert.Equal(t, Vector4[float64]{0.5, 1, 1.5, 2}, a.DivScalar(2))

	// Operations are non-mutating: a is unchanged by all of the above.
	assert.Equal(t, Vector4[float64]{1, 2, 3, 4}, a)
}

func TestVector4_DotAndLength(t *testing.T) {
	a := NewVector4[float64](1, 2, 3, 4)
	b := NewVector4[float64](5, 6, 7, 8)

	assert.Equal(t, float64(70), a.Dot(b)) // 5 + 12 + 21 + 32

	// A 3-4-0-0 vector has length 5; the W component participates like any other.
	assert.Equal(t, float64(25), Vector4[float64]{3, 4, 0, 0}.LengthSq())
	assert.Equal(t, float64(5), Vector4[float64]{3, 4, 0, 0}.Length())
	assert.InDelta(t, 2, Vector4[float64]{1, 1, 1, 1}.Length(), 1e-12)
}

func TestVector4_Normalize(t *testing.T) {
	normalised := Vector4[float64]{0, 0, 0, 5}.Normalize()
	assert.Equal(t, Vector4[float64]{0, 0, 0, 1}, normalised)

	// A zero vector has no direction; normalising it yields zero rather than NaN.
	assert.Equal(t, Vector4[float64]{}, Vector4[float64]{}.Normalize())
}

func TestVector4_Lerp(t *testing.T) {
	a := NewVector4[float64](0, 0, 0, 0)
	b := NewVector4[float64](10, 20, 30, 40)

	assert.Equal(t, a, a.Lerp(b, 0))
	assert.Equal(t, b, a.Lerp(b, 1))
	assert.Equal(t, Vector4[float64]{5, 10, 15, 20}, a.Lerp(b, 0.5))
}

func TestVector4_Component(t *testing.T) {
	v := NewVector4[float64](1, 2, 3, 4)

	for i, want := range []float64{1, 2, 3, 4} {
		assert.Equal(t, want, v.Component(i))
	}

	assert.Panics(t, func() { v.Component(4) })
	assert.Panics(t, func() { v.Component(-1) })
}

func TestVector4_ToVector3AndPerspectiveDivide(t *testing.T) {
	// ToVector3 simply drops W.
	assert.Equal(t, Vector3[float64]{2, 4, 6}, Vector4[float64]{2, 4, 6, 2}.ToVector3())

	// PerspectiveDivide divides through by W, which is what turns a clip-space
	// position into normalised device coordinates.
	assert.Equal(t, Vector3[float64]{1, 2, 3}, Vector4[float64]{2, 4, 6, 2}.PerspectiveDivide())

	// W of zero has no finite projection, so XYZ is returned untouched rather
	// than as infinities.
	assert.Equal(t, Vector3[float64]{2, 4, 6}, Vector4[float64]{2, 4, 6, 0}.PerspectiveDivide())
}

func TestVector4_ToVector4FromVector3(t *testing.T) {
	// A point carries W=1 so that it picks up translation; a direction carries
	// W=0 so that it does not.
	assert.Equal(t, Vector4[float64]{1, 2, 3, 1}, Vector3[float64]{1, 2, 3}.ToVector4(1))
	assert.Equal(t, Vector4[float64]{1, 2, 3, 0}, Vector3[float64]{1, 2, 3}.ToVector4(0))
}

func TestVector4_String(t *testing.T) {
	assert.Equal(t, "Vector4(1, 2, 3, 4)", NewVector4[float64](1, 2, 3, 4).String())
}
