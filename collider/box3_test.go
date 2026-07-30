package collider

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wolfgarnet/gomath"
)

func axisAlignedUnit(center gomath.Vector3[float64]) gomath.Box3[float64] {
	return gomath.NewBox3(center, gomath.One3[float64]())
}

func TestBox3Box3(t *testing.T) {
	origin := axisAlignedUnit(gomath.Vector3[float64]{})

	cases := []struct {
		name    string
		other   gomath.Box3[float64]
		overlap bool
	}{
		{"coincident", origin, true},
		{"overlapping", axisAlignedUnit(gomath.Vector3[float64]{X: 1}), true},
		{"touching on a face", axisAlignedUnit(gomath.Vector3[float64]{X: 2}), true},
		{"touching at a corner", axisAlignedUnit(gomath.Vector3[float64]{X: 2, Y: 2, Z: 2}), true},
		{"separated on X", axisAlignedUnit(gomath.Vector3[float64]{X: 2.1}), false},
		{"separated on Y only", axisAlignedUnit(gomath.Vector3[float64]{X: 1, Y: 2.1}), false},
		{"separated on Z only", axisAlignedUnit(gomath.Vector3[float64]{Z: -2.1}), false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.overlap, Box3Box3(origin, testCase.other))
			// Overlap is symmetric.
			assert.Equal(t, testCase.overlap, Box3Box3(testCase.other, origin))
		})
	}
}

func TestBox3Box3PenetrationSeparated(t *testing.T) {
	penetration, _, overlapping := Box3Box3Penetration(
		axisAlignedUnit(gomath.Vector3[float64]{}),
		axisAlignedUnit(gomath.Vector3[float64]{X: 5}))

	assert.False(t, overlapping)
	assert.Zero(t, penetration)
}

func TestBox3Box3PenetrationPicksTheShallowestAxis(t *testing.T) {
	// Overlaps of 1.5 on X, 0.5 on Y and 1.0 on Z: the minimum translation is
	// along Y, which is the only correct answer.
	a := gomath.NewBox3(gomath.Vector3[float64]{}, gomath.Vector3[float64]{X: 1, Y: 1, Z: 1})
	b := gomath.NewBox3(gomath.Vector3[float64]{X: 0.5, Y: 1.5, Z: 1},
		gomath.Vector3[float64]{X: 1, Y: 1, Z: 1})

	penetration, normal, overlapping := Box3Box3Penetration(a, b)

	require.True(t, overlapping)
	assert.InDelta(t, 0.5, penetration, 1e-9)
	assert.Equal(t, gomath.Vector3[float64]{Y: 1}, normal)
}

func TestBox3Box3PenetrationNormalPointsFromFirstToSecond(t *testing.T) {
	origin := axisAlignedUnit(gomath.Vector3[float64]{})

	offsets := []gomath.Vector3[float64]{
		{X: 1.5}, {X: -1.5}, {Y: 1.5}, {Y: -1.5}, {Z: 1.5}, {Z: -1.5},
	}

	for _, offset := range offsets {
		_, normal, overlapping := Box3Box3Penetration(origin, axisAlignedUnit(offset))
		require.True(t, overlapping, "offset %v should overlap", offset)

		assert.Greater(t, normal.Dot(offset), 0.0,
			"the normal must point from a towards b, got %v for offset %v", normal, offset)
	}
}

// TestBox3Box3PenetrationResolvesTheOverlap is the property that makes the
// result useful.
func TestBox3Box3PenetrationResolvesTheOverlap(t *testing.T) {
	rng := rand.New(rand.NewSource(51))

	resolved := 0

	for i := 0; i < 3000; i++ {
		makeBox := func() gomath.Box3[float64] {
			return gomath.NewBox3(
				gomath.Vector3[float64]{
					X: (rng.Float64()*2 - 1) * 2,
					Y: (rng.Float64()*2 - 1) * 2,
					Z: (rng.Float64()*2 - 1) * 2,
				},
				gomath.Vector3[float64]{
					X: rng.Float64() + 0.3,
					Y: rng.Float64() + 0.3,
					Z: rng.Float64() + 0.3,
				})
		}

		a, b := makeBox(), makeBox()

		penetration, normal, overlapping := Box3Box3Penetration(a, b)
		require.Equal(t, Box3Box3(a, b), overlapping,
			"the boolean and penetration tests must agree")

		if !overlapping {
			continue
		}
		resolved++

		pushed := b.Translate(normal.MultiplyScalar(penetration + 1e-9))
		assert.False(t, Box3Box3(a, pushed),
			"translating by the reported MTV should separate the boxes")
	}

	assert.Greater(t, resolved, 300, "the sample should contain plenty of overlaps")
}
