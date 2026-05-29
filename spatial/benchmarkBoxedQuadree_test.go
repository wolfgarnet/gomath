package spatial

import (
	"math/rand"
	"testing"

	"github.com/wolfgarnet/gomath"
)

func (i BoxedItem[T, ID]) ID() ID {
	return i.ItemID
}

func (i BoxedItem[T, ID]) BoundingBox() gomath.AABB[T] {
	return i.AABB
}

func (i BoxedItem[T, ID]) GetBounds() any {
	return i.AABB
}

func makeBoxedQuadtreeItems(w, h float32, num int) (items []BoxedItem[float32, uint64]) {
	for i := 0; i < num; i++ {
		items = append(items, BoxedItem[float32, uint64]{
			ItemID: uint64(i),
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
	tree := NewQuadtree[float32, BoxedItem[float32, uint64], uint64](0, 0, 100)
	items := makeBoxedQuadtreeItems(200, 200, 100_000)

	for i := 0; i < b.N; i++ {
		for _, item := range items {
			tree.Insert(item)
		}
		tree.Clear()
	}
}

func BenchmarkBoxedQuadtreeInsert(b *testing.B) {
	tree := NewBoxedQuadtree[float32, uint64](0, 0, 100)
	items := makeBoxedQuadtreeItems(200, 200, 100_000)

	for i := 0; i < b.N; i++ {
		for _, item := range items {
			_, _ = tree.Insert(item)
		}
		tree.Clear()
	}
}

func BenchmarkBoxedQuadtreeInsert2(b *testing.B) {
	tree := newBoxedQuadtree[float32, uint64](0, 0, 100, nil, 0, 8, 8)
	items := makeBoxedQuadtreeItems(200, 200, 100_000)

	for i := 0; i < b.N; i++ {
		for _, item := range items {
			_, _ = tree.Insert(item)
		}
		tree.Clear()
	}
}

/*
111289190
107287845

10.316.375
10.777.693

*/
