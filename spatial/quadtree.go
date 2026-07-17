package spatial

import (
	"fmt"
	"iter"
	"strings"

	"github.com/wolfgarnet/gomath"
	"github.com/wolfgarnet/gomath/intersection"
)

const (
	NodeCapacity = 4
)

// Quadtree represents a quadtree that may contain elements inside the bounds defined as its boundary.
// The quadtree has a node capacity when exceeded it divides itself into four new sub-quadtrees.
// An element can only be contained in one quadtree, the quadtree that contains the element completely.
type Quadtree[T gomath.Number, I BoundedItem[ID], ID comparable] struct {
	Items     []I
	Boundary  gomath.SquaredAABB2D[T]
	NorthEast *Quadtree[T, I, ID]
	SouthEast *Quadtree[T, I, ID]
	SouthWest *Quadtree[T, I, ID]
	NorthWest *Quadtree[T, I, ID]

	// Checker determines the specific bounds check for the quadtree.
	// By default, the bounds check is using floating points, but is can be set to use integer as well.
	Checker BoundsCheck

	parent   *Quadtree[T, I, ID]
	maxDepth int
	depth    int
	allItems map[ID]*Quadtree[T, I, ID]

	// mayExpand determines if the quad tree may expand itself
	mayExpand bool
}

func NewQuadtree[T gomath.Number, I BoundedItem[ID], ID comparable](x, y, halfSize T) *Quadtree[T, I, ID] {
	return newQuadtree[T, I, ID](x, y, halfSize, nil, 0, DefaultBoundsCheck{}, 8)
}

func NewIntegerQuadtree[T gomath.Number, I BoundedItem[ID], ID comparable](x, y, halfSize T) *Quadtree[T, I, ID] {
	return newQuadtree[T, I, ID](x, y, halfSize, nil, 0, IntegerBoundsCheck{}, 8)
}

func newQuadtree[T gomath.Number, I BoundedItem[ID], ID comparable](x, y, halfSize T, parent *Quadtree[T, I, ID], depth int, checker BoundsCheck, maxDepth int) *Quadtree[T, I, ID] {
	qt := &Quadtree[T, I, ID]{
		parent:   parent,
		Boundary: gomath.NewSquaredAABB2DXY(x, y, halfSize),
		maxDepth: maxDepth,
		depth:    depth,
		Checker:  checker,
	}

	return qt
}

// SetMayExpand sets whether the quad tree may expand itself if items outside its boundary is inserted.
// Can only be set for the root tree.
func (tree *Quadtree[T, I, ID]) SetMayExpand(mayExpand bool) {
	if tree.depth > 0 {
		return
	}
	tree.mayExpand = mayExpand
}

func (tree *Quadtree[T, I, ID]) Depth() (d int) {
	subQT := tree
	for subQT != nil {
		subQT = subQT.NorthEast
		d++
	}

	return
}

// FindQuadTree will find the quadtree storing the given item.
func (tree *Quadtree[T, I, ID]) FindQuadTree(i I) (*Quadtree[T, I, ID], bool) {
	root := tree
	for root.parent != nil {
		root = root.parent
	}

	if root.allItems == nil {
		return nil, false
	}

	v, ok := root.allItems[i.ID()]
	return v, ok
}

func (tree *Quadtree[T, I, ID]) Subdivide() {
	size := tree.Boundary.HalfSize / 2
	tree.NorthEast = newQuadtree(tree.Boundary.Center.X+size, tree.Boundary.Center.Y+size, size, tree, tree.depth+1, tree.Checker, tree.maxDepth)
	tree.SouthEast = newQuadtree(tree.Boundary.Center.X+size, tree.Boundary.Center.Y-size, size, tree, tree.depth+1, tree.Checker, tree.maxDepth)
	tree.SouthWest = newQuadtree(tree.Boundary.Center.X-size, tree.Boundary.Center.Y-size, size, tree, tree.depth+1, tree.Checker, tree.maxDepth)
	tree.NorthWest = newQuadtree(tree.Boundary.Center.X-size, tree.Boundary.Center.Y+size, size, tree, tree.depth+1, tree.Checker, tree.maxDepth)

	// Re-arrange points
	points := tree.Items
	tree.Items = make([]I, 0)
	for _, p := range points {
		tree.Insert(p)
	}
}

