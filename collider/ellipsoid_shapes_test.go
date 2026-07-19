package collider

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wolfgarnet/gomath"
)

func circle64(x, y, r float64) gomath.Circle[float64] {
	return gomath.Circle[float64]{Center: gomath.Vector2[float64]{X: x, Y: y}, Radius: r}
}

func aabb64(x, y, hw, hh float64) gomath.AABB[float64] {
	return gomath.AABB[float64]{Center: gomath.Vector2[float64]{X: x, Y: y}, HalfWidth: hw, HalfHeight: hh}
}

func TestEllipsoidCircle(t *testing.T) {
	e := newEllipsoid64(10, 5, 2, 1) // wide-and-flat, offset

	assert.True(t, EllipsoidCircle(e, circle64(13, 5, 1.01)), "circle just reaches the +X vertex")
	assert.False(t, EllipsoidCircle(e, circle64(13, 5, 0.99)), "circle just short of the +X vertex")
	assert.True(t, EllipsoidCircle(e, circle64(10, 5, 0.1)), "small circle at the centre")
	assert.False(t, EllipsoidCircle(e, circle64(10, 8, 1.5)), "circle clears above the flat axis")
}

func TestEllipsoidAABB(t *testing.T) {
	e := newEllipsoid64(0, 0, 2, 1)

	assert.True(t, EllipsoidAABB(e, aabb64(0, 0, 0.5, 0.5)), "box inside the ellipse")
	assert.True(t, EllipsoidAABB(e, aabb64(2.4, 0, 0.5, 0.5)), "box straddling the +X vertex")
	assert.False(t, EllipsoidAABB(e, aabb64(4, 0, 0.5, 0.5)), "box clear to the +X")
	// The corner test is where a bounding-circle would be wrong: (1.5,0.9) is
	// outside the flat ellipse even though it is within radius 2 of the centre.
	assert.False(t, EllipsoidAABB(e, aabb64(2.0, 1.4, 0.5, 0.5)), "box off the top-right, clear of the flat ellipse")
}

func TestEllipsoidEllipsoidPenetration(t *testing.T) {
	t.Run("separated", func(t *testing.T) {
		_, _, ok := EllipsoidEllipsoidPenetration(newEllipsoid64(0, 0, 2, 1), newEllipsoid64(10, 0, 2, 1))
		assert.False(t, ok)
	})

	t.Run("normal along line of centres", func(t *testing.T) {
		pen, n, ok := EllipsoidEllipsoidPenetration(newEllipsoid64(0, 0, 2, 1), newEllipsoid64(3, 0, 2, 1))
		assert.True(t, ok)
		// Each reaches 2 toward the other along X; centres are 3 apart.
		assert.InDelta(t, 1, pen, 1e-9)
		assert.InDelta(t, 1, n.X, 1e-9)
		assert.InDelta(t, 0, n.Y, 1e-9)
	})

	t.Run("orientation matters: same gap, different axes", func(t *testing.T) {
		// Ellipses 2 wide, 1 tall, centres 2.5 apart. Along X they overlap
		// (reach 2+2 > 2.5); along Y they do not (reach 1+1 < 2.5). A circle
		// model could not tell these apart.
		_, _, okX := EllipsoidEllipsoidPenetration(newEllipsoid64(0, 0, 2, 1), newEllipsoid64(2.5, 0, 2, 1))
		_, _, okY := EllipsoidEllipsoidPenetration(newEllipsoid64(0, 0, 2, 1), newEllipsoid64(0, 2.5, 2, 1))
		assert.True(t, okX, "overlap along the wide axis")
		assert.False(t, okY, "clear along the flat axis")
	})

	t.Run("coincident centres push the short way", func(t *testing.T) {
		pen, n, ok := EllipsoidEllipsoidPenetration(newEllipsoid64(0, 0, 3, 1), newEllipsoid64(0, 0, 3, 1))
		assert.True(t, ok)
		assert.InDelta(t, 2, pen, 1e-9) // b1+b2 = 1+1 < a1+a2 = 6
		assert.InDelta(t, 1, n.Y, 1e-9) // pushed along Y (the short axis)
	})
}

