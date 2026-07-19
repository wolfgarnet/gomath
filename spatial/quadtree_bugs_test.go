package spatial

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
)

func point2D(id uint64, x, y float32) TestItem2D {
	return TestItem2D{id: id, bounds: gomath.Vector2[float32]{X: x, Y: y}}
}

// TestQuadtree_Depth_AsymmetricSubtree pins that Depth reports the deepest
// quadrant, not just the length of the North East spine. The bug walked only
// tree.NorthEast, so a tree subdivided under North West reported the wrong depth.
func TestQuadtree_Depth_AsymmetricSubtree(t *testing.T) {
	tree := NewQuadtree[float32, TestItem2D, uint64](0, 0, 10)
	assert.Equal(t, 1, tree.Depth(), "a leaf is one level")

	tree.Subdivide()
	assert.Equal(t, 2, tree.Depth(), "one subdivision is two levels")

	// Deepen a NON-North-East quadrant. North East stays a leaf.
	tree.NorthWest.Subdivide()
	assert.Equal(t, 3, tree.Depth(), "depth must follow the deepest quadrant (North West)")
}

func box2D(id uint64, x, y, half float32) TestItem2D {
	return TestItem2D{id: id, bounds: gomath.NewSquaredAABB2DXY[float32](x, y, half)}
}

// assertParentsConsistent walks the tree asserting every child points back at its
// actual parent, and that depth increases by one per level.
func assertParentsConsistent(t *testing.T, tree *Quadtree[float32, TestItem2D, uint64]) {
	t.Helper()
	if tree.NorthEast == nil {
		return
	}
	for _, child := range []*Quadtree[float32, TestItem2D, uint64]{
		tree.NorthEast, tree.SouthEast, tree.SouthWest, tree.NorthWest,
	} {
		assert.Samef(t, tree, child.parent, "child at depth %d has the wrong parent", child.depth)
		assert.Equalf(t, tree.depth+1, child.depth, "child depth should be parent depth + 1")
		assertParentsConsistent(t, child)
	}
}

// TestQuadtree_ExpandKeepsParentsConsistent pins that Expand rebuilds the tree
// with correct parent links. Expand previously passed newNorthWest as the parent
// of every new quadrant's children, so three quarters of the rebuilt tree pointed
// at the wrong parent.
func TestQuadtree_ExpandKeepsParentsConsistent(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D, uint64](0, 0, 10)
	qt.SetMayExpand(true)

	// Five points force a subdivision so Expand hits its restructuring branch.
	qt.Insert(point2D(1, 1, 1))
	qt.Insert(point2D(2, 2, 2))
	qt.Insert(point2D(3, 3, 3))
	qt.Insert(point2D(4, 4, 4))
	qt.Insert(point2D(5, 6, 6))
	assert.NotNil(t, qt.NorthEast, "tree should have subdivided")

	// An item far outside the boundary triggers Expand.
	qt.Insert(point2D(6, 25, 25))

	assert.Nil(t, qt.parent, "root parent stays nil")
	assertParentsConsistent(t, qt)
}

// TestQuadtree_Update_ReinsertsWithNewBounds pins that Update relocates an item:
// it removes the item by ID and re-inserts it at its current bounds, leaving no
// stale copy at the old location.
func TestQuadtree_Update_ReinsertsWithNewBounds(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D, uint64](0, 0, 10)
	qt.Insert(box2D(1, -6, -6, 0.5)) // starts in the SW
	assert.Equal(t, 1, qt.Size())

	moved := box2D(1, 6, 6, 0.5) // same ID, now in the NE
	assert.True(t, qt.Update(moved), "re-insert of in-bounds item succeeds")
	assert.Equal(t, 1, qt.Size(), "still exactly one item, no ghost")

	atNew := QueryQuadtreeWithResult(qt, gomath.NewSquaredAABB2DXY[float32](6, 6, 0.5))
	assert.Len(t, atNew, 1, "item is found at its new location")
	atOld := QueryQuadtreeWithResult(qt, gomath.NewSquaredAABB2DXY[float32](-6, -6, 0.5))
	assert.Len(t, atOld, 0, "item is gone from its old location")
}

// TestIterateQuadtreeWithoutCheck_SkipsItemLevelCheck pins that the "without
// check" traversal returns every item in a node whose boundary overlaps the
// query, without testing individual items. It previously recursed into the
// checked variant, so child items were filtered after all.
func TestIterateQuadtreeWithoutCheck_SkipsItemLevelCheck(t *testing.T) {
	qt := NewQuadtree[float32, TestItem2D, uint64](0, 0, 10)
	// One item per quadrant, plus a second in the NE, to subdivide exactly once.
	qt.Insert(box2D(1, 6, 6, 0.5))   // NE
	qt.Insert(box2D(2, -6, 6, 0.5))  // NW
	qt.Insert(box2D(3, -6, -6, 0.5)) // SW
	qt.Insert(box2D(4, 6, -6, 0.5))  // SE
	qt.Insert(box2D(5, 7, 7, 0.5))   // NE again -> forces subdivision
	assert.NotNil(t, qt.NorthEast, "tree should have subdivided")

	// A small box near the origin. It overlaps the NE quadrant's boundary
	// ([0,10]x[0,10]) but not the NE items at (6,6) and (7,7).
	query := gomath.NewSquaredAABB2DXY[float32](1, 1, 0.5)

	found := map[uint64]bool{}
	IterateQuadtreeWithoutCheck(qt, query, func(item TestItem2D) bool {
		found[item.id] = true
		return true
	})

	assert.True(t, found[1], "unchecked traversal should return NE item 1 regardless of item-level overlap")
	assert.True(t, found[5], "unchecked traversal should return NE item 5")

	// The checked traversal, by contrast, must exclude them.
	checked := map[uint64]bool{}
	IterateQuadtree(qt, query, func(item TestItem2D) bool {
		checked[item.id] = true
		return true
	})
	assert.False(t, checked[1], "checked traversal should exclude the non-overlapping item")
}