func (tree *Quadtree[T, I, ID]) Collapse() {
	// Leaf node, don't collapse
	if tree.NorthEast == nil {
		return
	}

	collapse := false
	// Make sure none of the sub quad tree have children.
	if tree.NorthEast.NorthEast == nil && tree.SouthEast.NorthEast == nil && tree.SouthWest.NorthEast == nil && tree.NorthWest.NorthEast == nil &&
		tree.Unique() <= NodeCapacity {
		collapse = true
	}

	if !collapse {
		tree.NorthEast.Collapse()
		tree.SouthEast.Collapse()
		tree.SouthWest.Collapse()
		tree.NorthWest.Collapse()
	} else {
		tree.collapse()
	}
}

func (tree *Quadtree[T, I, ID]) collapse() {
	// Gather all child items into this quadtrees items
	tree.Items = append(tree.Items, tree.NorthEast.Items...)
	tree.NorthEast = nil
	tree.Items = append(tree.Items, tree.SouthEast.Items...)
	tree.SouthEast = nil
	tree.Items = append(tree.Items, tree.SouthWest.Items...)
	tree.SouthWest = nil
	tree.Items = append(tree.Items, tree.NorthWest.Items...)
	tree.NorthWest = nil

	// Only unique items
	u := make(map[ID]I, len(tree.Items))
	for _, i := range tree.Items {
		u[i.ID()] = i
	}

	tree.Items = make([]I, len(u))
	i := 0
	for _, it := range u {
		tree.Items[i] = it
		i++
	}
}

func (tree *Quadtree[T, I, ID]) Size() int {
	if tree.NorthEast == nil {
		return len(tree.Items)
	}

	size := tree.NorthEast.Size()
	size += tree.SouthEast.Size()
	size += tree.SouthWest.Size()
	size += tree.NorthWest.Size()

	return size + len(tree.Items)
}

// Unique will return the number of unique items in the quadtree.
func (tree *Quadtree[T, I, ID]) Unique() int {
	u := map[ID]struct{}{}
	for _, i := range tree.NorthEast.Items {
		u[i.ID()] = struct{}{}
	}
	for _, i := range tree.SouthEast.Items {
		u[i.ID()] = struct{}{}
	}
	for _, i := range tree.SouthWest.Items {
		u[i.ID()] = struct{}{}
	}
	for _, i := range tree.NorthWest.Items {
		u[i.ID()] = struct{}{}
	}

	return len(u)
}

// Update will update whether or not the item will be in a new quad
func (tree *Quadtree[T, I, ID]) Update(item I, newPos gomath.Vector2[T]) {
	// Test to see if the update changes the quad
	quad, ok := tree.FindQuadTree(item)
	//if ok && quad.Boundary.Contains(newPos) {
	if ok && intersection.SquaredAABBPoint(quad.Boundary, newPos) {
		return
	}
	tree.Remove(item)
	tree.Insert(item)
}

func (tree *Quadtree[T, I, ID]) Remove(item I) bool {
	if !tree.Checker.Check(tree.Boundary, item.GetBounds()) {
		return false
	}

	i := 0
	wasRemoved := false
	for _, pp := range tree.Items {
		if pp.ID() == item.ID() {
			wasRemoved = true
			continue
		}

		tree.Items[i] = pp
		i++
	}
	tree.Items = tree.Items[:i]

	if tree.NorthEast != nil {
		if !wasRemoved {
			wasRemoved = tree.NorthEast.Remove(item)
			if !wasRemoved {
				wasRemoved = tree.SouthEast.Remove(item)
				if !wasRemoved {
					wasRemoved = tree.SouthWest.Remove(item)
					if !wasRemoved {
						wasRemoved = tree.NorthWest.Remove(item)
					}
				}
			}
		}
	}

	if wasRemoved {
		tree.Collapse()
	}

	return wasRemoved
}

