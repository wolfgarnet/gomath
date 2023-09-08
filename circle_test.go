package gomath

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestCircle_TangentLinePoints(t *testing.T) {
	tests := []struct {
		center      Vector2[float32]
		radius      float32
		point       Vector2[float32]
		t1, t2      Vector2[float32]
		hasSolution bool
	}{
		{
			center:      Vector2[float32]{0, 0},
			radius:      1,
			point:       Vector2[float32]{2, 0},
			t1:          Vector2[float32]{0.5, 0.8660254},
			t2:          Vector2[float32]{0.5, -0.8660254},
			hasSolution: true,
		},
		{
			center:      Vector2[float32]{0, 0},
			radius:      1,
			point:       Vector2[float32]{0.5, 0},
			t1:          Vector2[float32]{0.0, 0.0},
			t2:          Vector2[float32]{0.0, 0.0},
			hasSolution: false,
		},
		{
			center:      Vector2[float32]{0, 0},
			radius:      1,
			point:       Vector2[float32]{10, 0},
			t1:          Vector2[float32]{0.10000001, 0.9949874},
			t2:          Vector2[float32]{0.10000001, -0.9949874},
			hasSolution: true,
		},
	}

	for i, test := range tests {
		circle := Circle[float32]{
			Center: test.center,
			Radius: test.radius,
		}
		t1, t2, hasSolution := circle.TangentLinePoints(test.point)
		require.Equalf(t, test.hasSolution, hasSolution, "Failed test %v", i)
		if !hasSolution {
			continue
		}

		assert.Equalf(t, test.t1, t1, "Failed test %v", i)
		assert.Equalf(t, test.t2, t2, "Failed test %v", i)
	}
}
