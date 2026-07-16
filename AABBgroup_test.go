package gomath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAABBGroup_Add(t *testing.T) {
	group := &AABBGroup[float64]{}
	a := &AABB[float64]{Center: Vector2[float64]{1, 1}, HalfWidth: 1, HalfHeight: 1}
	b := &AABB[float64]{Center: Vector2[float64]{5, 5}, HalfWidth: 2, HalfHeight: 2}

	assert.True(t, group.Add(a))
	assert.True(t, group.Add(b))

	assert.Equal(t, []*AABB[float64]{a, b}, group.AABBs)
}

// TestAABBGroups_Add pins that adding through the outer AABBGroups container
// reaches the groups it holds, rather than mutating copies of them.
func TestAABBGroups_Add(t *testing.T) {
	groups := &AABBGroups[float64]{
		Groups: []AABBGroup[float64]{{}, {}},
	}
	a := &AABB[float64]{Center: Vector2[float64]{1, 1}, HalfWidth: 1, HalfHeight: 1}

	groups.Add(a)

	for i, group := range groups.Groups {
		assert.Lenf(t, group.AABBs, 1, "group %d should have received the AABB", i)
		if len(group.AABBs) == 1 {
			assert.Equal(t, a, group.AABBs[0])
		}
	}
}

func TestAABBGroups_AddEmpty(t *testing.T) {
	groups := &AABBGroups[float64]{}
	a := &AABB[float64]{HalfWidth: 1, HalfHeight: 1}

	assert.NotPanics(t, func() { groups.Add(a) })
}
