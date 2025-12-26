package spatial

import (
	"iter"

	"github.com/wolfgarnet/gomath"
	"golang.org/x/exp/constraints"
)

type BoxedItem[T constraints.Float, ID comparable] struct {
	ItemID ID
	AABB   gomath.AABB[T]
}

type BoxedQuadtree[T constraints.Float, I comparable] struct {
	Items []BoxedItem[T, I]

	Boundary  gomath.SquaredAABB2D[T]
	NorthEast *BoxedQuadtree[T, I]
	SouthEast *BoxedQuadtree[T, I]
	SouthWest *BoxedQuadtree[T, I]
	NorthWest *BoxedQuadtree[T, I]

	parent   *BoxedQuadtree[T, I]
	nodeSize int
	maxDepth int
	depth    int
}

func NewBoxedQuadtree[T constraints.Float, I comparable](x, y, halfSize T) *BoxedQuadtree[T, I] {
	return newBoxedQuadtree[T, I](x, y, halfSize, nil, 0, 8, 4)
}

func newBoxedQuadtree[T constraints.Float, I comparable](x, y, halfSize T, parent *BoxedQuadtree[T, I], depth int, maxDepth int, nodeSize int) *BoxedQuadtree[T, I] {
	return &BoxedQuadtree[T, I]{
		Boundary: gomath.SquaredAABB2D[T]{
			Center:   gomath.Vector2[T]{X: x, Y: y},
			HalfSize: halfSize,
		},
		parent:   parent,
		maxDepth: maxDepth,
		nodeSize: nodeSize,
		depth:    depth,
	}
}

func (tree *BoxedQuadtree[T, I]) Subdivide() {
	size := tree.Boundary.HalfSize / 2
	tree.NorthEast = newBoxedQuadtree(tree.Boundary.Center.X+size, tree.Boundary.Center.Y+size, size, tree, tree.depth+1, tree.maxDepth, tree.nodeSize)
	tree.SouthEast = newBoxedQuadtree(tree.Boundary.Center.X+size, tree.Boundary.Center.Y-size, size, tree, tree.depth+1, tree.maxDepth, tree.nodeSize)
	tree.SouthWest = newBoxedQuadtree(tree.Boundary.Center.X-size, tree.Boundary.Center.Y-size, size, tree, tree.depth+1, tree.maxDepth, tree.nodeSize)
	tree.NorthWest = newBoxedQuadtree(tree.Boundary.Center.X-size, tree.Boundary.Center.Y+size, size, tree, tree.depth+1, tree.maxDepth, tree.nodeSize)

	// Re-arrange points
	points := tree.Items
	tree.Items = make([]BoxedItem[T, I], 0)
	for _, p := range points {
		tree.Insert(p)
	}
}

func (tree *BoxedQuadtree[T, I]) Clear() {
	tree.Items = make([]BoxedItem[T, I], 0)
	tree.NorthEast = nil
	tree.SouthEast = nil
	tree.SouthWest = nil
	tree.NorthWest = nil
	tree.parent = nil
}

//go:inline
func (tree *BoxedQuadtree[T, I]) ContainedIn(item BoxedItem[T, I]) bool {
	return SquaredAABBAABB(tree.Boundary, item.AABB)
}

func (tree *BoxedQuadtree[T, I]) Insert(item BoxedItem[T, I]) bool {
	// First, verify that the item belongs in this tree.
	//if !tree.ContainedIn(item) {
	if !SquaredAABBAABB(tree.Boundary, item.AABB) {
		return false
	}

	// If max depth has been reached, just put it in!
	if tree.depth == tree.maxDepth {
		tree.Items = append(tree.Items, item)
		return true
	}

	// No sub quad trees and more room, just plug it in
	if tree.NorthEast == nil {
		if len(tree.Items) < tree.nodeSize {
			tree.Items = append(tree.Items, item)
			return true
		}

		// Obviously there wasn't any room
		tree.Subdivide()
	}

	// Figure out which children the item can be in.
	var insertInto *BoxedQuadtree[T, I]
	multiple := false
	//if tree.NorthEast.ContainedIn(item) {
	if SquaredAABBAABB(tree.NorthEast.Boundary, item.AABB) {
		insertInto = tree.NorthEast
	}
	//if tree.SouthEast.ContainedIn(item) {
	if SquaredAABBAABB(tree.SouthEast.Boundary, item.AABB) {
		if insertInto != nil {
			multiple = true
		}
		insertInto = tree.SouthEast
	}
	//if tree.SouthWest.ContainedIn(item) {
	if SquaredAABBAABB(tree.SouthWest.Boundary, item.AABB) {
		if insertInto != nil {
			multiple = true
		}
		insertInto = tree.SouthWest
	}
	//if tree.NorthWest.ContainedIn(item) {
	if SquaredAABBAABB(tree.NorthWest.Boundary, item.AABB) {
		if insertInto != nil {
			multiple = true
		}
		insertInto = tree.NorthWest
	}

	if multiple {
		// The item can fit in multiple leafs,
		// put it into the current sub tree.
		tree.Items = append(tree.Items, item)
		return true
	}

	if insertInto != nil {
		return insertInto.Insert(item)
	}

	return false
}

func (tree *BoxedQuadtree[T, I]) Remove(item BoxedItem[T, I]) bool {
	var i int
	for range tree.Items {
		if tree.Items[i].ItemID == item.ItemID {
			tree.Items = append(tree.Items[:i], tree.Items[i+1:]...)
			return true
		}
	}

	if tree.NorthEast == nil {
		return false
	}
	if tree.NorthEast.Remove(item) {
		return true
	}
	if tree.SouthEast.Remove(item) {
		return true
	}
	if tree.SouthWest.Remove(item) {
		return true
	}
	return tree.NorthWest.Remove(item)
}

func (tree *BoxedQuadtree[T, I]) Iterator(bounds gomath.AABB[T]) iter.Seq[BoxedItem[T, I]] {
	return func(yield func(item BoxedItem[T, I]) bool) {
		IteratePointQuadtree(tree, bounds, yield)
	}
}

func IteratePointQuadtree[T constraints.Float, I comparable](tree *BoxedQuadtree[T, I], bounds gomath.AABB[T], yield func(item BoxedItem[T, I]) bool) {
	// Determine if the contained is inside the quadtree boundary
	if !SquaredAABBAABB(tree.Boundary, bounds) {
		return
	}
	// Leaf node
	for _, item := range tree.Items {
		//if tree.Checker.Check(item.GetBounds(), containing) {
		if AABBAABB3(item.AABB, bounds) {
			if !yield(item) {
				return
			}
		}
	}

	if tree.NorthEast == nil {
		return
	}

	IteratePointQuadtree(tree.NorthEast, bounds, yield)
	IteratePointQuadtree(tree.SouthEast, bounds, yield)
	IteratePointQuadtree(tree.SouthWest, bounds, yield)
	IteratePointQuadtree(tree.NorthWest, bounds, yield)
}