func (tree *Quadtree[T, I, ID]) Expand() {
	oldSize := tree.Boundary.HalfSize
	tree.Boundary.HalfSize *= 2
	if tree.NorthEast == nil {
		// No need to do more
		return
	}

	halfSize := oldSize / 2

	// North West (south east)
	newNorthWest := newQuadtree(tree.Boundary.Center.X-oldSize, tree.Boundary.Center.Y+oldSize, oldSize, tree, tree.depth+1, tree.Checker, tree.maxDepth)
	tree.NorthWest.addDepth()
	newNorthWest.SouthEast = tree.NorthWest
	tree.NorthWest.parent = newNorthWest
	newNorthWest.SouthWest = newQuadtree(newNorthWest.Boundary.Center.X-halfSize, newNorthWest.Boundary.Center.Y-halfSize, halfSize, newNorthWest, newNorthWest.depth+1, tree.Checker, tree.maxDepth)
	newNorthWest.NorthWest = newQuadtree(newNorthWest.Boundary.Center.X-halfSize, newNorthWest.Boundary.Center.Y+halfSize, halfSize, newNorthWest, newNorthWest.depth+1, tree.Checker, tree.maxDepth)
	newNorthWest.NorthEast = newQuadtree(newNorthWest.Boundary.Center.X+halfSize, newNorthWest.Boundary.Center.Y+halfSize, halfSize, newNorthWest, newNorthWest.depth+1, tree.Checker, tree.maxDepth)
	tree.NorthWest = newNorthWest

	// North East (south west)
	newNorthEast := newQuadtree(tree.Boundary.Center.X+oldSize, tree.Boundary.Center.Y+oldSize, oldSize, tree, tree.depth+1, tree.Checker, tree.maxDepth)
	tree.NorthEast.addDepth()
	newNorthEast.SouthWest = tree.NorthEast
	tree.NorthEast.parent = newNorthEast
	newNorthEast.SouthEast = newQuadtree(newNorthEast.Boundary.Center.X+halfSize, newNorthEast.Boundary.Center.Y-halfSize, halfSize, newNorthEast, newNorthEast.depth+1, tree.Checker, tree.maxDepth)
	newNorthEast.NorthWest = newQuadtree(newNorthEast.Boundary.Center.X-halfSize, newNorthEast.Boundary.Center.Y+halfSize, halfSize, newNorthEast, newNorthEast.depth+1, tree.Checker, tree.maxDepth)
	newNorthEast.NorthEast = newQuadtree(newNorthEast.Boundary.Center.X+halfSize, newNorthEast.Boundary.Center.Y+halfSize, halfSize, newNorthEast, newNorthEast.depth+1, tree.Checker, tree.maxDepth)
	tree.NorthEast = newNorthEast

	// South East (north west)
	newSouthEast := newQuadtree(tree.Boundary.Center.X+oldSize, tree.Boundary.Center.Y-oldSize, oldSize, tree, tree.depth+1, tree.Checker, tree.maxDepth)
	tree.SouthEast.addDepth()
	newSouthEast.NorthWest = tree.SouthEast
	tree.SouthEast.parent = newSouthEast
	newSouthEast.SouthEast = newQuadtree(newSouthEast.Boundary.Center.X+halfSize, newSouthEast.Boundary.Center.Y-halfSize, halfSize, newSouthEast, newSouthEast.depth+1, tree.Checker, tree.maxDepth)
	newSouthEast.SouthWest = newQuadtree(newSouthEast.Boundary.Center.X-halfSize, newSouthEast.Boundary.Center.Y-halfSize, halfSize, newSouthEast, newSouthEast.depth+1, tree.Checker, tree.maxDepth)
	newSouthEast.NorthEast = newQuadtree(newSouthEast.Boundary.Center.X+halfSize, newSouthEast.Boundary.Center.Y+halfSize, halfSize, newSouthEast, newSouthEast.depth+1, tree.Checker, tree.maxDepth)
	tree.SouthEast = newSouthEast

	// South West (north east)
	newSouthWest := newQuadtree(tree.Boundary.Center.X-oldSize, tree.Boundary.Center.Y-oldSize, oldSize, tree, tree.depth+1, tree.Checker, tree.maxDepth)
	tree.SouthWest.addDepth()
	newSouthWest.NorthEast = tree.SouthWest
	tree.SouthWest.parent = newSouthWest
	newSouthWest.SouthEast = newQuadtree(newSouthWest.Boundary.Center.X+halfSize, newSouthWest.Boundary.Center.Y-halfSize, halfSize, newSouthWest, newSouthWest.depth+1, tree.Checker, tree.maxDepth)
	newSouthWest.SouthWest = newQuadtree(newSouthWest.Boundary.Center.X-halfSize, newSouthWest.Boundary.Center.Y-halfSize, halfSize, newSouthWest, newSouthWest.depth+1, tree.Checker, tree.maxDepth)
	newSouthWest.NorthWest = newQuadtree(newSouthWest.Boundary.Center.X-halfSize, newSouthWest.Boundary.Center.Y+halfSize, halfSize, newSouthWest, newSouthWest.depth+1, tree.Checker, tree.maxDepth)
	tree.SouthWest = newSouthWest
}

