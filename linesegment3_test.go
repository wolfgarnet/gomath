package gomath

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLineSegment3ClosestPoint(t *testing.T) {
	segment := NewLineSegment3(Vector3[float64]{X: 1}, Vector3[float64]{X: 4})

	tests := []struct {
		name  string
		point Vector3[float64]
		want  Vector3[float64]
	}{
		{"perpendicular onto the middle", Vector3[float64]{X: 2, Y: 5}, Vector3[float64]{X: 2}},
		{"before the start clamps to P1", Vector3[float64]{X: -10, Z: 3}, Vector3[float64]{X: 1}},
		{"past the end clamps to P2", Vector3[float64]{X: 99, Y: -2}, Vector3[float64]{X: 4}},
		{"a point on the segment is its own closest", Vector3[float64]{X: 2.5}, Vector3[float64]{X: 2.5}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, segment.ClosestPoint(test.point))
		})
	}
}

// A degenerate segment is a point, and every routine that divides by its length
// has to survive it — this is the case that turns a Capsule3 into a Sphere, so it
// is not a corner of the input space but the middle of it.
func TestLineSegment3Degenerate(t *testing.T) {
	point := Vector3[float64]{X: 2, Y: 3, Z: 4}
	segment := NewLineSegment3(point, point)

	require.True(t, segment.IsDegenerate())

	assert.Equal(t, float64(0), segment.Length())
	assert.Equal(t, point, segment.ClosestPoint(Vector3[float64]{X: 100}))
	assert.Equal(t, Vector3[float64]{}, segment.UnitDirection())
	assert.InDelta(t, 98, segment.DistanceToPoint(Vector3[float64]{X: 100, Y: 3, Z: 4}), 1e-12)

	other := NewLineSegment3(Vector3[float64]{Y: 3, Z: 4}, Vector3[float64]{X: 10, Y: 3, Z: 4})
	s, other2 := segment.ClosestParameters(other)
	assert.Equal(t, float64(0), s, "a point has only the parameter 0")
	assert.InDelta(t, 0.2, other2, 1e-12, "and projects onto the other segment")

	assert.InDelta(t, 0, segment.DistanceToSegment(other), 1e-12)
}

// The parallel case has no unique closest pair, and the choice made is part of the
// contract: the middle of the overlapping interval. Two upright characters
// standing side by side are exactly this, so a contact that jumps to their feet
// would be a visible bug rather than a numerical curiosity.
func TestLineSegment3ParallelClosestPairIsTheOverlapMidpoint(t *testing.T) {
	tests := []struct {
		name    string
		first   LineSegment3[float64]
		second  LineSegment3[float64]
		wantS   float64
		wantT   float64
		wantGap float64
	}{
		{
			name:    "level and side by side",
			first:   NewLineSegment3(Vector3[float64]{}, Vector3[float64]{Y: 2}),
			second:  NewLineSegment3(Vector3[float64]{X: 1}, Vector3[float64]{X: 1, Y: 2}),
			wantS:   0.5,
			wantT:   0.5,
			wantGap: 1,
		},
		{
			name:    "one offset upwards, overlapping over its lower half",
			first:   NewLineSegment3(Vector3[float64]{}, Vector3[float64]{Y: 2}),
			second:  NewLineSegment3(Vector3[float64]{X: 1, Y: 1}, Vector3[float64]{X: 1, Y: 3}),
			wantS:   0.75,
			wantT:   0.25,
			wantGap: 1,
		},
		{
			name:    "the other pointing the opposite way",
			first:   NewLineSegment3(Vector3[float64]{}, Vector3[float64]{Y: 2}),
			second:  NewLineSegment3(Vector3[float64]{X: 1, Y: 2}, Vector3[float64]{X: 1}),
			wantS:   0.5,
			wantT:   0.5,
			wantGap: 1,
		},
		{
			name:    "end to end, no overlap in projection",
			first:   NewLineSegment3(Vector3[float64]{}, Vector3[float64]{Y: 2}),
			second:  NewLineSegment3(Vector3[float64]{X: 1, Y: 5}, Vector3[float64]{X: 1, Y: 7}),
			wantS:   1,
			wantT:   0,
			wantGap: math.Sqrt(1 + 9),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, u := test.first.ClosestParameters(test.second)

			assert.InDelta(t, test.wantS, s, 1e-12, "parameter on the first segment")
			assert.InDelta(t, test.wantT, u, 1e-12, "parameter on the second segment")
			assert.InDelta(t, test.wantGap, test.first.DistanceToSegment(test.second), 1e-12)
		})
	}
}

// Very nearly parallel must land in the same place as exactly parallel. The
// textbook denominator is fine at a degree of misalignment and useless at a
// microradian, and the whole point of the relative threshold is that nothing
// jumps as a pair rotates into alignment.
func TestLineSegment3NearlyParallelIsStable(t *testing.T) {
	exact := NewLineSegment3(Vector3[float64]{X: 1}, Vector3[float64]{X: 1, Y: 2})

	for _, tilt := range []float64{1e-9, 1e-8, 1e-7} {
		first := NewLineSegment3(Vector3[float64]{}, Vector3[float64]{Y: 2})
		second := NewLineSegment3(exact.P1, exact.P2.Add(Vector3[float64]{Z: tilt}))

		s, u := first.ClosestParameters(second)

		assert.InDeltaf(t, 0.5, s, 1e-3, "tilt %g: first parameter", tilt)
		assert.InDeltaf(t, 0.5, u, 1e-3, "tilt %g: second parameter", tilt)
	}
}

