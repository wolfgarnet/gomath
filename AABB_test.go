package gomath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSquaredAABB2D_GetCenter pins that GetCenter returns the centre in the
// box's own numeric type. It previously recursed into itself unconditionally.
func TestSquaredAABB2D_GetCenter(t *testing.T) {
	t.Run("float64", func(t *testing.T) {
		box := NewSquaredAABB2D(Vector2[float64]{3, -4}, 2)
		assert.Equal(t, Vector2[float64]{3, -4}, box.GetCenter())
	})

	t.Run("float32", func(t *testing.T) {
		box := NewSquaredAABB2DXY[float32](1, 2, 5)
		assert.Equal(t, Vector2[float32]{1, 2}, box.GetCenter())
	})

	t.Run("zero value", func(t *testing.T) {
		var box SquaredAABB2D[float64]
		assert.Equal(t, Vector2[float64]{0, 0}, box.GetCenter())
	})
}
