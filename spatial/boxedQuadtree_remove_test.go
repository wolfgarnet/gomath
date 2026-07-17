package spatial

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
)

func boxedItem(id uint64, x, y, half float32) BoxedItem[float32, uint64] {
	return BoxedItem[float32, uint64]{
		ItemID: id,
		AABB:   gomath.AABB[float32]{Center: gomath.Vector2[float32]{X: x, Y: y}, HalfWidth: half, HalfHeight: half},
	}
}

func countBoxedItems(tree *BoxedQuadtree[float32, uint64]) int {
	n := len(tree.Items)
	if tree.NorthEast != nil {
		n += countBoxedItems(tree.NorthEast)
		n += countBoxedItems(tree.SouthEast)
		n += countBoxedItems(tree.SouthWest)
		n += countBoxedItems(tree.NorthWest)
	}
	return n
}

func boxedContains(tree *BoxedQuadtree[float32, uint64], id uint64) bool {
	for _, it := range tree.Items {
		if it.ItemID == id {
			return true
		}
	}
	if tree.NorthEast != nil {
		return boxedContains(tree.NorthEast, id) ||
			boxedContains(tree.SouthEast, id) ||
			boxedContains(tree.SouthWest, id) ||
			boxedContains(tree.NorthWest, id)
	}
	return false
}

// TestBoxedQuadtree_Remove asserts an item can be removed regardless of its
// position within its node's slice. Remove previously never advanced its index,
// so it only ever inspected Items[0] and failed to find anything else.
func TestBoxedQuadtree_Remove(t *testing.T) {
	t.Run("removes item that is not first in its node", func(t *testing.T) {
		tree := NewBoxedQuadtree[float32, uint64](0, 0, 100)
		// Four small items at the same spot all land in the root node (nodeSize 4).
		for id := uint64(1); id <= 4; id++ {
			tree.Insert(boxedItem(id, 1, 1, 0.5))
		}
		assert.Len(t, tree.Items, 4)

		assert.True(t, tree.Remove(boxedItem(3, 1, 1, 0.5)), "removing a middle item should succeed")
		assert.Equal(t, 3, countBoxedItems(tree))
		assert.False(t, boxedContains(tree, 3), "item 3 should be gone")
		assert.True(t, boxedContains(tree, 1))
		assert.True(t, boxedContains(tree, 2))
		assert.True(t, boxedContains(tree, 4))
	})

	t.Run("removes item from a subdivided node", func(t *testing.T) {
		tree := NewBoxedQuadtree[float32, uint64](0, 0, 100)
		// Spread items so the tree subdivides.
		tree.Insert(boxedItem(1, 10, 10, 0.5))
		tree.Insert(boxedItem(2, -10, 10, 0.5))
		tree.Insert(boxedItem(3, -10, -10, 0.5))
		tree.Insert(boxedItem(4, 10, -10, 0.5))
		tree.Insert(boxedItem(5, 20, 20, 0.5))
		tree.Insert(boxedItem(6, 30, 30, 0.5))
		before := countBoxedItems(tree)

		assert.True(t, tree.Remove(boxedItem(6, 30, 30, 0.5)))
		assert.Equal(t, before-1, countBoxedItems(tree))
		assert.False(t, boxedContains(tree, 6))
	})

	t.Run("removing an absent item reports false", func(t *testing.T) {
		tree := NewBoxedQuadtree[float32, uint64](0, 0, 100)
		tree.Insert(boxedItem(1, 1, 1, 0.5))

		assert.False(t, tree.Remove(boxedItem(99, 1, 1, 0.5)))
		assert.Equal(t, 1, countBoxedItems(tree))
	})

	t.Run("removes every item one by one", func(t *testing.T) {
		tree := NewBoxedQuadtree[float32, uint64](0, 0, 100)
		ids := []uint64{1, 2, 3, 4, 5, 6, 7, 8}
		for i, id := range ids {
			tree.Insert(boxedItem(id, float32(i*5), float32(i*5), 0.5))
		}
		for _, id := range ids {
			assert.Truef(t, tree.Remove(boxedItem(id, 0, 0, 0.5)), "removing id %d", id)
		}
		assert.Equal(t, 0, countBoxedItems(tree))
	})
}

// TestBoxedQuadtree_Iterator asserts the iterator yields exactly the items whose
// AABB overlaps the query box.
func TestBoxedQuadtree_Iterator(t *testing.T) {
	tree := NewBoxedQuadtree[float32, uint64](0, 0, 100)
	tree.Insert(boxedItem(1, 10, 10, 1))
	tree.Insert(boxedItem(2, 12, 12, 1))
	tree.Insert(boxedItem(3, -50, -50, 1))

	query := gomath.AABB[float32]{Center: gomath.Vector2[float32]{X: 11, Y: 11}, HalfWidth: 5, HalfHeight: 5}

	found := map[uint64]bool{}
	for item := range tree.Iterator(query) {
		found[item.ItemID] = true
	}

	assert.True(t, found[1], "item 1 overlaps the query")
	assert.True(t, found[2], "item 2 overlaps the query")
	assert.False(t, found[3], "item 3 is far from the query")
}
