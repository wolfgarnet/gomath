package spatial

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
	"math/rand"
	"testing"
)

type TestItem2D struct {
	bounds interface{}
}

func (ti TestItem2D) GetBounds() interface{} {
	return ti.bounds
}

func TestQuadTree_insert_one_circle(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	inserted := qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](0.0, 0, 0.5)})

	fmt.Printf("Length: %v - %v\n", len(qt.Items), inserted)

	if len(qt.Items) != 1 {
		t.Errorf("Not the correct number of nodes.")
	}
}

func TestQuadTree_insert_four_circles(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](1, 1, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2, 2, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](3, 3, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](4, 4, 0.5)})

	fmt.Printf("Length: %v\n", len(qt.Items))

	if len(qt.Items) != 4 {
		t.Errorf("Not the correct number of nodes.")
	}
}

func TestQuadTree_insert_points_circles(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{1, 1}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{2, 2}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{3, 3}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{4, 4}})

	fmt.Printf("Length: %v\n", len(qt.Items))

	if len(qt.Items) != 4 {
		t.Errorf("Not the correct number of nodes.")
	}
}

func TestQuadTree_insert_five(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{1, 1}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{2, 2}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{3, 3}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{4, 4}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{6, 6}})

	fmt.Printf("Length: %v\n", len(qt.Items))
	qt.Stringify()

	if len(qt.NorthEast.SouthWest.Items) != 4 {
		t.Errorf("Not the correct number of nodes.")
	}
	if len(qt.NorthEast.NorthEast.Items) != 1 {
		t.Errorf("Not the correct number of nodes.")
	}
	if len(qt.NorthEast.SouthEast.Items) != 0 {
		t.Errorf("Not the correct number of nodes.")
	}
}

func TestQuadTree_insert_five_circles(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](1, 1, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2, 2, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](3, 3, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](4, 4, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](6, 6, 0.5)})

	fmt.Printf("Length: %v\n", len(qt.Items))

	if len(qt.NorthEast.SouthWest.Items) != 4 {
		t.Errorf("Not the correct number of nodes.")
	}
	if len(qt.NorthEast.NorthEast.Items) != 1 {
		t.Errorf("Not the correct number of nodes.")
	}
	if len(qt.NorthEast.SouthEast.Items) != 0 {
		t.Errorf("Not the correct number of nodes.")
	}
}

func TestQuadTree_insert_five_aabb(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](1, 1, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](2, 2, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](3, 3, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](4, 4, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](6, 6, 0.5)})

	fmt.Printf("Length: %v\n", len(qt.Items))

	if len(qt.NorthEast.SouthWest.Items) != 4 {
		t.Errorf("Not the correct number of nodes.")
	}
	if len(qt.NorthEast.NorthEast.Items) != 1 {
		t.Errorf("Not the correct number of nodes.")
	}
	if len(qt.NorthEast.SouthEast.Items) != 0 {
		t.Errorf("Not the correct number of nodes.")
	}
}

func TestQuadTree_insert_four_aabb_crossing(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](1, 1, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](2, 2, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](3, 3, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](4, 4, 1.5)})

	fmt.Printf("Length: %v\n", len(qt.Items))

	if len(qt.Items) != 4 {
		t.Errorf("Not the correct number of nodes.")
	}
}

func TestQuadTree_insert_five_aabb_crossing(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](1, 1, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](2, 2, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](3, 3, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](4, 4, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](5, 5, 1.5)})

	assert.Len(t, qt.Items, 1)
	assert.Len(t, qt.NorthEast.Items, 4)
	assert.Len(t, qt.NorthWest.Items, 0)
	assert.Len(t, qt.SouthWest.Items, 0)
	assert.Len(t, qt.SouthEast.Items, 0)
}

func TestQuadTree_insert_five_circles_overlapping(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](1, 1, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2, 2, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](3, 3, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](4, 4, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](5, 5, 1.5)})

	assert.Len(t, qt.Items, 1)
	assert.Len(t, qt.NorthEast.Items, 4)
	assert.Len(t, qt.NorthWest.Items, 0)
	assert.Len(t, qt.SouthWest.Items, 0)
	assert.Len(t, qt.SouthEast.Items, 0)
}

func TestQuadTree_insert_five_circles_overlapping2(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2, 2, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.1, 2.1, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.2, 2.2, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.3, 2.3, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.4, 2.4, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.5, 2.5, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.1, 2.2, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.1, 2.3, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.1, 2.4, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.1, 2.5, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.2, 2.1, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.2, 2.2, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.2, 2.3, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.2, 2.4, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2.2, 2.5, 0.5)})

	//qt.Stringify()

	assert.Len(t, qt.Items, 0)
	assert.Len(t, qt.NorthEast.SouthWest.Items, 14)
	assert.Len(t, qt.NorthEast.SouthEast.Items, 0)
	assert.Len(t, qt.NorthEast.NorthEast.Items, 0)
	assert.Len(t, qt.NorthEast.NorthWest.Items, 0)
	assert.Len(t, qt.NorthEast.SouthWest.SouthWest.Items, 1)
	assert.Len(t, qt.SouthWest.Items, 0)
	assert.Len(t, qt.SouthEast.Items, 0)
}

