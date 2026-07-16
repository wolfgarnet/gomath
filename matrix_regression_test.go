package gomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func rotationMatrix(theta float64) Matrix[float64] {
	m := NewMatrix[float64]()
	m.Rotation(theta)
	return m
}

func translationMatrix(x, y float64) Matrix[float64] {
	m := NewMatrix[float64]()
	m.SetPosition(x, y)
	return m
}

func assertMatrixInDelta(t *testing.T, want, got Matrix[float64]) {
	t.Helper()
	for i := range want {
		assert.InDeltaf(t, want[i], got[i], 0.0001, "element %d: want %v, got %v", i, want, got)
	}
}

// TestMatrix_MultiplyMatchesMatrixMultiply asserts the in-place Multiply agrees
// with the value-returning MatrixMultiply, which computes into a separate result.
func TestMatrix_MultiplyMatchesMatrixMultiply(t *testing.T) {
	tests := map[string]struct {
		a Matrix[float64]
		b Matrix[float64]
	}{
		"identity x identity":    {a: NewMatrix[float64](), b: NewMatrix[float64]()},
		"identity x rotation":    {a: NewMatrix[float64](), b: rotationMatrix(math.Pi / 4)},
		"rotation x identity":    {a: rotationMatrix(math.Pi / 4), b: NewMatrix[float64]()},
		"rotation x rotation":    {a: rotationMatrix(math.Pi / 4), b: rotationMatrix(math.Pi / 3)},
		"translation x rotation": {a: translationMatrix(3, -2), b: rotationMatrix(math.Pi / 6)},
		"rotation x translation": {a: rotationMatrix(math.Pi / 6), b: translationMatrix(3, -2)},
		"arbitrary x arbitrary":  {a: Matrix[float64]{1, 2, 3, 4, 5, 6, 7, 8, 10}, b: Matrix[float64]{9, 8, 7, 6, 5, 4, 3, 2, 1}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			want := MatrixMultiply(&test.a, &test.b)

			got := test.a
			got.Multiply(&test.b)

			assertMatrixInDelta(t, want, got)
		})
	}
}

// TestMatrix_MultiplyByIdentity asserts multiplying by the identity is a no-op.
func TestMatrix_MultiplyByIdentity(t *testing.T) {
	identity := NewMatrix[float64]()
	original := Matrix[float64]{1, 2, 3, 4, 5, 6, 7, 8, 10}

	got := original
	got.Multiply(&identity)

	assertMatrixInDelta(t, original, got)
}

// TestMatrix_RotationComposition asserts rotating by a then b equals rotating by a+b.
func TestMatrix_RotationComposition(t *testing.T) {
	a := rotationMatrix(math.Pi / 6)
	b := rotationMatrix(math.Pi / 4)

	got := a
	got.Multiply(&b)

	assertMatrixInDelta(t, rotationMatrix(math.Pi/6+math.Pi/4), got)
}

func TestMatrix_GetInverse(t *testing.T) {
	tests := map[string]struct {
		src Matrix[float64]
	}{
		"identity":    {src: NewMatrix[float64]()},
		"rotation":    {src: rotationMatrix(math.Pi / 3)},
		"translation": {src: translationMatrix(5, -7)},
		"arbitrary":   {src: Matrix[float64]{2, 0, 0, 0, 3, 0, 4, 5, 1}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var inverse Matrix[float64]
			err := inverse.GetInverse(test.src)
			assert.NoError(t, err)

			// src * inverse must be the identity.
			product := MatrixMultiply(&test.src, &inverse)
			assertMatrixInDelta(t, NewMatrix[float64](), product)
		})
	}
}

// TestMatrix_GetInverseSingular asserts a singular matrix reports an error
// rather than silently returning a garbage matrix.
func TestMatrix_GetInverseSingular(t *testing.T) {
	singular := Matrix[float64]{} // all zeroes, determinant 0

	var inverse Matrix[float64]
	err := inverse.GetInverse(singular)

	assert.Error(t, err)
	assertMatrixInDelta(t, NewMatrix[float64](), inverse)
}
