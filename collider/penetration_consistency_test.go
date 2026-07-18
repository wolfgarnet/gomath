package collider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// All *Penetration functions must share one normal convention: the returned
// normal points from the first shape (the mover, at the origin here) toward the
// second shape (the obstacle, placed to the +X side). pixelmaster relies on this
// shared convention — its per-pair normal reversals are written assuming every
// function agrees with CircleCircle (the dominant, working creep-creep path).
//
// This pins that agreement so CircleLineSegment / AABBAABB can't drift back to
// their previously-flipped normals.
func TestPenetration_NormalConventionIsConsistent(t *testing.T) {
	const eps = 1e-4

	assertRight := func(t *testing.T, name string, nx, ny float32, overlapping bool) {
		t.Helper()
		assert.True(t, overlapping, "%s: expected overlap", name)
		assert.InDeltaf(t, 1, nx, eps, "%s: normal.X (want +1, toward obstacle)", name)
		assert.InDeltaf(t, 0, ny, eps, "%s: normal.Y", name)
	}

	// Circle (origin) vs Circle (to the right).
	_, n, o := CircleCirclePenetration[float32](1, 0, 0, 1, 1.5, 0)
	assertRight(t, "CircleCircle", n.X, n.Y, o)

	// Circle (origin) vs AABB (to the right, spanning x in [0.5, 2.5]).
	_, n, o = CircleAABBPenetration[float32](1, 0, 0, 1, 1, 1.5, 0)
	assertRight(t, "CircleAABB", n.X, n.Y, o)

	// AABB (origin) vs AABB (to the right).
	_, n, o = AABBAABBPenetration[float32](0, 0, 1, 1, 1.5, 0, 1, 1)
	assertRight(t, "AABBAABB", n.X, n.Y, o)

	// Circle (origin) vs vertical segment at x = 0.5 (to the right).
	_, n, o = CircleLineSegmentPenetration[float32](1, 0, 0, 0.5, -1, 0.5, 1)
	assertRight(t, "CircleLineSegment", n.X, n.Y, o)

	// AABB (origin, spanning [-1,1]) vs segment from (0.5,0) inside to (3,0) right.
	_, n, o = AABBLineSegmentPenetration[float32](0, 0, 1, 1, 0.5, 0, 3, 0)
	assertRight(t, "AABBLineSegment", n.X, n.Y, o)
}