func (tree *Quadtree[T, I, ID]) addDepth() {
	tree.depth++
	if tree.NorthEast == nil {
		return
	}
	tree.NorthEast.addDepth()
	tree.NorthWest.addDepth()
	tree.SouthWest.addDepth()
	tree.SouthEast.addDepth()
}

func (tree *Quadtree[T, I, ID]) Insert(item I) bool {
	// First, verify that the item belongs in this tree.
	if !tree.Checker.Check(tree.Boundary, item.GetBounds()) {
		// If we're outside the bounds of the outermost quadtree, we might want to expand the quadtree.
		if tree.depth == 0 && tree.mayExpand {
			tree.Expand()
			return tree.Insert(item)
		} else {
			return false
		}
	}

	// If max depth has been reached, just put it in!
	if tree.depth == tree.maxDepth {
		tree.Items = append(tree.Items, item)
		return true
	}

	// No sub quad trees and more room, just plug it in
	if tree.NorthEast == nil {
		if len(tree.Items) < NodeCapacity {
			tree.Items = append(tree.Items, item)
			return true
		}

		// Obviously there wasn't any room
		tree.Subdivide()
	}

	// Figure out which children the item can be in.
	var insertInto *Quadtree[T, I, ID]
	multiple := false
	if tree.NorthEast.ContainedIn(item) {
		insertInto = tree.NorthEast
	}
	if tree.SouthEast.ContainedIn(item) {
		if insertInto != nil {
			multiple = true
		}
		insertInto = tree.SouthEast
	}
	if tree.SouthWest.ContainedIn(item) {
		if insertInto != nil {
			multiple = true
		}
		insertInto = tree.SouthWest
	}
	if tree.NorthWest.ContainedIn(item) {
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
	} else {
		if insertInto != nil {
			return insertInto.Insert(item)
		}

		return false
	}
}

func (tree *Quadtree[T, I, ID]) Clear() {
	tree.NorthEast = nil
	tree.NorthWest = nil
	tree.SouthEast = nil
	tree.SouthWest = nil

	tree.Items = nil
}

