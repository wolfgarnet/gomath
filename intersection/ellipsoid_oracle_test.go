package intersection

import (
	"math"
	"math/rand"
	"testing"

	"github.com/wolfgarnet/gomath"
)

// insideness reports the value of the ellipse's implicit function at p.
// It is negative strictly inside, zero on the boundary and positive outside.
func insideness(p gomath.Vector2[float64], e gomath.Ellipsoid[float64]) float64 {
	dx := (p.X - e.Position.X) / e.Radii.X
	dy := (p.Y - e.Position.Y) / e.Radii.Y
	return dx*dx + dy*dy - 1
}

// bruteForceOverlap is an independent, deliberately naive oracle: it densely
// samples each ellipse's boundary and asks whether any sample falls inside the
// other, plus a containment check for the case where one swallows the other.
// It returns the overlap verdict along with how close the call was, so callers
// can skip near-tangent cases the sampling cannot resolve.
func bruteForceOverlap(e1, e2 gomath.Ellipsoid[float64]) (overlap bool, margin float64) {
	const samples = 20000

	minOn2 := math.Inf(1) // insideness of e2's boundary w.r.t. e1
	minOn1 := math.Inf(1) // insideness of e1's boundary w.r.t. e2

	for i := 0; i < samples; i++ {
		theta := 2 * math.Pi * float64(i) / samples
		cos, sin := math.Cos(theta), math.Sin(theta)

		on2 := gomath.Vector2[float64]{
			X: e2.Position.X + e2.Radii.X*cos,
			Y: e2.Position.Y + e2.Radii.Y*sin,
		}
		minOn2 = math.Min(minOn2, insideness(on2, e1))

		on1 := gomath.Vector2[float64]{
			X: e1.Position.X + e1.Radii.X*cos,
			Y: e1.Position.Y + e1.Radii.Y*sin,
		}
		minOn1 = math.Min(minOn1, insideness(on1, e2))
	}

	// One fully containing the other: no boundary crossing, but they do overlap.
	if insideness(e2.Position, e1) <= 0 || insideness(e1.Position, e2) <= 0 {
		return true, math.Abs(math.Min(minOn1, minOn2))
	}

	best := math.Min(minOn1, minOn2)
	return best <= 0, math.Abs(best)
}

// TestEllipsoidEllipsoid_AgainstBruteForce cross-checks EllipsoidEllipsoid against
// an independent sampling oracle over random axis-aligned ellipses. Near-tangent
// configurations are skipped, since the oracle cannot resolve them either.
func TestEllipsoidEllipsoid_AgainstBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(20260716))

	const (
		cases     = 400
		ambiguous = 1e-3
	)

	checked := 0
	for i := 0; i < cases; i++ {
		e1 := gomath.Ellipsoid[float64]{
			Position: gomath.Vector2[float64]{X: rng.Float64()*20 - 10, Y: rng.Float64()*20 - 10},
			Radii:    gomath.Vector2[float64]{X: rng.Float64()*4 + 0.25, Y: rng.Float64()*4 + 0.25},
		}
		e2 := gomath.Ellipsoid[float64]{
			Position: gomath.Vector2[float64]{X: rng.Float64()*20 - 10, Y: rng.Float64()*20 - 10},
			Radii:    gomath.Vector2[float64]{X: rng.Float64()*4 + 0.25, Y: rng.Float64()*4 + 0.25},
		}

		want, margin := bruteForceOverlap(e1, e2)
		if margin < ambiguous {
			continue // too close to tangency for the oracle to be authoritative
		}
		checked++

		if got := EllipsoidEllipsoid(e1, e2); got != want {
			t.Errorf("e1=%v e2=%v: got %v, oracle says %v (margin %v)", e1, e2, got, want, margin)
		}

		// The predicate is symmetric.
		if got := EllipsoidEllipsoid(e2, e1); got != want {
			t.Errorf("swapped e1=%v e2=%v: got %v, oracle says %v", e2, e1, got, want)
		}
	}

	t.Logf("cross-checked %d/%d random ellipse pairs against the oracle", checked, cases)
	if checked < cases/2 {
		t.Fatalf("too few resolvable cases (%d); the generator or oracle is suspect", checked)
	}
}

// TestEllipsoidEllipsoid_Containment covers the case the boundary-sampling oracle
// needs a special path for: one ellipse entirely swallowing the other.
func TestEllipsoidEllipsoid_Containment(t *testing.T) {
	big := newEllipsoid64(5, 5, 10, 10)
	small := newEllipsoid64(5, 5, 1, 1)

	if !EllipsoidEllipsoid(big, small) {
		t.Error("a big ellipse containing a small one must overlap it")
	}
	if !EllipsoidEllipsoid(small, big) {
		t.Error("containment must be symmetric")
	}
}
