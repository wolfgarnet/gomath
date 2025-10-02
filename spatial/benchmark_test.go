package spatial

import (
	"math/rand"
	"testing"

	"github.com/wolfgarnet/gomath"
)

type testItem struct {
	IDValue       uint64
	PositionValue gomath.Vector2[float32]
	radius        float32
	circle        gomath.Circle[float32]
}

func (i *testItem) Position() gomath.Vector2[float32] {
	return i.PositionValue
}

func (i *testItem) ID() uint64 {
	return i.IDValue
}

func (i *testItem) GetBounds() any {
	return i.circle
}

func generateRandomTestItems(count int, halfBounds float32) []*testItem {
	items := make([]*testItem, count)
	for i := 0; i < count; i++ {
		item := &testItem{
			IDValue:       uint64(i + 1),
			PositionValue: gomath.Vector2[float32]{X: rand.Float32()*halfBounds*2 - halfBounds, Y: rand.Float32()*halfBounds*2 - halfBounds},
			radius:        rand.Float32()*2 + 1,
		}
		item.circle = gomath.Circle[float32]{Center: item.PositionValue, Radius: item.radius}

		items[i] = item
	}
	return items
}

func BenchmarkQuadtreeAddItem(b *testing.B) {
	tree := NewQuadtree[float32, *testItem, uint64](0, 0, 250)
	items := generateRandomTestItems(100000, 250)

	for n := 0; n < b.N; n++ {
		for _, item := range items {
			tree.Insert(item)
		}
	}
}

func BenchmarkHashGridAddItem(b *testing.B) {
	grid := NewHashGrid[float32, *testItem, uint64](64, 64, 10.0)
	items := generateRandomTestItems(100000, 320)

	for n := 0; n < b.N; n++ {
		for _, item := range items {
			grid.AddItem(item)
		}
	}
}
