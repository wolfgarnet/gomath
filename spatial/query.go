package spatial

import "github.com/wolfgarnet/gomath"

// QueryQuadtree queries the quadtree given a containing geometry compatible with intersection.Intersecting2D[T]().
// The resulting items are piped into the drain channel.
func QueryQuadtree[T gomath.Number](tree *Quadtree[T], containing interface{}, drain chan BoundedItem) {
	// Determine if the contained is inside the quadtree boundary
	if !tree.Checker.Check(tree.Boundary, containing) {
		return
	}
	// Leaf node
	for _, item := range tree.Items {
		if tree.Checker.Check(item.GetBounds(), containing) {
			drain <- item
		}
	}

	if tree.NorthEast == nil {
		return
	}

	QueryQuadtree(tree.NorthEast, containing, drain)
	QueryQuadtree(tree.SouthEast, containing, drain)
	QueryQuadtree(tree.SouthWest, containing, drain)
	QueryQuadtree(tree.NorthWest, containing, drain)
}

// QueryQuadtreeWithResult does the same as QueryQuadtree, but returns a result slice.
func QueryQuadtreeWithResult[T gomath.Number](tree *Quadtree[T], containing interface{}) (items []BoundedItem) {
	// Determine if the contained is inside the quadtree boundary
	if !tree.Checker.Check(tree.Boundary, containing) {
		return
	}
	// Leaf node
	for _, item := range tree.Items {
		if tree.Checker.Check(item.GetBounds(), containing) {
			items = append(items, item)
		}
	}

	if tree.NorthEast == nil {
		return
	}

	items = append(items, QueryQuadtreeWithResult(tree.NorthEast, containing)...)
	items = append(items, QueryQuadtreeWithResult(tree.SouthEast, containing)...)
	items = append(items, QueryQuadtreeWithResult(tree.SouthWest, containing)...)
	items = append(items, QueryQuadtreeWithResult(tree.NorthWest, containing)...)

	return items
}

func QueryQuadtreeWithBoundedItem[T gomath.Number](tree *Quadtree[T], containing BoundedItem, drain chan BoundedItem) {
	QueryQuadtree(tree, containing.GetBounds(), drain)
}

func QueryQuadtreeWithBoundedItemWithResult[T gomath.Number](tree *Quadtree[T], containing BoundedItem) (items []BoundedItem) {
	return QueryQuadtreeWithResult(tree, containing.GetBounds())
}

// QueryQuadtreeWithoutCheck queries the quadtree given a containing geometry compatible with intersection.Intersecting2D[T]().
// But, it will not check the individual items against the containing object.
// Thus returning all the items of the sub quadtrees the containing object overlaps.
func QueryQuadtreeWithoutCheck[T gomath.Number](tree *Quadtree[T], containing interface{}, drain chan BoundedItem) {
	// Determine if the contained is inside the quadtree boundary
	if !tree.Checker.Check(tree.Boundary, containing) {
		return
	}
	// Leaf node
	for _, p := range tree.Items {
		drain <- p
	}

	if tree.NorthEast == nil {
		return
	}

	QueryQuadtreeWithoutCheck(tree.NorthEast, containing, drain)
	QueryQuadtreeWithoutCheck(tree.SouthEast, containing, drain)
	QueryQuadtreeWithoutCheck(tree.SouthWest, containing, drain)
	QueryQuadtreeWithoutCheck(tree.NorthWest, containing, drain)
}

// QueryQuadtreeWithoutCheckWithResult does the same as QueryQuadtreeWithoutCheck, but returns a result slice.
func QueryQuadtreeWithoutCheckWithResult[T gomath.Number](tree *Quadtree[T], containing interface{}) (items []BoundedItem) {
	// Determine if the contained is inside the quadtree boundary
	if !tree.Checker.Check(tree.Boundary, containing) {
		return
	}
	// Leaf node
	for _, item := range tree.Items {
		items = append(items, item)
	}

	if tree.NorthEast == nil {
		return
	}

	items = append(items, QueryQuadtreeWithResult(tree.NorthEast, containing)...)
	items = append(items, QueryQuadtreeWithResult(tree.SouthEast, containing)...)
	items = append(items, QueryQuadtreeWithResult(tree.SouthWest, containing)...)
	items = append(items, QueryQuadtreeWithResult(tree.NorthWest, containing)...)

	return items
}

func QueryQuadtreeWithoutCheckWithBoundedItem[T gomath.Number](tree *Quadtree[T], containing BoundedItem, drain chan BoundedItem) {
	QueryQuadtreeWithoutCheck(tree, containing.GetBounds(), drain)
}

func QueryQuadtreeWithBoundedItemWithoutCheckWithResult[T gomath.Number](tree *Quadtree[T], containing BoundedItem) (items []BoundedItem) {
	return QueryQuadtreeWithoutCheckWithResult(tree, containing.GetBounds())
}