// Crossing segments meet, so the distance is zero and the closest points coincide.
func TestLineSegment3CrossingSegmentsTouch(t *testing.T) {
	first := NewLineSegment3(Vector3[float64]{X: -1}, Vector3[float64]{X: 1})
	second := NewLineSegment3(Vector3[float64]{Y: -1}, Vector3[float64]{Y: 1})

	onFirst, onSecond := first.ClosestPoints(second)

	assert.InDelta(t, 0, onFirst.DistanceTo(onSecond), 1e-12)
	assert.InDelta(t, 0, first.DistanceToSegment(second), 1e-12)
}

func TestLineSegment3Helpers(t *testing.T) {
	segment := NewLineSegment3(Vector3[float64]{X: 1, Y: 2, Z: 3}, Vector3[float64]{X: 1, Y: 2, Z: 7})

	assert.Equal(t, Vector3[float64]{X: 1, Y: 2, Z: 5}, segment.Midpoint())
	assert.Equal(t, Vector3[float64]{X: 1, Y: 2, Z: 5}, segment.PointAt(0.5))
	assert.Equal(t, float64(4), segment.Length())
	assert.Equal(t, float64(16), segment.LengthSquared())
	assert.Equal(t, Vector3[float64]{Z: 1}, segment.UnitDirection())
	assert.Equal(t, segment.P2, segment.Reversed().P1)
	assert.Equal(t, NewBox3FromMinMax(segment.P1, segment.P2), segment.BoundingBox())
	assert.True(t, segment.Equals(NewLineSegment3(segment.P1, segment.P2)))
	assert.False(t, segment.Equals(segment.Reversed()))

	moved := segment.Translate(Vector3[float64]{X: 10})
	assert.Equal(t, Vector3[float64]{X: 11, Y: 2, Z: 3}, moved.P1)
}

// The element-wise helpers are exact for an integer T; nothing here divides or
// takes a root.
func TestLineSegment3IntegerStorage(t *testing.T) {
	segment := NewLineSegment3(Vector3[int]{X: 2, Y: 4}, Vector3[int]{X: 6, Y: 8})

	assert.Equal(t, Vector3[int]{X: 4, Y: 6}, segment.Midpoint())
	assert.Equal(t, 32, segment.LengthSquared())
	assert.Equal(t, NewBox3FromMinMax(Vector3[int]{X: 2, Y: 4}, Vector3[int]{X: 6, Y: 8}), segment.BoundingBox())
}

// A brute-force oracle for segment-to-segment distance: sample both parameters on
// a grid and take the smallest separation found. Sampling can only ever OVERSTATE
// the minimum, so the analytic answer must sit at or below it — and it must also
// be achieved by a real pair of points on the two segments, which is what the
// closest-points check covers. Together those trap the answer from both sides.
func TestLineSegment3DistanceAgainstBruteForceOracle(t *testing.T) {
	const (
		trials  = 300
		samples = 400
	)

	random := rand.New(rand.NewSource(20260824))

	randomSegment := func() LineSegment3[float64] {
		point := func() Vector3[float64] {
			return Vector3[float64]{
				X: random.Float64()*4 - 2,
				Y: random.Float64()*4 - 2,
				Z: random.Float64()*4 - 2,
			}
		}

		return NewLineSegment3(point(), point())
	}

	for trial := range trials {
		first, second := randomSegment(), randomSegment()

		sampled := math.Inf(1)
		for i := 0; i <= samples; i++ {
			onFirst := first.PointAt(float64(i) / samples)
			for j := 0; j <= samples/10; j++ {
				sampled = math.Min(sampled, onFirst.DistanceTo(second.PointAt(float64(j)/float64(samples/10))))
			}
		}

		analytic := first.DistanceToSegment(second)

		assert.LessOrEqualf(t, analytic, sampled+1e-12,
			"trial %d: analytic distance exceeds the sampled minimum", trial)

		// And the answer must be realised by points that genuinely lie on the
		// segments, which rules out a distance that is too small.
		onFirst, onSecond := first.ClosestPoints(second)
		assert.InDeltaf(t, analytic, onFirst.DistanceTo(onSecond), 1e-12,
			"trial %d: distance must be the gap between the returned points", trial)
		assert.InDeltaf(t, 0, onFirst.DistanceTo(first.ClosestPoint(onFirst)), 1e-12,
			"trial %d: the returned point must lie on the first segment", trial)
		assert.InDeltaf(t, 0, onSecond.DistanceTo(second.ClosestPoint(onSecond)), 1e-12,
			"trial %d: the returned point must lie on the second segment", trial)
	}
}
