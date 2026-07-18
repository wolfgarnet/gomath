package collider

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wolfgarnet/gomath"
)

func seg64(x1, y1, x2, y2 float64) gomath.LineSegment[float64] {
	return gomath.LineSegment[float64]{
		P1: gomath.Vector2[float64]{X: x1, Y: y1},
		P2: gomath.Vector2[float64]{X: x2, Y: y2},
	}
}

// implicit is the ellipse's implicit value at p: <0 inside, 0 on the boundary,
// >0 outside.
func implicit(e gomath.Ellipsoid[float64], p gomath.Vector2[float64]) float64 {
	dx := (p.X - e.Position.X) / e.Radii.X
	dy := (p.Y - e.Position.Y) / e.Radii.Y
	return dx*dx + dy*dy - 1
}

func TestEllipsoidPoint(t *testing.T) {
	// Offset ellipse: catches the old bug where Position was ignored.
	e := newEllipsoid64(10, 5, 2, 1)

	tests := map[string]struct {
		p    gomath.Vector2[float64]
		want bool
	}{
		"centre":             {gomath.Vector2[float64]{X: 10, Y: 5}, true},
		"inside":             {gomath.Vector2[float64]{X: 11, Y: 5}, true},
		"on boundary +x":     {gomath.Vector2[float64]{X: 12, Y: 5}, true},
		"on boundary +y":     {gomath.Vector2[float64]{X: 10, Y: 6}, true},
		"just outside +x":    {gomath.Vector2[float64]{X: 12.01, Y: 5}, false},
		"outside diagonal":   {gomath.Vector2[float64]{X: 12, Y: 6}, false},
		"world origin (far)": {gomath.Vector2[float64]{X: 0, Y: 0}, false},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, EllipsoidPoint(e, tc.p))
		})
	}
}

func TestEllipsoidLineSegment(t *testing.T) {
	e := newEllipsoid64(10, 5, 2, 1)

	tests := map[string]struct {
		s    gomath.LineSegment[float64]
		want bool
	}{
		"crosses horizontally":    {seg64(5, 5, 15, 5), true},
		"fully inside":            {seg64(9, 5, 11, 5), true},
		"endpoint inside":         {seg64(10, 5, 100, 100), true},
		"tangent to top":          {seg64(0, 6, 20, 6), true},
		"clears above":            {seg64(0, 6.5, 20, 6.5), false},
		"clears to the side":      {seg64(13, 0, 13, 10), false},
		"parallel just outside x": {seg64(12.01, 0, 12.01, 10), false},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, EllipsoidLineSegment(e, tc.s))
		})
	}
}

