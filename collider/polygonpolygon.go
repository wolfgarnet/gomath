package collider

import (
	"github.com/wolfgarnet/gomath"
	"math"
)

func PolygonPolygon[T gomath.Float](poly1, poly2 gomath.Polygon[T]) (penetration T, surfaceNormal gomath.Vector2[T], overlapping bool) {
	var (
		p1 gomath.Vector2[T]
	)
	penetration = math.MaxFloat32
	l1 := len(poly1.Points)
	for i, p0 := range poly1.Points {
		if i < l1-1 {
			p1 = poly1.Points[i+1]
		} else {
			p1 = poly1.Points[0]
		}
		vector1 := p0.Sub(p1)
		axis := gomath.GetNormal(vector1)
		if hit := overlapOnAxisPolygonPolygon(poly1, poly2, axis, &penetration, &surfaceNormal); !hit {
			return 0, surfaceNormal, false
		}
	}

	l2 := len(poly2.Points)
	for i, p1_0 := range poly2.Points {
		if i < l2-1 {
			p1 = poly2.Points[i+1]
		} else {
			p1 = poly2.Points[0]
		}

		vector1_1 := p1_0.Sub(p1)

		axis := gomath.GetNormal(vector1_1)
		if hit := overlapOnAxisPolygonPolygon(poly1, poly2, axis, &penetration, &surfaceNormal); !hit {
			return 0, surfaceNormal, false
		}
	}

	return penetration, surfaceNormal, true
}

func overlapOnAxisPolygonPolygon[T gomath.Float](poly1, poly2 gomath.Polygon[T], axis gomath.Vector2[T], penetration2 *T, normal *gomath.Vector2[T]) bool {
	if axis.LengthSq() < 0.0001 {
		return false
	}
	p1min, p1max, _, _ := gomath.TransformToAxis(poly1, axis)
	p2min, p2max, _, _ := gomath.TransformToAxis(poly2, axis)

	p1project := (p1max - p1min) / 2
	p2project := (p2max - p2min) / 2
	p1center := p1min + p1project
	p2center := p2min + p2project
	distance := p1center - p2center
	penetration := (p1project + p2project) - gomath.Abs(distance)

	if penetration < *penetration2 && penetration >= 0 {
		*penetration2 = penetration
		*normal = axis
	}

	return penetration >= 0
}
