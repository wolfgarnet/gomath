package gomath

type AABBGroups[T Number] struct {
	Groups []AABBGroup[T]
}

func (g *AABBGroups[T]) Add(aabb *AABB[T]) {
	for i := range g.Groups {
		g.Groups[i].Add(aabb)
	}
}

type AABBGroup[T Number] struct {
	AABBs     []*AABB[T]
	Container AABB[T]
}

func (g *AABBGroup[T]) Add(aabb *AABB[T]) bool {
	//for _, g := range g.AABBs {
	// TODO
	//}

	g.AABBs = append(g.AABBs, aabb)
	return true
}
