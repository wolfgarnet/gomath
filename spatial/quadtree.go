package spatial

import (
	"fmt"
	"github.com/wolfgarnet/gomath"
	"github.com/wolfgarnet/gomath/intersection"
	"strings"
)

const (
	NodeCapacity = 4
)

type QuadTree[T gomath.Number] struct {
	Items     []BoundedItem
	Boundary  gomath.SquaredAABB2D[T]
	NorthEast *QuadTree[T]
	SouthEast *QuadTree[T]
	SouthWest *QuadTree[T]
	NorthWest *QuadTree[T]
	Checker   BoundsCheck

	parent   *QuadTree[T]
	maxDepth int
	depth    int
	allItems map[BoundedItem]*QuadTree[T]

	// mayExpand determines if the quad tree may expand itself
	mayExpand bool
}

func NewQuadTree[T gomath.Number](x, y, halfSize T) *QuadTree[T] {
	return newQuadTree(x, y, halfSize, nil, 0, DefaultBoundsCheck{}, 8)
}

func NewIntegerQuadTree[T gomath.Number](x, y, halfSize T) *QuadTree[T] {
	return newQuadTree(x, y, halfSize, nil, 0, IntegerBoundsCheck{}, 8)
}

func newQuadTree[T gomath.Number](x, y, halfSize T, parent *QuadTree[T], depth int, checker BoundsCheck, maxDepth int) *QuadTree[T] {
	qt := &QuadTree[T]{
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
func (qt *QuadTree[T]) SetMayExpand(mayExpand bool) {
	if qt.depth > 0 {
		return
	}
	qt.mayExpand = mayExpand
}

func (qt *QuadTree[T]) Depth() (d int) {
	subQT := qt
	for subQT != nil {
		subQT = subQT.NorthEast
		d++
	}

	return
}

// FindQuadTree will find the quadtree storing the given item.
func (qt *QuadTree[T]) FindQuadTree(i BoundedItem) (*QuadTree[T], bool) {
	parent := qt.parent
	for parent != nil {
		parent = parent.parent
	}

	if parent.allItems != nil {
		v, ok := qt.allItems[i]
		return v, ok
	}

	return nil, false
}

func (qt *QuadTree[T]) Subdivide() {
	size := qt.Boundary.HalfSize / 2
	qt.NorthEast = newQuadTree(qt.Boundary.Center.X+size, qt.Boundary.Center.Y+size, size, qt, qt.depth+1, qt.Checker, qt.maxDepth)
	qt.SouthEast = newQuadTree(qt.Boundary.Center.X+size, qt.Boundary.Center.Y-size, size, qt, qt.depth+1, qt.Checker, qt.maxDepth)
	qt.SouthWest = newQuadTree(qt.Boundary.Center.X-size, qt.Boundary.Center.Y-size, size, qt, qt.depth+1, qt.Checker, qt.maxDepth)
	qt.NorthWest = newQuadTree(qt.Boundary.Center.X-size, qt.Boundary.Center.Y+size, size, qt, qt.depth+1, qt.Checker, qt.maxDepth)

	// Re-arrange points
	points := qt.Items
	qt.Items = make([]BoundedItem, 0)
	for _, p := range points {
		qt.Insert(p)
	}
}

func (qt *QuadTree[T]) Collapse() {
	// Leaf node, don't collapse
	if qt.NorthEast == nil {
		return
	}

	collapse := false
	// Make sure none of the sub quad tree have children.
	if qt.NorthEast.NorthEast == nil && qt.SouthEast.NorthEast == nil && qt.SouthWest.NorthEast == nil && qt.NorthWest.NorthEast == nil &&
		qt.Unique() <= NodeCapacity {
		collapse = true
	}

	if !collapse {
		qt.NorthEast.Collapse()
		qt.SouthEast.Collapse()
		qt.SouthWest.Collapse()
		qt.NorthWest.Collapse()
	} else {
		qt.collapse()
	}
}

func (qt *QuadTree[T]) collapse() {
	// Gather all child items into this quadtrees items
	qt.Items = append(qt.Items, qt.NorthEast.Items...)
	qt.NorthEast = nil
	qt.Items = append(qt.Items, qt.SouthEast.Items...)
	qt.SouthEast = nil
	qt.Items = append(qt.Items, qt.SouthWest.Items...)
	qt.SouthWest = nil
	qt.Items = append(qt.Items, qt.NorthWest.Items...)
	qt.NorthWest = nil

	// Only unique items
	u := map[BoundedItem]struct{}{}
	for _, i := range qt.Items {
		u[i] = struct{}{}
	}

	qt.Items = make([]BoundedItem, len(u))
	i := 0
	for ui, _ := range u {
		qt.Items[i] = ui
		i++
	}
}

func (qt *QuadTree[T]) Size() int {
	if qt.NorthEast == nil {
		return len(qt.Items)
	}

	size := qt.NorthEast.Size()
	size += qt.SouthEast.Size()
	size += qt.SouthWest.Size()
	size += qt.NorthWest.Size()

	return size
}

func (qt *QuadTree[T]) Unique() int {
	u := map[BoundedItem]struct{}{}
	for _, i := range qt.NorthEast.Items {
		u[i] = struct{}{}
	}
	for _, i := range qt.SouthEast.Items {
		u[i] = struct{}{}
	}
	for _, i := range qt.SouthWest.Items {
		u[i] = struct{}{}
	}
	for _, i := range qt.NorthWest.Items {
		u[i] = struct{}{}
	}

	return len(u)
}

// Update will update whether or not the item will be in a new quad
func (qt *QuadTree[T]) Update(item BoundedItem, newPos gomath.Vector2[T]) {
	// Test to see if the update changes the quad
	quad, ok := qt.FindQuadTree(item)
	//if ok && quad.Boundary.Contains(newPos) {
	if ok && intersection.SquaredAABBPoint(quad.Boundary, newPos) {
		return
	}
	qt.Remove(item)
	qt.Insert(item)
}

func (qt *QuadTree[T]) Remove(item BoundedItem) bool {
	if !qt.Checker.Check(qt.Boundary, item.GetBounds()) {
		return false
	}

	i := 0
	wasRemoved := false
	for _, pp := range qt.Items {
		if pp == item {
			wasRemoved = true
			continue
		}

		qt.Items[i] = pp
		i++
	}
	qt.Items = qt.Items[:i]

	if qt.NorthEast != nil {
		if !wasRemoved {
			wasRemoved = qt.NorthEast.Remove(item)
			if !wasRemoved {
				wasRemoved = qt.SouthEast.Remove(item)
				if !wasRemoved {
					wasRemoved = qt.SouthWest.Remove(item)
					if !wasRemoved {
						wasRemoved = qt.NorthWest.Remove(item)
					}
				}
			}
		}
	}

	if wasRemoved {
		qt.Collapse()
	}

	return wasRemoved
}

func (qt *QuadTree[T]) Expand() {
	oldSize := qt.Boundary.HalfSize
	qt.Boundary.HalfSize *= 2
	if qt.NorthEast == nil {
		// No need to do more
		return
	}

	halfSize := oldSize / 2

	// North West (south east)
	newNorthWest := newQuadTree(qt.Boundary.Center.X-oldSize, qt.Boundary.Center.Y+oldSize, oldSize, qt, qt.depth+1, qt.Checker, qt.maxDepth)
	qt.NorthWest.addDepth()
	newNorthWest.SouthEast = qt.NorthWest
	newNorthWest.SouthWest = newQuadTree(newNorthWest.Boundary.Center.X-halfSize, newNorthWest.Boundary.Center.Y-halfSize, halfSize, newNorthWest, newNorthWest.depth+1, qt.Checker, qt.maxDepth)
	newNorthWest.NorthWest = newQuadTree(newNorthWest.Boundary.Center.X-halfSize, newNorthWest.Boundary.Center.Y+halfSize, halfSize, newNorthWest, newNorthWest.depth+1, qt.Checker, qt.maxDepth)
	newNorthWest.NorthEast = newQuadTree(newNorthWest.Boundary.Center.X+halfSize, newNorthWest.Boundary.Center.Y+halfSize, halfSize, newNorthWest, newNorthWest.depth+1, qt.Checker, qt.maxDepth)
	qt.NorthWest = newNorthWest

	// North East (south west)
	newNorthEast := newQuadTree(qt.Boundary.Center.X+oldSize, qt.Boundary.Center.Y+oldSize, oldSize, qt, qt.depth+1, qt.Checker, qt.maxDepth)
	qt.NorthEast.addDepth()
	newNorthEast.SouthWest = qt.NorthEast
	newNorthEast.SouthEast = newQuadTree(newNorthEast.Boundary.Center.X+halfSize, newNorthEast.Boundary.Center.Y-halfSize, halfSize, newNorthWest, newNorthWest.depth+1, qt.Checker, qt.maxDepth)
	newNorthEast.NorthWest = newQuadTree(newNorthEast.Boundary.Center.X-halfSize, newNorthEast.Boundary.Center.Y+halfSize, halfSize, newNorthWest, newNorthWest.depth+1, qt.Checker, qt.maxDepth)
	newNorthEast.NorthEast = newQuadTree(newNorthEast.Boundary.Center.X+halfSize, newNorthEast.Boundary.Center.Y+halfSize, halfSize, newNorthWest, newNorthWest.depth+1, qt.Checker, qt.maxDepth)
	qt.NorthEast = newNorthEast

	// South East (north west)
	newSouthEast := newQuadTree(qt.Boundary.Center.X+oldSize, qt.Boundary.Center.Y-oldSize, oldSize, qt, qt.depth+1, qt.Checker, qt.maxDepth)
	qt.SouthEast.addDepth()
	newSouthEast.NorthWest = qt.SouthEast
	newSouthEast.SouthEast = newQuadTree(newSouthEast.Boundary.Center.X+halfSize, newSouthEast.Boundary.Center.Y-halfSize, halfSize, newNorthWest, newNorthWest.depth+1, qt.Checker, qt.maxDepth)
	newSouthEast.SouthWest = newQuadTree(newSouthEast.Boundary.Center.X-halfSize, newSouthEast.Boundary.Center.Y-halfSize, halfSize, newNorthWest, newNorthWest.depth+1, qt.Checker, qt.maxDepth)
	newSouthEast.NorthEast = newQuadTree(newSouthEast.Boundary.Center.X+halfSize, newSouthEast.Boundary.Center.Y+halfSize, halfSize, newNorthWest, newNorthWest.depth+1, qt.Checker, qt.maxDepth)
	qt.SouthEast = newSouthEast

	// South West (north east)
	newSouthWest := newQuadTree(qt.Boundary.Center.X-oldSize, qt.Boundary.Center.Y-oldSize, oldSize, qt, qt.depth+1, qt.Checker, qt.maxDepth)
	qt.SouthWest.addDepth()
	newSouthWest.NorthEast = qt.SouthWest
	newSouthWest.SouthEast = newQuadTree(newSouthWest.Boundary.Center.X+halfSize, newSouthWest.Boundary.Center.Y-halfSize, halfSize, newNorthWest, newNorthWest.depth+1, qt.Checker, qt.maxDepth)
	newSouthWest.SouthWest = newQuadTree(newSouthWest.Boundary.Center.X-halfSize, newSouthWest.Boundary.Center.Y-halfSize, halfSize, newNorthWest, newNorthWest.depth+1, qt.Checker, qt.maxDepth)
	newSouthWest.NorthWest = newQuadTree(newSouthWest.Boundary.Center.X-halfSize, newSouthWest.Boundary.Center.Y+halfSize, halfSize, newNorthWest, newNorthWest.depth+1, qt.Checker, qt.maxDepth)
	qt.SouthWest = newSouthWest
}

func (qt *QuadTree[T]) addDepth() {
	qt.depth++
	if qt.NorthEast == nil {
		return
	}
	qt.NorthEast.addDepth()
	qt.NorthWest.addDepth()
	qt.SouthWest.addDepth()
	qt.SouthEast.addDepth()
}

func (qt *QuadTree[T]) Insert(item BoundedItem) bool {
	// First, verify that the item belongs in this tree.
	if !qt.Checker.Check(qt.Boundary, item.GetBounds()) {
		if qt.depth == 0 && qt.mayExpand {
			qt.Expand()
			return qt.Insert(item)
		} else {
			return false
		}
	}

	// If max depth has been reached, just put it in!
	if qt.depth == qt.maxDepth {
		qt.Items = append(qt.Items, item)
		return true
	}

	// No sub quad trees and more room, just plug it in
	if qt.NorthEast == nil {
		if len(qt.Items) < NodeCapacity {
			qt.Items = append(qt.Items, item)
			return true
		}

		// Obviously there wasn't any room
		qt.Subdivide()
	}

	// Figure out which children the item can be in.
	var insertInto *QuadTree[T]
	multiple := false
	if qt.NorthEast.ContainedIn(item) {
		insertInto = qt.NorthEast
	}
	if qt.SouthEast.ContainedIn(item) {
		if insertInto != nil {
			multiple = true
		}
		insertInto = qt.SouthEast
	}
	if qt.SouthWest.ContainedIn(item) {
		if insertInto != nil {
			multiple = true
		}
		insertInto = qt.SouthWest
	}
	if qt.NorthWest.ContainedIn(item) {
		if insertInto != nil {
			multiple = true
		}
		insertInto = qt.NorthWest
	}

	if multiple {
		// The item can fit in multiple leafs,
		// put it into the current sub tree.
		qt.Items = append(qt.Items, item)
		return true
	} else {
		if insertInto != nil {
			return insertInto.Insert(item)
		}

		return false
	}
}

func (qt *QuadTree[T]) Clear() {
	qt.NorthEast = nil
	qt.NorthWest = nil
	qt.SouthEast = nil
	qt.SouthWest = nil

	qt.Items = nil
}

func (qt *QuadTree[T]) ContainedIn(item BoundedItem) bool {
	//return intersection.IntersectingSquaredAABB(qt.Boundary, item.GetBounds())
	return qt.Checker.Check(qt.Boundary, item.GetBounds())
}

type QueryResultWrapper struct {
	Items []BoundedItem
}

func NewQueryResultWrapper(items []BoundedItem) *QueryResultWrapper {
	return &QueryResultWrapper{items}
}

// RemoveDuplicates will remove duplicates.
// Remember to use the wrapped result.
func (w *QueryResultWrapper) RemoveDuplicates() {
	u := map[BoundedItem]struct{}{}
	for _, i := range w.Items {
		u[i] = struct{}{}
	}

	i := 0
	for ui, _ := range u {
		w.Items[i] = ui
		i++
	}
	w.Items = w.Items[:i]
}

func (qt *QuadTree[T]) Query(contained BoundedItem) []BoundedItem {
	return qt.Query2(contained, false)
}

func (qt *QuadTree[T]) Query2(contained BoundedItem, testMutual bool) []BoundedItem {
	//if !intersection.IntersectingSquaredAABB(qt.Boundary, contained.GetBounds()) {
	if !qt.Checker.Check(qt.Boundary, contained.GetBounds()) {
		return nil
	}
	// Leaf node
	points := []BoundedItem{}
	for _, p := range qt.Items {
		//if intersection.IntersectingSquaredAABB(qt.Boundary, contained.GetBounds()) {
		if qt.Checker.Check(qt.Boundary, contained.GetBounds()) {
			if testMutual {
				//if intersection.Intersecting2D[T](p.GetBounds(), contained.GetBounds()) {
				if qt.Checker.Check(p.GetBounds(), contained.GetBounds()) {
					points = append(points, p)
				}
			} else {
				points = append(points, p)
			}
		}
	}

	if qt.NorthEast == nil {
		return points
	}

	points = append(points, qt.NorthEast.Query2(contained, testMutual)...)
	points = append(points, qt.SouthEast.Query2(contained, testMutual)...)
	points = append(points, qt.SouthWest.Query2(contained, testMutual)...)
	points = append(points, qt.NorthWest.Query2(contained, testMutual)...)

	return points
}

func (qt *QuadTree[T]) QueryGeometry(contained interface{}) []BoundedItem {
	return qt.QueryGeometry2(contained, false)
}

func (qt *QuadTree[T]) QueryGeometry2(contained interface{}, testMutual bool) []BoundedItem {
	//if !intersection.IntersectingSquaredAABB(qt.Boundary, contained) {
	if !qt.Checker.Check(qt.Boundary, contained) {
		return nil
	}
	// Leaf node
	points := []BoundedItem{}
	for _, p := range qt.Items {
		//if intersection.IntersectingSquaredAABB(qt.Boundary, contained) {
		if qt.Checker.Check(qt.Boundary, contained) {
			if testMutual {
				//if intersection.Intersecting2D[T](p.GetBounds(), contained) {
				if qt.Checker.Check(qt.Boundary, contained) {
					points = append(points, p)
				}
			} else {
				points = append(points, p)
			}
		}
	}

	if qt.NorthEast == nil {
		return points
		return points
	}

	points = append(points, qt.NorthEast.QueryGeometry2(contained, testMutual)...)
	points = append(points, qt.SouthEast.QueryGeometry2(contained, testMutual)...)
	points = append(points, qt.SouthWest.QueryGeometry2(contained, testMutual)...)
	points = append(points, qt.NorthWest.QueryGeometry2(contained, testMutual)...)

	return points
}

func (qt *QuadTree[T]) QueryBoundedWithDrain(contained BoundedItem, drain chan BoundedItem) {
	qt.QueryWithDrain2(contained.GetBounds(), false, drain)
}

func (qt *QuadTree[T]) QueryWithDrain(contained interface{}, drain chan BoundedItem) {
	qt.QueryWithDrain2(contained, false, drain)
	return
}

func (qt *QuadTree[T]) QueryWithDrainBoth(contained interface{}, drain chan BoundedItem) {
	qt.QueryWithDrain2(contained, true, drain)
	return
}

func (qt *QuadTree[T]) QueryBoundedWithDrain2(contained BoundedItem, testMutual bool, drain chan BoundedItem) {
	qt.QueryWithDrain2(contained.GetBounds(), testMutual, drain)
}

func (qt *QuadTree[T]) QueryWithDrain2(contained interface{}, testMutual bool, drain chan BoundedItem) {
	//if !intersection.IntersectingSquaredAABB(qt.Boundary, contained) {
	if !qt.Checker.Check(qt.Boundary, contained) {
		return
	}
	// Leaf node
	for _, p := range qt.Items {
		//if intersection.IntersectingSquaredAABB(qt.Boundary, contained) {
		if qt.Checker.Check(qt.Boundary, contained) {
			if testMutual {
				//if intersection.Intersecting2D[T](p.GetBounds(), contained) {
				if qt.Checker.Check(qt.Boundary, contained) {
					drain <- p
				}
			} else {
				drain <- p
			}
		}
	}

	if qt.NorthEast == nil {
		return
	}

	qt.NorthEast.QueryWithDrain(contained, drain)
	qt.SouthEast.QueryWithDrain(contained, drain)
	qt.SouthWest.QueryWithDrain(contained, drain)
	qt.NorthWest.QueryWithDrain(contained, drain)

	return
}

func (qt *QuadTree[T]) Clone() *QuadTree[T] {
	newQT := NewQuadTree(qt.Boundary.Center.X, qt.Boundary.Center.Y, qt.Boundary.HalfSize)
	if qt.NorthEast != nil {
		newQT.NorthEast = qt.NorthEast.Clone()
		newQT.SouthEast = qt.SouthEast.Clone()
		newQT.SouthWest = qt.SouthWest.Clone()
		newQT.NorthWest = qt.NorthWest.Clone()
	} else {
		newQT.Items = make([]BoundedItem, len(qt.Items))
		for i, item := range qt.Items {
			newQT.Items[i] = item
		}
	}
	return newQT
}

func (qt *QuadTree[T]) Stringify() {
	qt.stringify(0)
}

func (qt *QuadTree[T]) stringify(lvl int) {
	indent := fmt.Sprintf("%s", strings.Repeat(" ", lvl*2))
	//fmt.Printf("%sBounds: %v, depth: %d, points: %v\n", indent, qt.Boundary, qt.depth, qt.Items)
	fmt.Printf("Bounds: %v, depth: %d, points: %v\n", qt.Boundary, qt.depth, qt.Items)
	if qt.NorthEast != nil {
		fmt.Printf("%s* North east: ", indent)
		qt.NorthEast.stringify(lvl + 1)
		fmt.Printf("%s* South east: ", indent)
		qt.SouthEast.stringify(lvl + 1)
		fmt.Printf("%s* South west: ", indent)
		qt.SouthWest.stringify(lvl + 1)
		fmt.Printf("%s* North west: ", indent)
		qt.NorthWest.stringify(lvl + 1)
	} else {
		//fmt.Printf("%sNo sub division\n", indent)
	}
}

func (qt *QuadTree[T]) Visit(cb func(tree *QuadTree[T])) {
	cb(qt)
	if qt.NorthEast == nil {
		return
	}

	qt.NorthEast.Visit(cb)
	qt.SouthEast.Visit(cb)
	qt.SouthWest.Visit(cb)
	qt.NorthWest.Visit(cb)
}
