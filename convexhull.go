package gomath

import (
	"math"
	"sort"
)

type convexHull[T Float] []Vector2[T]

func (a convexHull[T]) Len() int      { return len(a) }
func (a convexHull[T]) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a convexHull[T]) Less(i, j int) bool {
	if i == 0 {
		return false
	}
	if j == 0 {
		return false
	}

	return CCW[T](a[0], a[i], a[j]) > 0
}

func ConvexHullFromPolygons[T Float](polygons []Polygon[T]) []Vector2[T] {
	points := make([]Vector2[T], 0)
	for _, p := range polygons {
		points = append(points, p.Points...)
	}

	return ConvexHull(points)
}

// https://en.wikipedia.org/wiki/Graham_scan
func ConvexHull[T Float](ps []Vector2[T]) []Vector2[T] {
	// Find point with lowest y value and put it first
	var by = T(math.MaxFloat32)
	var bx = T(-math.MaxFloat32)
	for i, p := range ps {
		if p.Y < by || (p.Y == by && p.X > bx) {
			by = p.Y
			bx = p.X
			ps[0], ps[i] = ps[i], ps[0]
		}
	}

	sort.Sort(convexHull[T](ps))

	convexHull := []Vector2[T]{ps[0], ps[1], ps[2]}
	for i := 3; i < len(ps); i++ {
		p := ps[i]
		for {
			l := len(convexHull)
			if l < 3 {

			}
			top0 := convexHull[l-1]
			top1 := convexHull[l-2]

			if CCW(top1, top0, p) > 0 {
				convexHull = append(convexHull, p)
				break
			}

			convexHull = convexHull[:l-1]
		}

	}

	return convexHull
}

/*
for(auto v : polygon)
{
if(isLeft(p, left, v) > 0.0)
{
if(abs(isLeft(p, right, v)) > Epsilon)
left = p;
else { handle special case where P is parallel,}
}
if(isLeft(p, right, v) < 0.0)
{
if(abs(isLeft(p, right, v)) > Epsilon)
right = p;
else { handle special case where P is parallel,}
}
*/
func ExtremePoints[T Float](p Vector2[T], polygon []Vector2[T]) (left Vector2[T], right Vector2[T]) {
	if len(polygon) == 0 {
		return
	}
	left, right = polygon[0], polygon[0]
	for _, v := range polygon {
		if IsLeft(p, left, v) > 0 {
			if Abs(IsLeft(p, right, v)) > Epsilon {
				left = v
			} else {
				// handle special case where P is parallel
			}
		}
		if IsLeft(p, right, v) < 0 {
			if Abs(IsLeft(p, right, v)) > Epsilon {
				right = v
			} else {
				// handle special case where P is parallel
			}
		}
	}

	return left, right
}
