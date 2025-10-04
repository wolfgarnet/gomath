package spatial

import (
	"math/rand"
	"testing"

	"github.com/wolfgarnet/gomath"
)

type BoxedTestItem32 struct {
	IDValue uint64
	gomath.AABB[float32]
}

func (i BoxedTestItem32) ID() uint64 {
	return i.IDValue
}

func (i BoxedTestItem32) BoundingBox() gomath.AABB[float32] {
	return i.AABB
}

func (i BoxedTestItem32) GetBounds() any {
	return i.AABB
}

func makeBoxedQuadtreeItems(w, h float32, num int) (items []BoxedTestItem32) {
	for i := 0; i < num; i++ {
		items = append(items, BoxedTestItem32{
			IDValue: uint64(i),
			AABB: gomath.AABB[float32]{
				Center:     gomath.Vector2[float32]{X: rand.Float32()*w - w/2, Y: rand.Float32()*h - h/2},
				HalfWidth:  5,
				HalfHeight: 5,
			},
		})
	}
	return items
}

func BenchmarkQuadtreeInsert(b *testing.B) {
	tree := NewQuadtree[float32, BoxedTestItem32, uint64](0, 0, 100)
	items := makeBoxedQuadtreeItems(200, 200, 100000)

	for i := 0; i < b.N; i++ {
		for _, item := range items {
			tree.Insert(item)
		}
		tree.Clear()
	}
}

func BenchmarkBoxedQuadtreeInsert(b *testing.B) {
	tree := NewBoxedQuadtree[float32, BoxedTestItem32](0, 0, 100)
	items := makeBoxedQuadtreeItems(200, 200, 100_000)

	for i := 0; i < b.N; i++ {
		for _, item := range items {
			tree.Insert(item)
		}
		tree.Clear()
	}
}

func BenchmarkBoxedQuadtreeInsert2(b *testing.B) {
	tree := newBoxedQuadtree[float32, BoxedTestItem32](0, 0, 100, nil, 0, 8, 8)
	items := makeBoxedQuadtreeItems(200, 200, 100_000)

	for i := 0; i < b.N; i++ {
		for _, item := range items {
			tree.Insert(item)
		}
		tree.Clear()
	}
}
