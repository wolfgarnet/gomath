package spatial

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wolfgarnet/gomath"
)

type TestItem2D struct {
	bounds interface{}
	id     uint64
}

func (ti TestItem2D) GetBounds() interface{} {
	return ti.bounds
}

func (ti TestItem2D) ID() uint64 {
	return ti.id
}

func TestQuadtree_insert_one_circle(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
	inserted := qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](0.0, 0, 0.5)})
	assert.True(t, inserted)
	assert.Len(t, qt.Items, 1, "Should be one")
}

func TestQuadtree_insert_one_circle_outside(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
	inserted := qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](15.0, 15.0, 0.5)})
	assert.False(t, inserted)
	assert.Len(t, qt.Items, 0, "Should be zero")
}

func TestQuadtree_insert_one_circle_outside_allow_expand(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
	qt.mayExpand = true
	inserted := qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](15.0, 15.0, 0.5)})
	assert.True(t, inserted)
	assert.Len(t, qt.Items, 1, "Should be one")
	assert.Equal(t, float32(20.0), qt.Boundary.HalfSize)
}

func TestQuadtree_insert_four_circles(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](1, 1, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2, 2, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](3, 3, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](4, 4, 0.5)})

	assert.Len(t, qt.Items, 4, "Not the correct number of nodes")
}

func TestQuadtree_insert_four_circles_outside(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
	qt.mayExpand = true
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](15, 1, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2, 12, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](-3, 3, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](-25, 4, 0.5)})

	assert.Len(t, qt.Items, 4, "Not the correct number of nodes.")
	assert.Equal(t, float32(40.0), qt.Boundary.HalfSize)
}

func TestQuadtree_insert_points_circles(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 1, Y: 1}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 2, Y: 2}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 3, Y: 3}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 4, Y: 4}})

	assert.Len(t, qt.Items, 4, "Not the correct number of nodes.")
}

func TestQuadtree_insert_five(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D, uint64](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 1, Y: 1}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 2, Y: 2}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 3, Y: 3}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 4, Y: 4}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 6, Y: 6}})

	assert.Len(t, qt.Items, 0, "Should be empty")
	fmt.Printf("Length: %v\n", len(qt.Items))
	qt.Stringify()

	assertItems(t, qt.NorthEast, 1, 0, 0, 4)
}

func TestQuadtree_insert_five_outside(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
	qt.mayExpand = true
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 1, Y: 1}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 2, Y: 2}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 3, Y: 3}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 4, Y: 4}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 15, Y: 6}})
	assert.Equal(t, float32(20.0), qt.Boundary.HalfSize)

	assert.Len(t, qt.Items, 0, "Should be empty")
	fmt.Printf("Length: %v\n", len(qt.Items))
	qt.Stringify()

	assertItems(t, qt.NorthEast, 0, 0, 1, 4)
}

func TestQuadtree_insert_five_inside_one_outside(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
	qt.mayExpand = true
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 1, Y: 1}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 2, Y: 2}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 3, Y: 3}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 4, Y: 4}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 6, Y: 6}})
	qt.Insert(TestItem2D{bounds: gomath.Vector2[float32]{X: 15, Y: 6}})
	assert.Equal(t, float32(20.0), qt.Boundary.HalfSize)

	assert.Len(t, qt.Items, 0, "Should be empty")
	fmt.Printf("Length: %v\n", len(qt.Items))
	qt.Stringify()

	assertItems(t, qt.NorthEast, 0, 0, 1, 0)
	assertItems(t, qt.NorthWest, 0, 0, 0, 0)
	assertItems(t, qt.SouthWest, 0, 0, 0, 0)
	assertItems(t, qt.SouthEast, 0, 0, 0, 0)
	assertItems(t, qt.NorthEast.SouthWest, 1, 0, 0, 4)
}

func TestQuadtree_insert_five_circles(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](1, 1, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](2, 2, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](3, 3, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](4, 4, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewCircle[float32](6, 6, 0.5)})

	assert.Len(t, qt.Items, 0, "Should be empty")
	fmt.Printf("Length: %v\n", len(qt.Items))
	qt.Stringify()

	assert.Len(t, qt.NorthEast.SouthWest.Items, 4, "Not the correct number of nodes.")
	assert.Len(t, qt.NorthEast.NorthEast.Items, 1, "Not the correct number of nodes.")
	assert.Len(t, qt.NorthEast.SouthEast.Items, 0, "Not the correct number of nodes.")
	assert.Len(t, qt.NorthEast.NorthWest.Items, 0, "Not the correct number of nodes.")
}

func TestQuadtree_insert_five_aabb(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](1, 1, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](2, 2, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](3, 3, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](4, 4, 0.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](6, 6, 0.5)})

	assert.Len(t, qt.Items, 0, "Should be empty")
	fmt.Printf("Length: %v\n", len(qt.Items))
	qt.Stringify()

	assert.Len(t, qt.NorthEast.SouthWest.Items, 4, "Not the correct number of nodes")
	assert.Len(t, qt.NorthEast.NorthEast.Items, 1, "Not the correct number of nodes")
	assert.Len(t, qt.NorthEast.SouthEast.Items, 0, "Not the correct number of nodes")
	assert.Len(t, qt.NorthEast.NorthWest.Items, 0, "Not the correct number of nodes")
}

