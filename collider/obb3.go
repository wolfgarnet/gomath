package collider

import "github.com/wolfgarnet/gomath"

// Oriented-box overlap by the separating-axis theorem.
//
// Two convex boxes are disjoint if and only if some axis exists on which their
// projections do not overlap, and for a pair of boxes it suffices to test
// fifteen candidates: the three face normals of each box, and the nine
// cross-products of their edge directions.
//
// # Why this form and not the obvious one
//
// The direct transcription builds those nine cross-products as vectors and
// projects onto each. It has two problems, and the second is not obvious:
//
//  1. It is slow. Nine cross-products plus thirty projections, each of which is
//     three dot products.
//
//  2. It is WRONG for near-parallel boxes. When two edge directions are nearly
//     parallel their cross-product is nearly zero, dominated by cancellation
//     error, so the projections and the centre distance are all tiny numbers
//     with enormous relative error — and the test reports separation for boxes
//     that clearly overlap.
//
// Exactly parallel is worse in a subtle way. The cross-product is then exactly
// zero, so the projections and the distance are all zero, and whether the axis
// separates depends on whether the comparison is written `distance > sum` or
// `distance >= sum`. With `>=` a pair of parallel overlapping boxes is reported
// as separated. That failure mode is invisible until something rotates into
// alignment.
//
// The form below is Ericson's (Real-Time Collision Detection, 4.4.1). It
// precomputes the rotation of b in a's frame and works entirely in scalars, so
// there are no cross-products to lose precision in, and the epsilon added to
// absRotation absorbs the near-parallel case where it arises. It is both faster
// and robust, which is the unusual case where those do not trade off.
//
// The epsilon makes the test very slightly conservative: boxes separated by less
// than about a millionth of their extent may report as touching. For collision
// response that is the right direction to err in.
const parallelEpsilon = 1e-6

// OBB3OBB3 reports whether two oriented boxes overlap.
//
// It early-outs on the first separating axis found, and orders the tests face
// normals first, since those separate most pairs in practice.
func OBB3OBB3[T gomath.Float](a gomath.OBB3[T], b gomath.OBB3[T]) bool {
	var rotation, absRotation [3][3]T

	for i := 0; i < 3; i++ {
		axis := a.Axis(i)
		for j := 0; j < 3; j++ {
			rotation[i][j] = axis.Dot(b.Axis(j))
			absRotation[i][j] = gomath.Abs(rotation[i][j]) + T(parallelEpsilon)
		}
	}

	// The vector between centres, expressed in a's frame.
	offset := b.Center.Sub(a.Center)
	translation := [3]T{
		offset.Dot(a.Axis(0)),
		offset.Dot(a.Axis(1)),
		offset.Dot(a.Axis(2)),
	}

	extentA := [3]T{a.HalfSize.X, a.HalfSize.Y, a.HalfSize.Z}
	extentB := [3]T{b.HalfSize.X, b.HalfSize.Y, b.HalfSize.Z}

	// Face normals of a.
	for i := 0; i < 3; i++ {
		radiusA := extentA[i]
		radiusB := extentB[0]*absRotation[i][0] +
			extentB[1]*absRotation[i][1] +
			extentB[2]*absRotation[i][2]

		if gomath.Abs(translation[i]) > radiusA+radiusB {
			return false
		}
	}

	// Face normals of b.
	for j := 0; j < 3; j++ {
		radiusA := extentA[0]*absRotation[0][j] +
			extentA[1]*absRotation[1][j] +
			extentA[2]*absRotation[2][j]
		radiusB := extentB[j]

		distance := translation[0]*rotation[0][j] +
			translation[1]*rotation[1][j] +
			translation[2]*rotation[2][j]

		if gomath.Abs(distance) > radiusA+radiusB {
			return false
		}
	}

	// The nine edge-edge axes, written out because each has a different pair of
	// contributing extents. The indices follow from expanding the cross product
	// symbolically, which is what removes the need to compute it.
	edgeCases := [9]struct {
		radiusA  T
		radiusB  T
		distance T
	}{
		// A0 x B0, A0 x B1, A0 x B2
		{
			extentA[1]*absRotation[2][0] + extentA[2]*absRotation[1][0],
			extentB[1]*absRotation[0][2] + extentB[2]*absRotation[0][1],
			translation[2]*rotation[1][0] - translation[1]*rotation[2][0],
		},
		{
			extentA[1]*absRotation[2][1] + extentA[2]*absRotation[1][1],
			extentB[0]*absRotation[0][2] + extentB[2]*absRotation[0][0],
			translation[2]*rotation[1][1] - translation[1]*rotation[2][1],
		},
		{
			extentA[1]*absRotation[2][2] + extentA[2]*absRotation[1][2],
			extentB[0]*absRotation[0][1] + extentB[1]*absRotation[0][0],
			translation[2]*rotation[1][2] - translation[1]*rotation[2][2],
		},
		// A1 x B0, A1 x B1, A1 x B2
		{
			extentA[0]*absRotation[2][0] + extentA[2]*absRotation[0][0],
			extentB[1]*absRotation[1][2] + extentB[2]*absRotation[1][1],
			translation[0]*rotation[2][0] - translation[2]*rotation[0][0],
		},
		{
			extentA[0]*absRotation[2][1] + extentA[2]*absRotation[0][1],
			extentB[0]*absRotation[1][2] + extentB[2]*absRotation[1][0],
			translation[0]*rotation[2][1] - translation[2]*rotation[0][1],
		},
		{
			extentA[0]*absRotation[2][2] + extentA[2]*absRotation[0][2],
			extentB[0]*absRotation[1][1] + extentB[1]*absRotation[1][0],
			translation[0]*rotation[2][2] - translation[2]*rotation[0][2],
		},
		// A2 x B0, A2 x B1, A2 x B2
		{
			extentA[0]*absRotation[1][0] + extentA[1]*absRotation[0][0],
			extentB[1]*absRotation[2][2] + extentB[2]*absRotation[2][1],
			translation[1]*rotation[0][0] - translation[0]*rotation[1][0],
		},
		{
			extentA[0]*absRotation[1][1] + extentA[1]*absRotation[0][1],
			extentB[0]*absRotation[2][2] + extentB[2]*absRotation[2][0],
			translation[1]*rotation[0][1] - translation[0]*rotation[1][1],
		},
		{
			extentA[0]*absRotation[1][2] + extentA[1]*absRotation[0][2],
			extentB[0]*absRotation[2][1] + extentB[1]*absRotation[2][0],
			translation[1]*rotation[0][2] - translation[0]*rotation[1][2],
		},
	}

	for _, edge := range edgeCases {
		if gomath.Abs(edge.distance) > edge.radiusA+edge.radiusB {
			return false
		}
	}

	return true
}