func TestRayEllipsoid(t *testing.T) {
	e := newEllipsoid64(0, 0, 2, 1)

	t.Run("through", func(t *testing.T) {
		n, p1, p2 := RayEllipsoid(gomath.NewRay(
			gomath.Vector2[float64]{X: -5, Y: 0},
			gomath.Vector2[float64]{X: 1, Y: 0},
		), e)
		assert.Equal(t, 2, n)
		assert.InDelta(t, 2, p1.X, 1e-9)  // far
		assert.InDelta(t, -2, p2.X, 1e-9) // near
	})

	t.Run("origin inside yields only exit", func(t *testing.T) {
		n, p1, _ := RayEllipsoid(gomath.NewRay(
			gomath.Vector2[float64]{X: 0, Y: 0},
			gomath.Vector2[float64]{X: 1, Y: 0},
		), e)
		assert.Equal(t, 1, n)
		assert.InDelta(t, 2, p1.X, 1e-9)
		assert.InDelta(t, 0, p1.Y, 1e-9)
	})

	t.Run("tangent to top", func(t *testing.T) {
		n, p1, _ := RayEllipsoid(gomath.NewRay(
			gomath.Vector2[float64]{X: -5, Y: 1},
			gomath.Vector2[float64]{X: 1, Y: 0},
		), e)
		assert.Equal(t, 1, n)
		assert.InDelta(t, 0, p1.X, 1e-9)
		assert.InDelta(t, 1, p1.Y, 1e-9)
	})

	t.Run("miss", func(t *testing.T) {
		n, _, _ := RayEllipsoid(gomath.NewRay(
			gomath.Vector2[float64]{X: -5, Y: 5},
			gomath.Vector2[float64]{X: 1, Y: 0},
		), e)
		assert.Equal(t, 0, n)
	})

	t.Run("entirely behind", func(t *testing.T) {
		n, _, _ := RayEllipsoid(gomath.NewRay(
			gomath.Vector2[float64]{X: 5, Y: 0},
			gomath.Vector2[float64]{X: 1, Y: 0},
		), e)
		assert.Equal(t, 0, n)
	})

	t.Run("hit points are independent of direction magnitude", func(t *testing.T) {
		origin := gomath.Vector2[float64]{X: -5, Y: 0}
		nUnit, farUnit, nearUnit := RayEllipsoid(gomath.NewRay(origin, gomath.Vector2[float64]{X: 1, Y: 0}), e)
		nScaled, farScaled, nearScaled := RayEllipsoid(gomath.NewRay(origin, gomath.Vector2[float64]{X: 7, Y: 0}), e)
		assert.Equal(t, nUnit, nScaled)
		assert.InDelta(t, farUnit.X, farScaled.X, 1e-9)
		assert.InDelta(t, nearUnit.X, nearScaled.X, 1e-9)
	})
}

// TestRayEllipsoid_AgainstRayCircle checks that on a circular ellipse RayEllipsoid
// agrees with the existing RayCircle over random rays.
func TestRayEllipsoid_AgainstRayCircle(t *testing.T) {
	rng := rand.New(rand.NewSource(20260718))
	for i := 0; i < 500; i++ {
		r := rng.Float64()*3 + 0.5
		centre := gomath.Vector2[float64]{X: rng.Float64()*10 - 5, Y: rng.Float64()*10 - 5}
		circle := gomath.Circle[float64]{Center: centre, Radius: r}
		ell := gomath.Ellipsoid[float64]{Position: centre, Radii: gomath.Vector2[float64]{X: r, Y: r}}

		origin := gomath.Vector2[float64]{X: rng.Float64()*20 - 10, Y: rng.Float64()*20 - 10}
		ang := rng.Float64() * 2 * math.Pi
		dir := gomath.Vector2[float64]{X: math.Cos(ang), Y: math.Sin(ang)}
		ray := gomath.NewRay(origin, dir)

		nc, fc, ncr := RayCircle(ray, circle)
		ne, fe, nee := RayEllipsoid(ray, ell)

		assert.Equalf(t, nc, ne, "count mismatch ray=%v circle=%v", ray, circle)
		if nc >= 1 {
			assert.InDeltaf(t, fc.X, fe.X, 1e-6, "far.X ray=%v", ray)
			assert.InDeltaf(t, fc.Y, fe.Y, 1e-6, "far.Y ray=%v", ray)
		}
		if nc == 2 {
			assert.InDeltaf(t, ncr.X, nee.X, 1e-6, "near.X ray=%v", ray)
			assert.InDeltaf(t, ncr.Y, nee.Y, 1e-6, "near.Y ray=%v", ray)
		}
	}
}

func TestEllipsoidPointPenetration(t *testing.T) {
	e := newEllipsoid64(10, 5, 2, 1)

	t.Run("outside does not overlap", func(t *testing.T) {
		_, _, ok := EllipsoidPointPenetration(e, gomath.Vector2[float64]{X: 13, Y: 5})
		assert.False(t, ok)
	})

	t.Run("on boundary does not overlap", func(t *testing.T) {
		_, _, ok := EllipsoidPointPenetration(e, gomath.Vector2[float64]{X: 12, Y: 5})
		assert.False(t, ok)
	})

	t.Run("centre pushes out along the minor axis", func(t *testing.T) {
		pen, normal, ok := EllipsoidPointPenetration(e, gomath.Vector2[float64]{X: 10, Y: 5})
		assert.True(t, ok)
		// Nearest boundary from the centre is the minor axis (b=1) direction.
		assert.InDelta(t, 1, pen, 1e-9)
		assert.InDelta(t, 0, normal.X, 1e-9)
		assert.InDelta(t, 1, math.Abs(normal.Y), 1e-9)
	})

	t.Run("MTV lands the point on the boundary", func(t *testing.T) {
		p := gomath.Vector2[float64]{X: 10.5, Y: 5.2}
		pen, normal, ok := EllipsoidPointPenetration(e, p)
		assert.True(t, ok)
		moved := p.Add(normal.MultiplyScalar(pen))
		assert.InDelta(t, 0, implicit(e, moved), 1e-6)
	})
}

