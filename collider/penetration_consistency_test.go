package collider

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wolfgarnet/gomath"
)

// The *WithPenetration functions that expose a shape1->shape2 MTV normal must
// agree on that convention: with shape1 at the origin and shape2 (the obstacle)
// placed to the +X side, the normal points toward +X. This mirrors the
// collider package's TestPenetration_NormalConventionIsConsistent so the whole
// library shares one MTV convention.
func TestPenetration_NormalConventionShape1ToShape2(t *testing.T) {
	const eps = 1e-4

	assertRight := func(t *testing.T, name string, n gomath.Vector2[float64], overlapping bool) {
		t.Helper()
		assert.True(t, overlapping, "%s: expected overlap", name)
		assert.InDeltaf(t, 1, n.X, eps, "%s: normal.X (want +1, toward obstacle)", name)
		assert.InDeltaf(t, 0, n.Y, eps, "%s: normal.Y", name)
	}

	// Circle (origin) vs Circle (to the right).
	_, n, o := CircleCircleWithPenetration(
		gomath.Circle[float64]{Center: gomath.Vector2[float64]{X: 0, Y: 0}, Radius: 1},
		gomath.Circle[float64]{Center: gomath.Vector2[float64]{X: 1.5, Y: 0}, Radius: 1},
	)
	assertRight(t, "CircleCircleWithPenetration", n, o)

	// Circle (origin) vs point (to the right).
	_, n, o = CirclePointWithPenetration(
		gomath.Circle[float64]{Center: gomath.Vector2[float64]{X: 0, Y: 0}, Radius: 2},
		gomath.Vector2[float64]{X: 1, Y: 0},
	)
	assertRight(t, "CirclePointWithPenetration", n, o)

	// AABB (origin) vs AABB (to the right).
	_, n, o = AABB2DAABB2DWithPenetration(
		gomath.AABB[float64]{Center: gomath.Vector2[float64]{X: 0, Y: 0}, HalfWidth: 1, HalfHeight: 1},
		gomath.AABB[float64]{Center: gomath.Vector2[float64]{X: 1.5, Y: 0}, HalfWidth: 1, HalfHeight: 1},
	)
	assertRight(t, "AABB2DAABB2DWithPenetration", n, o)
}
