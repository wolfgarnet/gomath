package gomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEllipsoid_Contains(t *testing.T) {
	e := NewEllipsoid(Vector2[int]{0, 0}, Vector2[int]{4, 2})
	assert.True(t, e.Contains(Vector2[int]{0, 0}))
	assert.True(t, e.Contains(Vector2[int]{4, 0})) // on the boundary (x apex)
	assert.True(t, e.Contains(Vector2[int]{0, 2})) // on the boundary (y apex)
	assert.True(t, e.Contains(Vector2[int]{3, 0}))
	assert.False(t, e.Contains(Vector2[int]{4, 2})) // outside
	assert.False(t, e.Contains(Vector2[int]{5, 0}))
}

func TestEllipsoid_BoundingBoxIsCircleEqualsTranslate(t *testing.T) {
	e := NewEllipsoid(Vector2[int]{1, 2}, Vector2[int]{4, 2})
	bb := e.BoundingBox()
	assert.Equal(t, Vector2[int]{-3, 0}, bb.Min())
	assert.Equal(t, Vector2[int]{5, 4}, bb.Max())

	assert.False(t, e.IsCircle())
	assert.True(t, NewEllipsoid(Vector2[int]{0, 0}, Vector2[int]{3, 3}).IsCircle())

	assert.True(t, e.Equals(NewEllipsoid(Vector2[int]{1, 2}, Vector2[int]{4, 2})))
	assert.Equal(t, NewEllipsoid(Vector2[int]{4, 6}, Vector2[int]{4, 2}), e.Translate(Vector2[int]{3, 4}))
}

func TestEllipsoid_AreaPointAt(t *testing.T) {
	e := NewEllipsoid(Vector2[float64]{0, 0}, Vector2[float64]{4, 2})
	assert.InDelta(t, math.Pi*8, e.Area(), 1e-9)

	p0 := e.PointAt(0)
	assert.InDelta(t, 4, p0.X, 1e-9)
	assert.InDelta(t, 0, p0.Y, 1e-9)

	p90 := e.PointAt(math.Pi / 2)
	assert.InDelta(t, 0, p90.X, 1e-9)
	assert.InDelta(t, 2, p90.Y, 1e-9)
}