// TestEllipsoidEllipsoidPenetration_AgainstCircle checks the reduction to
// CircleCirclePenetration for circular ellipses over random configurations.
func TestEllipsoidEllipsoidPenetration_AgainstCircle(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 500; i++ {
		r1 := rng.Float64()*3 + 0.5
		r2 := rng.Float64()*3 + 0.5
		c1 := circle64(rng.Float64()*8-4, rng.Float64()*8-4, r1)
		c2 := circle64(rng.Float64()*8-4, rng.Float64()*8-4, r2)
		e1 := gomath.Ellipsoid[float64]{Position: c1.Center, Radii: gomath.Vector2[float64]{X: r1, Y: r1}}
		e2 := gomath.Ellipsoid[float64]{Position: c2.Center, Radii: gomath.Vector2[float64]{X: r2, Y: r2}}

		pc, nc, okc := CircleCirclePenetration(c1, c2)
		pe, ne, oke := EllipsoidEllipsoidPenetration(e1, e2)
		assert.Equalf(t, okc, oke, "overlap i=%d", i)
		if okc {
			assert.InDeltaf(t, pc, pe, 1e-6, "depth i=%d", i)
			assert.InDeltaf(t, nc.X, ne.X, 1e-6, "normal.X i=%d", i)
			assert.InDeltaf(t, nc.Y, ne.Y, 1e-6, "normal.Y i=%d", i)
		}
	}
}

// TestEllipsoidAABBPenetration_AgainstCircle checks the reduction to
// CircleAABBPenetration for circular ellipses over random configurations,
// covering both the outside and centre-inside branches.
func TestEllipsoidAABBPenetration_AgainstCircle(t *testing.T) {
	rng := rand.New(rand.NewSource(8))
	for i := 0; i < 500; i++ {
		r := rng.Float64()*2 + 0.5
		c := circle64(rng.Float64()*6-3, rng.Float64()*6-3, r)
		box := aabb64(rng.Float64()*6-3, rng.Float64()*6-3, rng.Float64()*2+0.5, rng.Float64()*2+0.5)
		e := gomath.Ellipsoid[float64]{Position: c.Center, Radii: gomath.Vector2[float64]{X: r, Y: r}}

		pc, nc, okc := CircleAABBPenetration(c, box)
		pe, ne, oke := EllipsoidAABBPenetration(e, box)
		assert.Equalf(t, okc, oke, "overlap i=%d c=%v box=%v", i, c, box)
		if okc {
			assert.InDeltaf(t, pc, pe, 1e-6, "depth i=%d", i)
			assert.InDeltaf(t, nc.X, ne.X, 1e-6, "normal.X i=%d", i)
			assert.InDeltaf(t, nc.Y, ne.Y, 1e-6, "normal.Y i=%d", i)
		}
	}
}

func TestEllipsoidShapesPenetration_NormalConventionShape1ToShape2(t *testing.T) {
	e := newEllipsoid64(0, 0, 3, 1)
	assertRightNormal := func(name string, n gomath.Vector2[float64], ok bool) {
		assert.Truef(t, ok, "%s: expected overlap", name)
		assert.InDeltaf(t, 1, n.X, 1e-6, "%s: normal.X toward obstacle", name)
		assert.InDeltaf(t, 0, n.Y, 1e-6, "%s: normal.Y", name)
	}

	_, n, ok := EllipsoidEllipsoidPenetration(e, newEllipsoid64(3, 0, 2, 1))
	assertRightNormal("EllipsoidEllipsoid", n, ok)

	_, n, ok = EllipsoidCirclePenetration(e, circle64(3, 0, 1))
	assertRightNormal("EllipsoidCircle", n, ok)

	_, n, ok = EllipsoidAABBPenetration(e, aabb64(3.5, 0, 1, 1))
	assertRightNormal("EllipsoidAABB", n, ok)
}
