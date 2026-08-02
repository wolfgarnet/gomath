package gomath

import (
	"math"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRay3NormalisesItsDirection(t *testing.T) {
	ray := NewRay3(Vector3[float64]{X: 1, Y: 2, Z: 3}, Vector3[float64]{X: 0, Y: 0, Z: -7})

	assert.InDelta(t, 1, ray.Direction.Length(), 1e-12,
		"the direction must be unit, or every distance it returns is scaled by its length")
	assert.InDelta(t, -1, ray.Direction.Z, 1e-12)

	// The origin is untouched.
	assert.Equal(t, Vector3[float64]{X: 1, Y: 2, Z: 3}, ray.Origin)
}

// A zero direction defines no ray. It must come back as zero rather than as
// NaNs, since NaNs would propagate silently through every intersection instead
// of being rejected.
func TestZeroDirectionRayIsInvalidRatherThanNaN(t *testing.T) {
	ray := NewRay3(Vector3[float64]{X: 5}, Vector3[float64]{})

	assert.False(t, ray.IsValid())
	assert.True(t, ray.Direction.IsZero())
	assert.Equal(t, ray.Direction.X, ray.Direction.X, "must not be NaN")

	// The same for a ray between coincident points, which is how this arises in
	// practice: asking for the ray from a point towards itself.
	same := NewRay3Between(Vector3[float64]{X: 2}, Vector3[float64]{X: 2})
	assert.False(t, same.IsValid())
}

func TestAtWalksTrueDistances(t *testing.T) {
	ray := NewRay3(Vector3[float64]{Y: 10}, Vector3[float64]{X: 3, Y: 0, Z: 0})

	assert.Equal(t, ray.Origin, ray.At(0), "At(0) is the origin")

	// A unit direction means the parameter is a distance in world units, which
	// is what lets hits on different shapes be compared directly.
	assert.InDelta(t, 4, ray.At(4).X, 1e-12)
	assert.InDelta(t, 10, ray.At(4).Y, 1e-12)
	assert.InDelta(t, 4, ray.At(4).Sub(ray.Origin).Length(), 1e-12)
}

func TestNewRay3BetweenPointsAtItsTarget(t *testing.T) {
	from := Vector3[float64]{X: 1, Y: 2, Z: 3}
	to := Vector3[float64]{X: 1, Y: 8, Z: 3}

	ray := NewRay3Between(from, to)

	assert.Equal(t, from, ray.Origin)
	assert.InDelta(t, 1, ray.Direction.Y, 1e-12)
	// And walking the distance between them arrives exactly.
	arrival := ray.At(to.Sub(from).Length())
	assert.InDelta(t, to.Y, arrival.Y, 1e-12)
}

// A ray is a half-line, so the nearest point to something behind it is its
// ORIGIN — not the projection onto the infinite line, which would lie off the
// ray entirely and understate the distance.
func TestClosestPointClampsBehindTheOrigin(t *testing.T) {
	ray := NewRay3(Vector3[float64]{}, Vector3[float64]{Z: -1})

	ahead := Vector3[float64]{X: 3, Y: 0, Z: -5}
	assert.InDelta(t, -5, ray.ClosestPoint(ahead).Z, 1e-12, "a point ahead projects onto the ray")
	assert.InDelta(t, 3, ray.DistanceTo(ahead), 1e-12)

	behind := Vector3[float64]{X: 3, Y: 0, Z: 5}
	assert.Equal(t, ray.Origin, ray.ClosestPoint(behind),
		"a point behind the origin is nearest to the origin itself")

	// Distance to the point behind is the full distance to the origin, not the
	// perpendicular 3 the infinite line would give.
	assert.InDelta(t, behind.Length(), ray.DistanceTo(behind), 1e-12)
	assert.InDelta(t, behind.LengthSq(), ray.DistanceToSq(behind), 1e-9)
}

func TestFlippedReversesWithoutMoving(t *testing.T) {
	ray := NewRay3(Vector3[float64]{X: 1, Y: 2, Z: 3}, Vector3[float64]{Z: -1})
	flipped := ray.Flipped()

	assert.Equal(t, ray.Origin, flipped.Origin, "flipping must not move the origin")
	assert.InDelta(t, 1, flipped.Direction.Z, 1e-12)
	assert.Equal(t, ray, flipped.Flipped(), "flipping twice is the identity")
}

// The origin is a position and the direction is a direction, so a translation
// must move one and not the other. Getting this wrong sends every transformed
// ray off towards the matrix's translation.
func TestTransformedMovesTheOriginButNotTheDirection(t *testing.T) {
	ray := NewRay3(Vector3[float64]{}, Vector3[float64]{Z: -1})

	moved := ray.Transformed(TranslationMatrix4[float64](10, 20, 30))

	assert.InDelta(t, 10, moved.Origin.X, 1e-12)
	assert.InDelta(t, 20, moved.Origin.Y, 1e-12)
	assert.InDelta(t, -1, moved.Direction.Z, 1e-12,
		"a translation must not move a direction")
}

func TestTransformedRotatesBoth(t *testing.T) {
	ray := NewRay3(Vector3[float64]{X: 1}, Vector3[float64]{X: 1})

	// A quarter turn about Y takes +X to -Z.
	rotation := QuaternionFromAxisAngle(Vector3[float64]{Y: 1}, math.Pi/2).ToMatrix4()
	turned := ray.Transformed(rotation)

	assert.InDelta(t, -1, turned.Origin.Z, 1e-9, "the origin rotates")
	assert.InDelta(t, -1, turned.Direction.Z, 1e-9, "and so does the direction")
}

// A scaling matrix must still yield a unit ray, or distances in the target space
// would be scaled by however much the matrix stretched.
func TestTransformedRenormalisesUnderScale(t *testing.T) {
	ray := NewRay3(Vector3[float64]{}, Vector3[float64]{X: 1})

	scaled := ray.Transformed(ScaleMatrix4[float64](5, 5, 5))

	require.True(t, scaled.IsValid())
	assert.InDelta(t, 1, scaled.Direction.Length(), 1e-12,
		"a scaled ray must stay unit, so its distances stay true distances")
	assert.InDelta(t, 1, scaled.Direction.X, 1e-12)
}

func TestRay3String(t *testing.T) {
	ray := NewRay3(Vector3[float64]{}, Vector3[float64]{Z: -1})
	assert.Contains(t, ray.String(), "Ray3")
}