// TestEllipsoidPointPenetration_AgainstCircle checks the reduction to
// CirclePointWithPenetration on a circular ellipse.
func TestEllipsoidPointPenetration_AgainstCircle(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	centre := gomath.Vector2[float64]{X: 3, Y: -2}
	r := 4.0
	circle := gomath.Circle[float64]{Center: centre, Radius: r}
	ell := gomath.Ellipsoid[float64]{Position: centre, Radii: gomath.Vector2[float64]{X: r, Y: r}}

	for i := 0; i < 300; i++ {
		p := gomath.Vector2[float64]{X: centre.X + rng.Float64()*6 - 3, Y: centre.Y + rng.Float64()*6 - 3}
		pc, nc, okc := CirclePointWithPenetration(circle, p)
		pe, ne, oke := EllipsoidPointPenetration(ell, p)
		assert.Equalf(t, okc, oke, "overlap mismatch p=%v", p)
		if okc {
			assert.InDeltaf(t, pc, pe, 1e-6, "depth p=%v", p)
			assert.InDeltaf(t, nc.X, ne.X, 1e-6, "normal.X p=%v", p)
			assert.InDeltaf(t, nc.Y, ne.Y, 1e-6, "normal.Y p=%v", p)
		}
	}
}

func TestEllipsoidLineSegmentPenetration(t *testing.T) {
	e := newEllipsoid64(0, 0, 2, 1)

	t.Run("clearing segment does not overlap", func(t *testing.T) {
		_, _, ok := EllipsoidLineSegmentPenetration(e, seg64(-5, 3, 5, 3))
		assert.False(t, ok)
	})

	t.Run("MTV lands the deepest point on the boundary", func(t *testing.T) {
		s := seg64(-5, 0.3, 5, 0.3)
		pen, normal, ok := EllipsoidLineSegmentPenetration(e, s)
		assert.True(t, ok)
		// The deepest point (in ellipse metric) plus the MTV must reach the boundary.
		// Recover it the same way the function does and verify it lands on the ellipse.
		t0 := closestParamToOrigin(s.P1.X/2, s.P1.Y/1, s.P2.X/2, s.P2.Y/1)
		deep := gomath.Vector2[float64]{X: s.P1.X + t0*(s.P2.X-s.P1.X), Y: s.P1.Y + t0*(s.P2.Y-s.P1.Y)}
		moved := deep.Add(normal.MultiplyScalar(pen))
		assert.InDelta(t, 0, implicit(e, moved), 1e-6)
	})
}

// TestEllipsoidLineSegmentPenetration_AgainstCircle checks the reduction to
// CircleLineSegmentPenetration on a circular ellipse.
func TestEllipsoidLineSegmentPenetration_AgainstCircle(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	centre := gomath.Vector2[float64]{X: -1, Y: 4}
	r := 3.0
	circle := gomath.Circle[float64]{Center: centre, Radius: r}
	ell := gomath.Ellipsoid[float64]{Position: centre, Radii: gomath.Vector2[float64]{X: r, Y: r}}

	for i := 0; i < 400; i++ {
		s := gomath.LineSegment[float64]{
			P1: gomath.Vector2[float64]{X: centre.X + rng.Float64()*8 - 4, Y: centre.Y + rng.Float64()*8 - 4},
			P2: gomath.Vector2[float64]{X: centre.X + rng.Float64()*8 - 4, Y: centre.Y + rng.Float64()*8 - 4},
		}
		pc, nc, okc := CircleLineSegmentPenetration(circle, s)
		pe, ne, oke := EllipsoidLineSegmentPenetration(ell, s)
		assert.Equalf(t, okc, oke, "overlap mismatch s=%v", s)
		if okc {
			assert.InDeltaf(t, pc, pe, 1e-6, "depth s=%v", s)
			assert.InDeltaf(t, nc.X, ne.X, 1e-6, "normal.X s=%v", s)
			assert.InDeltaf(t, nc.Y, ne.Y, 1e-6, "normal.Y s=%v", s)
		}
	}
}