// faceAxisPreference biases the minimum-overlap search towards face normals.
//
// Edge-edge axes are derived from a cross product, so their direction is far
// less stable than a face normal's when two boxes are nearly aligned. Two axes
// whose overlaps are within a hair of each other will otherwise alternate
// between frames, and a contact normal that flips is what makes a resting box
// jitter. Requiring an edge axis to beat the best face axis by this factor
// keeps the normal steady.
const faceAxisPreference = 0.95

// OBB3OBB3Penetration returns the minimum translation that separates two
// overlapping oriented boxes.
//
// The normal points from a towards b, matching every other penetration function
// here, and the penetration is the distance along it.
//
// Unlike OBB3OBB3 this examines all fifteen axes — a minimum cannot early-out —
// and normalises each one, because the depth has to come back in world units
// rather than in units of an unnormalised cross product. Degenerate axes from
// parallel edges are skipped rather than fudged with an epsilon: an axis with no
// direction cannot be a separating axis, and it certainly cannot be the answer.
//
// That is why this is a separate function from the boolean test rather than the
// only one. A broadphase asks "do these touch?" of many pairs and "by how much?"
// of the few that do, and the cheap answer should stay cheap.
func OBB3OBB3Penetration[T gomath.Float](a gomath.OBB3[T], b gomath.OBB3[T]) (penetration T, normal gomath.Vector3[T], overlapping bool) {
	offset := b.Center.Sub(a.Center)

	var (
		bestOverlap T
		bestAxis    gomath.Vector3[T]
		foundAny    bool
	)

	consider := func(axis gomath.Vector3[T], isFace bool) bool {
		lengthSq := axis.LengthSq()
		if lengthSq < T(gomath.Epsilon)*T(gomath.Epsilon) {
			// A parallel edge pair: no direction, so nothing to test.
			return true
		}

		unit := axis.DivScalar(gomath.Sqrt(lengthSq))

		radiusA := gomath.Abs(unit.Dot(a.Axis(0)))*a.HalfSize.X +
			gomath.Abs(unit.Dot(a.Axis(1)))*a.HalfSize.Y +
			gomath.Abs(unit.Dot(a.Axis(2)))*a.HalfSize.Z
		radiusB := gomath.Abs(unit.Dot(b.Axis(0)))*b.HalfSize.X +
			gomath.Abs(unit.Dot(b.Axis(1)))*b.HalfSize.Y +
			gomath.Abs(unit.Dot(b.Axis(2)))*b.HalfSize.Z

		overlap := radiusA + radiusB - gomath.Abs(unit.Dot(offset))
		if overlap < 0 {
			return false
		}

		// Face axes are all considered before any edge axis, so an edge only
		// displaces the incumbent by beating it by the preference margin.
		better := !foundAny || overlap < bestOverlap
		if !isFace {
			better = !foundAny || overlap < bestOverlap*T(faceAxisPreference)
		}

		if better {
			bestOverlap = overlap
			bestAxis = unit
			foundAny = true
		}

		return true
	}

	// Face normals first, so a face axis is established before any edge axis is
	// allowed to displace it.
	for i := 0; i < 3; i++ {
		if !consider(a.Axis(i), true) {
			return 0, normal, false
		}
	}
	for j := 0; j < 3; j++ {
		if !consider(b.Axis(j), true) {
			return 0, normal, false
		}
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if !consider(a.Axis(i).Cross(b.Axis(j)), false) {
				return 0, normal, false
			}
		}
	}

	if !foundAny {
		// Every axis was degenerate, which cannot happen for boxes with an
		// orthonormal basis. Report no overlap rather than a meaningless normal.
		return 0, normal, false
	}

	// Point the normal from a towards b.
	if bestAxis.Dot(offset) < 0 {
		bestAxis = bestAxis.Negate()
	}

	return bestOverlap, bestAxis, true
}
