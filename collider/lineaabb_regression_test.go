package collider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
)

// LineAABB clips the segment against each axis slab, and skips an axis when the
// direction is "parallel" to it. The parallel guard used gomath.IsEpsilon, whose
// threshold is 0.9 — so any direction component smaller than 0.9 counted as
// parallel.
//
// Segment directions here are NOT normalised: they are P2-P1, so a component of
// 0.5 is entirely ordinary. Skipping that axis means it never constrains the
// intersection, and the test reports a hit where there is none.
//
// segmentAABBOracle below is an independent analytic solution used to confirm it.

// segmentAABBOracle solves the slab clipping directly, with no epsilon anywhere.
// Only an exactly-zero direction component is special-cased, which is the sole
// case that would divide zero by zero.
func segmentAABBOracle(segment gomath.LineSegment[float64], box gomath.AABB[float64]) bool {
	direction := segment.P2.Sub(segment.P1)

	minimum := gomath.Vector2[float64]{
		X: box.Center.X - box.HalfWidth,
		Y: box.Center.Y - box.HalfHeight,
	}
	maximum := gomath.Vector2[float64]{
		X: box.Center.X + box.HalfWidth,
		Y: box.Center.Y + box.HalfHeight,
	}

	enter, exit := 0.0, 1.0

	for axis := 0; axis < 2; axis++ {
		step := direction.Component(axis)
		origin := segment.P1.Component(axis)
		low := minimum.Component(axis)
		high := maximum.Component(axis)

		if step == 0 {
			// Genuinely parallel: the segment either lies within this slab for
			// its whole length, or misses entirely.
			if origin < low || origin > high {
				return false
			}

			continue
		}

		first := (low - origin) / step
		second := (high - origin) / step
		if first > second {
			first, second = second, first
		}

		if first > enter {
			enter = first
		}
		if second < exit {
			exit = second
		}
		if enter > exit {
			return false
		}
	}

	return true
}

func TestLineAABBParallelGuardDoesNotSwallowRealAxes(t *testing.T) {
	// The segment climbs steeply: dx is 0.5, dy is 100. Under a 0.9 threshold dx
	// counts as "parallel to X", so the X slab is never clipped against.
	segment := gomath.LineSegment[float64]{
		P1: gomath.Vector2[float64]{X: 0, Y: 0},
		P2: gomath.Vector2[float64]{X: 0.5, Y: 100},
	}

	// A tall, narrow box the segment passes beside rather than through: where the
	// segment is at the box's height, its X is 0.2 to 0.3, outside x in
	// [-0.1, 0.1].
	box := gomath.AABB[float64]{
		Center:     gomath.Vector2[float64]{X: 0, Y: 50},
		HalfWidth:  0.1,
		HalfHeight: 10,
	}

	oracle := segmentAABBOracle(segment, box)
	assert.False(t, oracle, "setup: the oracle must agree these do not intersect")

	assert.Equal(t, oracle, LineAABB(segment, box),
		"LineAABB should agree with the analytic solution")
}

// TestLineAABBAgreesWithOracle is the general property, over a range of slopes
// that straddle the old 0.9 threshold.
func TestLineAABBAgreesWithOracle(t *testing.T) {
	box := gomath.AABB[float64]{
		Center:     gomath.Vector2[float64]{X: 0, Y: 50},
		HalfWidth:  0.1,
		HalfHeight: 10,
	}

	// Direction components deliberately spanning below, at and above 0.9.
	for _, deltaX := range []float64{0.05, 0.2, 0.5, 0.89, 0.9, 1.0, 2.0, 5.0} {
		for _, offsetX := range []float64{-0.5, -0.05, 0, 0.05, 0.5} {
			segment := gomath.LineSegment[float64]{
				P1: gomath.Vector2[float64]{X: offsetX, Y: 0},
				P2: gomath.Vector2[float64]{X: offsetX + deltaX, Y: 100},
			}

			assert.Equal(t, segmentAABBOracle(segment, box), LineAABB(segment, box),
				"deltaX %v offsetX %v", deltaX, offsetX)
		}
	}
}

// TestLineAABBTrulyParallelSegments confirms the exactly-parallel case still
// works, since that is what the guard is legitimately for.
func TestLineAABBTrulyParallelSegments(t *testing.T) {
	box := gomath.AABB[float64]{
		Center:     gomath.Vector2[float64]{X: 0, Y: 0},
		HalfWidth:  1,
		HalfHeight: 1,
	}

	cases := []struct {
		name    string
		segment gomath.LineSegment[float64]
		want    bool
	}{
		{
			name: "horizontal through the box",
			segment: gomath.LineSegment[float64]{
				P1: gomath.Vector2[float64]{X: -5, Y: 0},
				P2: gomath.Vector2[float64]{X: 5, Y: 0},
			},
			want: true,
		},
		{
			name: "horizontal above the box",
			segment: gomath.LineSegment[float64]{
				P1: gomath.Vector2[float64]{X: -5, Y: 2},
				P2: gomath.Vector2[float64]{X: 5, Y: 2},
			},
			want: false,
		},
		{
			name: "vertical through the box",
			segment: gomath.LineSegment[float64]{
				P1: gomath.Vector2[float64]{X: 0, Y: -5},
				P2: gomath.Vector2[float64]{X: 0, Y: 5},
			},
			want: true,
		},
		{
			name: "vertical beside the box",
			segment: gomath.LineSegment[float64]{
				P1: gomath.Vector2[float64]{X: 3, Y: -5},
				P2: gomath.Vector2[float64]{X: 3, Y: 5},
			},
			want: false,
		},
		{
			name: "horizontal but too short to reach",
			segment: gomath.LineSegment[float64]{
				P1: gomath.Vector2[float64]{X: -5, Y: 0},
				P2: gomath.Vector2[float64]{X: -3, Y: 0},
			},
			want: false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.want, segmentAABBOracle(testCase.segment, box),
				"the oracle should agree with the expected answer")
			assert.Equal(t, testCase.want, LineAABB(testCase.segment, box))
		})
	}
}