// TestClosestPointOnEllipse_AgainstBruteForce cross-checks the exact Eberly
// closest-point against a dense boundary sampler, for query points inside and
// outside, across circular and elongated ellipses in every quadrant.
func TestClosestPointOnEllipse_AgainstBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(20260719))
	const samples = 200000

	bruteForce := func(a, b, px, py float64) (bx, by, best float64) {
		best = math.Inf(1)
		for i := 0; i < samples; i++ {
			theta := 2 * math.Pi * float64(i) / samples
			x := a * math.Cos(theta)
			y := b * math.Sin(theta)
			d := math.Hypot(x-px, y-py)
			if d < best {
				best, bx, by = d, x, y
			}
		}
		return
	}

	for i := 0; i < 60; i++ {
		a := rng.Float64()*4 + 0.5
		b := rng.Float64()*4 + 0.5
		px := rng.Float64()*12 - 6
		py := rng.Float64()*12 - 6

		gotX, gotY := closestPointOnEllipse(a, b, px, py)

		// The returned point must lie on the ellipse.
		onEllipse := (gotX/a)*(gotX/a) + (gotY/b)*(gotY/b)
		assert.InDeltaf(t, 1, onEllipse, 1e-6, "closest point off the ellipse a=%v b=%v p=(%v,%v)", a, b, px, py)

		gotDist := math.Hypot(gotX-px, gotY-py)
		_, _, wantDist := bruteForce(a, b, px, py)
		// The sampler can only get within one step of the true minimum.
		assert.InDeltaf(t, wantDist, gotDist, 5e-3, "distance a=%v b=%v p=(%v,%v)", a, b, px, py)
		// Eberly must be at least as good as the discrete sampler.
		assert.LessOrEqualf(t, gotDist, wantDist+1e-6, "Eberly worse than sampler a=%v b=%v p=(%v,%v)", a, b, px, py)
	}
}

// TestEllipsoidPenetration_NormalConventionShape1ToShape2 anchors the new
// penetration functions to the library-wide MTV convention: with the ellipse at
// the origin and the obstacle to the +X side, the normal points toward +X.
func TestEllipsoidPenetration_NormalConventionShape1ToShape2(t *testing.T) {
	e := newEllipsoid64(0, 0, 3, 1)

	// Past x = (a^2-b^2)/a = 8/3 the nearest boundary is the +X vertex, so the
	// MTV is unambiguously +X. (Closer to the centre it would point at the nearer
	// top/bottom boundary instead -- a real geometric fact, not a convention.)
	_, n, ok := EllipsoidPointPenetration(e, gomath.Vector2[float64]{X: 2.8, Y: 0})
	assert.True(t, ok)
	assert.InDelta(t, 1, n.X, 1e-6, "point: normal.X toward obstacle")
	assert.InDelta(t, 0, n.Y, 1e-6, "point: normal.Y")

	// A short segment sitting to the +X side, deepest point on the major axis.
	_, n, ok = EllipsoidLineSegmentPenetration(e, seg64(2.8, -0.1, 2.8, 0.1))
	assert.True(t, ok)
	assert.InDelta(t, 1, n.X, 1e-6, "segment: normal.X toward obstacle")
	assert.InDelta(t, 0, n.Y, 1e-6, "segment: normal.Y")
}
