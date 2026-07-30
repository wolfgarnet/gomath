package collider

import (
	"math"
	"math/rand"
	"testing"

	"github.com/wolfgarnet/gomath"
)

// Benchmarks comparing the optimised scalar formulation against the direct
// vector-cross-product transcription, which is what the naive implementation
// does. Run pinned, since this machine's hybrid cores make unpinned Go
// benchmarks meaningless:
//
//	taskset -c 0 env GOMAXPROCS=1 go test -bench=OBB3 -benchtime=500ms -count=8 ./collider/

func benchmarkPairs(count int) [][2]gomath.OBB3[float64] {
	rng := rand.New(rand.NewSource(71))

	pairs := make([][2]gomath.OBB3[float64], 0, count)
	for i := 0; i < count; i++ {
		makeBox := func() gomath.OBB3[float64] {
			axis := gomath.Vector3[float64]{
				X: rng.Float64()*2 - 1,
				Y: rng.Float64()*2 - 1,
				Z: rng.Float64()*2 - 1,
			}.Normalize()
			if axis.IsZero() {
				axis = gomath.Vector3[float64]{Y: 1}
			}

			return gomath.NewOBB3FromQuaternion(
				gomath.Vector3[float64]{
					X: (rng.Float64()*2 - 1) * 3,
					Y: (rng.Float64()*2 - 1) * 3,
					Z: (rng.Float64()*2 - 1) * 3,
				},
				gomath.Vector3[float64]{
					X: rng.Float64() + 0.3,
					Y: rng.Float64() + 0.3,
					Z: rng.Float64() + 0.3,
				},
				gomath.QuaternionFromAxisAngle(axis, (rng.Float64()*2-1)*math.Pi))
		}

		pairs = append(pairs, [2]gomath.OBB3[float64]{makeBox(), makeBox()})
	}

	return pairs
}

var (
	benchPairs = benchmarkPairs(512)
	benchSink  bool
)

func BenchmarkOBB3Optimised(b *testing.B) {
	for b.Loop() {
		result := false
		for _, pair := range benchPairs {
			result = result != OBB3OBB3(pair[0], pair[1])
		}
		benchSink = result
	}
}

func BenchmarkOBB3Naive(b *testing.B) {
	for b.Loop() {
		result := false
		for _, pair := range benchPairs {
			result = result != naiveOBB3OBB3(pair[0], pair[1])
		}
		benchSink = result
	}
}