func TestQuadtree_insert_four_aabb_crossing(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](1, 1, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](2, 2, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](3, 3, 1.5)})
	qt.Insert(TestItem2D{bounds: gomath.NewSquaredAABB2DXY[float32](4, 4, 1.5)})

	assert.Len(t, qt.Items, 4, "Not the correct number of nodes")
}

func TestQuadtree_insert_five_aabb_crossing(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
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

func TestQuadtree_insert_five_circles_overlapping(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
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

func TestQuadtree_insert_five_circles_overlapping2(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
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

func TestQuadtree_insert2(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
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

func TestQuadtree_remove(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
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

func TestQuadtree_remove_five_aabb(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
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

func TestQuadtree_remove_five_aabb_nocrossing(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
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

func TestQuadtree_remove_five_aabb_crossing(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
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

func TestQuadtree_remove2(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D](0.0, 0, 10)
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

func TestQuadtree_Query_simple(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D, uint64](0.0, 0, 10)
	item1 := TestItem2D{bounds: gomath.Vector2[float32]{1, 1}}
	item2 := TestItem2D{bounds: gomath.Vector2[float32]{2, 2}}
	item3 := TestItem2D{bounds: gomath.Vector2[float32]{3, 3}}
	item4 := TestItem2D{bounds: gomath.Vector2[float32]{4, 4}}
	expected := []TestItem2D{item1, item2, item3, item4}
	qt.Insert(item1)
	qt.Insert(item2)
	qt.Insert(item3)
	qt.Insert(item4)

	ls := TestItem2D{bounds: gomath.NewLineSegmentXY[float32](0, 0, 5, 5)}
	list := QueryQuadtreeWithResult(qt, ls)
	if len(list) != 4 {
		t.Errorf("Not the correct size, should have been 4")
		return
	}

	expected = checkItems(list, expected)

	if len(expected) > 0 {
		t.Errorf("All items should have been removed")
	}
}

func TestQuadtree_Query_simple2(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D, uint64](0.0, 0, 10)
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

	expected := []TestItem2D{item1, item2, item3, item4, item5}
	ls := TestItem2D{bounds: gomath.NewLineSegmentXY[float32](1, 1, 4, 4)}
	list := QueryQuadtreeWithBoundedItemWithoutCheckWithResult(qt, ls)
	expected = checkItems(list, expected)
	if len(list) != 5 {
		t.Errorf("Size should have been 5, not %v", len(list))
	}
	if len(expected) > 0 {
		t.Errorf("All items should have been removed")
	}

	expected = []TestItem2D{item1, item2, item3, item4, item5, item6, item7, item8, item9}
	ls = TestItem2D{bounds: gomath.NewLineSegmentXY[float32](1, 1, 6, 6)}
	list = QueryQuadtreeWithBoundedItemWithoutCheckWithResult(qt, ls)
	expected = checkItems(list, expected)
	if len(list) != 9 {
		t.Errorf("Size should have been 5, not %v", len(list))
	}
	if len(expected) > 0 {
		t.Errorf("All items should have been removed")
	}

	expected = []TestItem2D{item6, item7, item8, item9}
	ls = TestItem2D{bounds: gomath.NewLineSegmentXY[float32](6, 6, 9, 9)}
	list = QueryQuadtreeWithBoundedItemWithoutCheckWithResult(qt, ls)
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

	qt := NewQuadtree[float32, TestItem2D](0, 0, QTSize)
	for _, aabb := range aabbs {
		qt.Insert(aabb)
	}

	fmt.Printf("Depth: %v\n", qt.Depth())
}

// HELPERS

func checkItems(items, expected []TestItem2D) []TestItem2D {
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

func quadTreeHasItems[T gomath.Float, I BoundedItem[ID], ID comparable](qt *Quadtree[T, I, ID], items ...I) error {
	for _, i := range items {
		hasItem := false
		for _, p := range qt.Items {
			if p.GetBounds() == i.GetBounds() {
				hasItem = true
			}
		}

		if !hasItem {
			return fmt.Errorf("quad tree does not have item, %v", i)
		}
	}

	return nil
}

func assertItems[T gomath.Float, I BoundedItem[ID], ID comparable](t *testing.T, qt *Quadtree[T, I, ID], ne int, nw int, se int, sw int) {
	require.NotNil(t, qt.NorthEast)
	assert.Len(t, qt.SouthWest.Items, sw, "Incorrect number of items in south west node")
	assert.Len(t, qt.NorthEast.Items, ne, "Incorrect number of items in north east node")
	assert.Len(t, qt.SouthEast.Items, se, "Incorrect number of items in south east node")
	assert.Len(t, qt.NorthWest.Items, nw, "Incorrect number of items in north west node")
}