func TestQuadTree_insert2(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{1, 1}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{2, 2}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{3, 3}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{4, 4}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{1, 1.2}})

	fmt.Printf("Length: %v\n", len(qt.Items))

	if len(qt.NorthEast.SouthWest.SouthWest.Items) != 3 {
		t.Errorf("Not the correct number of nodes. %v", len(qt.NorthEast.SouthWest.SouthWest.Items))
	}
	if len(qt.NorthEast.SouthWest.NorthEast.Items) != 2 {
		t.Errorf("Not the correct number of nodes. %v", len(qt.NorthEast.SouthWest.NorthEast.Items))
	}
	if len(qt.NorthEast.SouthWest.NorthWest.Items) != 0 {
		t.Errorf("Not the correct number of nodes.")
	}
}

func TestQuadTree_remove(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{1, 1}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{2, 2}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{3, 3}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{4, 4}})
	item := TestItem2D{bounds: gomath.Vector2[float32]{6, 6}}
	qt.Insert(item)

	fmt.Printf("Length: %v\n", len(qt.Items))

	if len(qt.NorthEast.SouthWest.Items) != 4 {
		t.Errorf("Not the correct number of nodes.")
	}
	if len(qt.NorthEast.NorthEast.Items) != 1 {
		t.Errorf("Not the correct number of nodes.")
	}
	if len(qt.NorthEast.SouthEast.Items) != 0 {
		t.Errorf("Not the correct number of nodes.")
	}

	qt.Remove(item)

	if len(qt.Items) != 4 {
		t.Errorf("Failed")
	}
	if qt.NorthEast != nil {
		t.Errorf("Should be nil")
	}
}

func TestQuadTree_remove_five_aabb(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](1, 1, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](2, 2, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](3, 3, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](4, 4, 0.5)})
	item := TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](6, 6, 0.5)}
	qt.Insert(item)

	qt.Remove(item)

	if len(qt.Items) != 4 {
		t.Errorf("Failed")
	}
	if qt.NorthEast != nil {
		t.Errorf("Should be nil")
	}
}

func TestQuadTree_remove_five_aabb_nocrossing(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	item1 := TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](1, 1, 0.5)}
	qt.Insert(item1)
	item2 := TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](2, 2, 0.5)}
	qt.Insert(item2)
	item3 := TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](3, 3, 0.5)}
	qt.Insert(item3)
	item4 := TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](4, 4, 0.5)}
	qt.Insert(item4)
	item5 := TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](-2, -2, 0.5)}
	qt.Insert(item5)

	qt.Remove(item5)

	qt.Stringify()

	if len(qt.Items) != 4 {
		t.Errorf("Failed, length of quad tree is %v", len(qt.Items))
	}
	if qt.NorthEast != nil {
		t.Errorf("Should be nil")
	}

	err := quadTreeHasItems(qt, item1)
	if err != nil {
		t.Errorf("Failed, %v", err)
	}
}

func TestQuadTree_remove_five_aabb_crossing(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	item1 := TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](1, 1, 0.5)}
	qt.Insert(item1)
	item2 := TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](2, 2, 0.5)}
	qt.Insert(item2)
	item3 := TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](3, 3, 0.5)}
	qt.Insert(item3)
	item4 := TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](4, 4, 0.5)}
	qt.Insert(item4)
	item6 := TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](5, 5, 0.5)}
	qt.Insert(item6)

	qt.Remove(item6)

	qt.Stringify()

	if len(qt.Items) != 4 {
		t.Errorf("Failed, length of quad tree is %v", len(qt.Items))
	}
	if qt.NorthEast != nil {
		t.Errorf("Should be nil")
	}

	err := quadTreeHasItems(qt, item1)
	if err != nil {
		t.Errorf("Failed, %v", err)
	}
}

func TestQuadTree_remove2(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{1, 1}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{2, 2}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{3, 3}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{4, 4}})
	item := TestItem2D{bounds: gomath.Vector2[float32]{6, 6}}
	qt.Insert(item)

	fmt.Printf("Length: %v\n", len(qt.Items))

	if len(qt.NorthEast.SouthWest.Items) != 4 {
		t.Errorf("Not the correct number of nodes. %v", len(qt.NorthEast.SouthWest.Items))
	}
	if len(qt.NorthEast.SouthEast.Items) != 0 {
		t.Errorf("Not the correct number of nodes.")
	}

	qt.Remove(item)

	if len(qt.Items) != 4 {
		t.Errorf("Failed")
	}
	if qt.NorthEast != nil {
		t.Errorf("Should be nil")
	}
}