func (tree *Quadtree[T, I, ID]) ContainedIn(item I) bool {
	//return intersection.IntersectingSquaredAABB(tree.Boundary, item.GetBounds())
	return tree.Checker.Check(tree.Boundary, item.GetBounds())
}

type QueryResultWrapper[I BoundedItem[ID], ID comparable] struct {
	Items []I
}

func NewQueryResultWrapper[I BoundedItem[ID], ID comparable](items []I) *QueryResultWrapper[I, ID] {
	return &QueryResultWrapper[I, ID]{items}
}

// RemoveDuplicates will remove duplicates.
// Remember to use the wrapped result.
func (w *QueryResultWrapper[I, ID]) RemoveDuplicates() {
	u := map[ID]I{}
	for _, i := range w.Items {
		u[i.ID()] = i
	}

	i := 0
	for _, it := range u {
		w.Items[i] = it
		i++
	}
	w.Items = w.Items[:i]
}

func (tree *Quadtree[T, I, ID]) Clone() *Quadtree[T, I, ID] {
	newQT := NewQuadtree[T, I, ID](tree.Boundary.Center.X, tree.Boundary.Center.Y, tree.Boundary.HalfSize)
	if tree.NorthEast != nil {
		newQT.NorthEast = tree.NorthEast.Clone()
		newQT.SouthEast = tree.SouthEast.Clone()
		newQT.SouthWest = tree.SouthWest.Clone()
		newQT.NorthWest = tree.NorthWest.Clone()
	} else {
		newQT.Items = make([]I, len(tree.Items))
		for i, item := range tree.Items {
			newQT.Items[i] = item
		}
	}
	return newQT
}

func (tree *Quadtree[T, I, ID]) Stringify() {
	tree.stringify(0)
}

func (tree *Quadtree[T, I, ID]) stringify(lvl int) {
	indent := fmt.Sprintf("%s", strings.Repeat(" ", lvl*2))
	//fmt.Printf("%sBounds: %v, depth: %d, points: %v\n", indent, tree.Boundary, tree.depth, tree.Items)
	fmt.Printf("Bounds: %v, depth: %d, points: %v\n", tree.Boundary, tree.depth, tree.Items)
	if tree.NorthEast != nil {
		fmt.Printf("%s* North east: ", indent)
		tree.NorthEast.stringify(lvl + 1)
		fmt.Printf("%s* South east: ", indent)
		tree.SouthEast.stringify(lvl + 1)
		fmt.Printf("%s* South west: ", indent)
		tree.SouthWest.stringify(lvl + 1)
		fmt.Printf("%s* North west: ", indent)
		tree.NorthWest.stringify(lvl + 1)
	} else {
		//fmt.Printf("%sNo sub division\n", indent)
	}
}

func (tree *Quadtree[T, I, ID]) Visit(cb func(tree *Quadtree[T, I, ID])) {
	cb(tree)
	if tree.NorthEast == nil {
		return
	}

	tree.NorthEast.Visit(cb)
	tree.SouthEast.Visit(cb)
	tree.SouthWest.Visit(cb)
	tree.NorthWest.Visit(cb)
}

func (tree *Quadtree[T, I, ID]) IteratorWithBoundedItem(item I) iter.Seq[I] {
	return func(yield func(I) bool) {
		IterateQuadtree[T, I, ID](tree, item.GetBounds(), func(b I) bool {
			return yield(b)
		})
	}
}

func (tree *Quadtree[T, I, ID]) Iterator(containing any) iter.Seq[I] {
	return func(yield func(I) bool) {
		IterateQuadtree[T, I, ID](tree, containing, func(b I) bool {
			return yield(b)
		})
	}
}

func (tree *Quadtree[T, I, ID]) IteratorWithoutCheck(containing any) iter.Seq[I] {
	return func(yield func(I) bool) {
		IterateQuadtree[T, I, ID](tree, containing, func(b I) bool {
			return yield(b)
		})
	}
}
