package spatial

import (
	"github.com/wolfgarnet/gomath"
)

func ForeachElement(tree *Quadtree[float32], position gomath.Vector2[float32], halfSize float32, f func(element BoundedItem)) {
	aabb := gomath.NewSquaredAABB2D[float32](position, halfSize)
	drain := make(chan BoundedItem, 16)
	go func() {
		tree.QueryWithDrain(aabb, drain)
		close(drain)
	}()

	for item := range drain {
		f(item)
	}
}
