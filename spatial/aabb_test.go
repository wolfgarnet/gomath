package spatial

import (
	"math/rand"
	"testing"

	"github.com/wolfgarnet/gomath"
	"github.com/wolfgarnet/gomath/intersection"
)

func makeAABB32(w, h float32, num int) (aabbs []AABB32) {
	for i := 0; i < num; i++ {
		aabbs = append(aabbs, AABB32{
			CenterX:    rand.Float32() * w,
			CenterY:    rand.Float32() * h,
			HalfWidth:  5,
			HalfHeight: 5,
		})
	}

	return aabbs
}

func makeAABB(w, h float32, num int) (aabbs []AABB[float32]) {
	for i := 0; i < num; i++ {
		aabbs = append(aabbs, AABB[float32]{
			CenterX:    rand.Float32() * w,
			CenterY:    rand.Float32() * h,
			HalfWidth:  5,
			HalfHeight: 5,
		})
	}

	return aabbs
}

func makeAABB2(w, h float32, num int) (aabbs []AABB2[float32]) {
	for i := 0; i < num; i++ {
		aabbs = append(aabbs, AABB2[float32]{
			Center:     [2]float32{rand.Float32() * w, rand.Float32() * h},
			HalfWidth:  5,
			HalfHeight: 5,
		})
	}

	return aabbs
}

func makeAABB3(w, h float32, num int) (aabbs []gomath.AABB[float32]) {
	for i := 0; i < num; i++ {
		aabbs = append(aabbs, gomath.AABB[float32]{
			Center:     gomath.Vector2[float32]{X: rand.Float32() * w, Y: rand.Float32() * h},
			HalfWidth:  5,
			HalfHeight: 5,
		})
	}

	return aabbs
}

func BenchmarkAABBAABB32(b *testing.B) {
	aabbs := makeAABB32(100, 100, 100000)
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(aabbs); i++ {
			for j := i + 1; j < len(aabbs); j++ {
				intersects := AABBAABB32(aabbs[i], aabbs[j])
				_ = intersects
			}
		}
	}
}

func BenchmarkAABBAABB(b *testing.B) {
	aabbs := makeAABB(100, 100, 100000)

	for n := 0; n < b.N; n++ {
		for i := 0; i < len(aabbs); i++ {
			for j := i + 1; j < len(aabbs); j++ {
				intersects := AABBAABB(aabbs[i], aabbs[j])
				_ = intersects
			}
		}
	}
}

func BenchmarkAABBAABB2(b *testing.B) {
	aabbs := makeAABB2(100, 100, 100000)

	for n := 0; n < b.N; n++ {
		for i := 0; i < len(aabbs); i++ {
			for j := i + 1; j < len(aabbs); j++ {
				intersects := AABBAABB2(aabbs[i], aabbs[j])
				_ = intersects
			}
		}
	}
}

func BenchmarkAABBAABB3(b *testing.B) {
	aabbs := makeAABB3(100, 100, 100000)

	for n := 0; n < b.N; n++ {
		for i := 0; i < len(aabbs); i++ {
			for j := i + 1; j < len(aabbs); j++ {
				intersects := AABBAABB3(aabbs[i], aabbs[j])
				_ = intersects
			}
		}
	}
}

func BenchmarkAABBAABB4(b *testing.B) {
	aabbs := makeAABB3(100, 100, 100000)

	for n := 0; n < b.N; n++ {
		for i := 0; i < len(aabbs); i++ {
			for j := i + 1; j < len(aabbs); j++ {
				intersects := intersection.AABBAABB(aabbs[i], aabbs[j])
				_ = intersects
			}
		}
	}
}