func TestQuadTree_Query_simple(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	item1 := TestItem2D{bounds: gomath.Vector2[float32]{1, 1}}
	item2 := TestItem2D{bounds: gomath.Vector2[float32]{2, 2}}
	item3 := TestItem2D{bounds: gomath.Vector2[float32]{3, 3}}
	item4 := TestItem2D{bounds: gomath.Vector2[float32]{4, 4}}
	expected := []BoundedItem{item1, item2, item3, item4}
	qt.Insert(item1)
	qt.Insert(item2)
	qt.Insert(item3)
	qt.Insert(item4)

	ls := TestItem2D{bounds: gomath.NewLineSegmentXY[float32](0, 0, 5, 5)}
	list := qt.Query(ls)
	if len(list) != 4 {
		t.Errorf("Not the correct size, should have been 4")
		return
	}

	expected = checkItems(list, expected)

	if len(expected) > 0 {
		t.Errorf("All items should have been removed")
	}
}

func TestQuadTree_Query_simple2(t *testing.T) {
	qt := NewQuadTree[float32](0.0, 0, 10)
	item1 := TestItem2D{bounds: gomath.Vector2[float32]{1, 1}}
	item2 := TestItem2D{bounds: gomath.Vector2[float32]{2, 2}}
	item3 := TestItem2D{bounds: gomath.Vector2[float32]{3, 3}}
	item4 := TestItem2D{bounds: gomath.Vector2[float32]{4, 4}}
	item5 := TestItem2D{bounds: gomath.Vector2[float32]{5, 5}}
	item6 := TestItem2D{bounds: gomath.Vector2[float32]{6, 6}}
	item7 := TestItem2D{bounds: gomath.Vector2[float32]{7, 7}}
	item8 := TestItem2D{bounds: gomath.Vector2[float32]{8, 8}}
	item9 := TestItem2D{bounds: gomath.Vector2[float32]{9, 9}}
	qt.Insert(item1)
	qt.Insert(item2)
	qt.Insert(item3)
	qt.Insert(item4)
	qt.Insert(item5)
	qt.Insert(item6)
	qt.Insert(item7)
	qt.Insert(item8)
	qt.Insert(item9)

	expected := []BoundedItem{item1, item2, item3, item4, item5}
	ls := TestItem2D{bounds: gomath.NewLineSegmentXY[float32](1, 1, 4, 4)}
	list := qt.Query(ls)
	expected = checkItems(list, expected)
	if len(list) != 5 {
		t.Errorf("Size should have been 5, not %v", len(list))
	}
	if len(expected) > 0 {
		t.Errorf("All items should have been removed")
	}

	expected = []BoundedItem{item1, item2, item3, item4, item5, item6, item7, item8, item9}
	ls = TestItem2D{bounds: gomath.NewLineSegmentXY[float32](1, 1, 6, 6)}
	list = qt.Query(ls)
	expected = checkItems(list, expected)
	if len(list) != 9 {
		t.Errorf("Size should have been 5, not %v", len(list))
	}
	if len(expected) > 0 {
		t.Errorf("All items should have been removed")
	}

	expected = []BoundedItem{item6, item7, item8, item9}
	ls = TestItem2D{bounds: gomath.NewLineSegmentXY[float32](6, 6, 9, 9)}
	list = qt.Query(ls)
	expected = checkItems(list, expected)
	if len(list) != 4 {
		t.Errorf("Size should have been 4, not %v", len(list))
	}
	if len(expected) > 0 {
		t.Errorf("All items should have been removed")
	}
}

func TestDepth(t *testing.T) {
	number := 10000
	QTSize := float32(1)
	QTSizeDouble := QTSize * 2
	aabbs := make([]TestItem2D, number)
	for i := 0; i < number; i++ {
		aabbs[i] = TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](rand.Float32()*QTSizeDouble-QTSize, rand.Float32()*QTSizeDouble-QTSize, 0.1)}
	}

	qt := NewQuadTree[float32](0, 0, QTSize)
	for _, aabb := range aabbs {
		qt.Insert(aabb)
	}

	fmt.Printf("Depth: %v\n", qt.Depth())
}

// HELPERS

func checkItems(items, expected []BoundedItem) []BoundedItem {
	for _, i := range items {
		c := 0
		for _, i2 := range expected {
			if i == i2 {
				continue
			}

			expected[c] = i2
			c++
		}
		expected = expected[:c]
	}

	return expected
}

func quadTreeHasItems[T gomath.Float](qt *QuadTree[T], items ...BoundedItem) error {
	for _, i := range items {
		hasItem := false
		for _, p := range qt.Items {
			if p == i {
				hasItem = true
			}
		}

		if !hasItem {
			return fmt.Errorf("Quad tree does not have item, %v", i)
		}
	}

	return nil
}
