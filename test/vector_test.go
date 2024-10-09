package test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
	math32 "github.com/wolfgarnet/gomath/g3n"
	"testing"
)

func Test_Vector(t *testing.T) {
	v1 := gomath.Vector2[float32]{X: 1, Y: 2}
	v2 := gomath.Vector2[float32]{X: 3, Y: 4}

	v1 = v1.Add(v2)
	v1 = v1.Mul(v2)
	v1 = v1.MultiplyScalar(2)
	assert.Equal(t, gomath.Vector2[float32]{X: 24, Y: 48}, v1)
}

func Test_G3nVector(t *testing.T) {
	v1 := math32.Vector2{X: 1, Y: 2}
	v2 := math32.Vector2{X: 3, Y: 4}

	v1.Add(&v2)
	v1.Multiply(&v2)
	v1.MultiplyScalar(2)
	assert.Equal(t, math32.Vector2{X: 24, Y: 48}, v1)

	// Make a more elaborate test using the g3n math32 vector library

}

func Benchmark_Vector(b *testing.B) {
	b.Run("gomath", func(b *testing.B) {
		v1 := gomath.Vector2[float32]{X: 1, Y: 2}
		v2 := gomath.Vector2[float32]{X: 3, Y: 4}
		for i := 0; i < b.N; i++ {
			v1 = v1.Add(v2)
			v1 = v1.Mul(v2)
			v1 = v1.MultiplyScalar(2)
		}
	})

	b.Run("g3n", func(b *testing.B) {
		v1 := math32.Vector2{X: 1, Y: 2}
		v2 := math32.Vector2{X: 3, Y: 4}
		for i := 0; i < b.N; i++ {
			v1.Add(&v2)
			v1.Multiply(&v2)
			v1.MultiplyScalar(2)
		}
	})
}
