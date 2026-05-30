package spatial

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
)

// TestBoxedQuadtreeInsertCases pins the (bool, error) contract introduced
// when the silent-`false` return became a real error path. The motivating
// problem (spatial-coverage spec §6.2 in pixelmaster) was that callers
// silently discarded the bool — the error-typed return is enforced by
// errcheck/govet so a future caller can't reintroduce the silent drop.
func TestBoxedQuadtreeInsertCases(t *testing.T) {
	makeItem := func(id uint64, cx, cy, halfW, halfH float32) BoxedItem[float32, uint64] {
		return BoxedItem[float32, uint64]{
			ItemID: id,
			AABB:   gomath.AABB[float32]{Center: gomath.Vector2[float32]{X: cx, Y: cy}, HalfWidth: halfW, HalfHeight: halfH},
		}
	}

	t.Run("item inside boundary is accepted", func(t *testing.T) {
		tree := NewBoxedQuadtree[float32, uint64](0, 0, 100)
		ok, err := tree.Insert(makeItem(1, 10, 10, 1, 1))

		assert.True(t, ok)
		assert.NoError(t, err)
	})

	t.Run("item entirely outside boundary returns ErrItemOutsideBoundary", func(t *testing.T) {
		tree := NewBoxedQuadtree[float32, uint64](0, 0, 100)
		ok, err := tree.Insert(makeItem(1, 500, 500, 1, 1))

		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrItemOutsideBoundary, "callers should be able to errors.Is the sentinel")

		// The wrapped type carries the geometry so log lines don't have to
		// re-fetch the boundary or item AABB.
		var detailed *InsertOutsideBoundaryError[float32]
		assert.True(t, errors.As(err, &detailed), "error should unwrap to InsertOutsideBoundaryError for geometry context")
		assert.Equal(t, float32(0), detailed.Boundary.Center.X)
		assert.Equal(t, float32(0), detailed.Boundary.Center.Y)
		assert.Equal(t, float32(100), detailed.Boundary.HalfSize)
		assert.Equal(t, float32(500), detailed.ItemAABB.Center.X)
	})

	t.Run("item straddling boundary edge is accepted", func(t *testing.T) {
		// The item's center is past the boundary but its half-extent reaches
		// back inside — SquaredAABBAABB treats this as overlap, so Insert
		// must accept it. Pinning the contract here so a future tightening
		// of the boundary check doesn't silently change behaviour.
		tree := NewBoxedQuadtree[float32, uint64](0, 0, 100)
		ok, err := tree.Insert(makeItem(1, 99, 0, 5, 5))

		assert.True(t, ok)
		assert.NoError(t, err)
	})
}
