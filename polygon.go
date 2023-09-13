package gomath

import (
	"math"
)

type Polygon[T Number] struct {
	Points []Vector2[T]
	center *Vector2[T]
}

func NewPolygon[T Number](points []Vector2[T]) Polygon[T] {
	polygon := Polygon[T]{
		Points: points,
	}
	polygon.CCWPoints()
	return polygon
}

// Edge returns the edge of the polygon with the given index.
func (p Polygon[T]) Edge(index int) (a, b Vector2[T]) {
	l := len(p.Points)
	if index >= l-1 {
		return p.Points[l-1], p.Points[0]
	}

	return p.Points[index], p.Points[index+1]
}

func (p Polygon[T]) Edges() int {
	return len(p.Points)
}

// Next returns the next counter clock wise vertex of the polygon and its index.
func (p Polygon[T]) Next(i int) (Vector2[T], int) {
	i = (i + 1) % len(p.Points)
	return p.Points[i], i
}

// Previous returns the previous, clock wise vertex and its index.
func (p Polygon[T]) Previous(i int) (Vector2[T], int) {
	if i == 0 {
		i = len(p.Points) - 1
		return p.Points[i], i
	}

	i = (i - 1) % len(p.Points)
	return p.Points[i], i
}

func (p *Polygon[T]) Center() Vector2[T] {
	if p.center != nil {
		return *p.center
	}
	center := Vector2[T]{}
	ratio := 1 / T(len(p.Points))
	for _, point := range p.Points {
		center = center.Add(point.MultiplyScalar(ratio))
	}
	p.center = &center

	return *p.center
}

func (p Polygon[T]) Center2() Vector2[T] {
	center := Vector2[T]{}
	var weight T
	last := len(p.Points) - 1
	for i := range p.Points {
		var previous Vector2[T]
		point := p.Points[i]
		var next Vector2[T]
		if i == 0 {
			previous = p.Points[last]
			next = p.Points[i+1]
		} else if i == last {
			previous = p.Points[i-1]
			next = p.Points[0]
		} else {
			previous = p.Points[i-1]
			next = p.Points[i+1]
		}
		
		w := point.Sub(next).Length() + point.Sub(previous).Length()
		center.Add(point.MultiplyScalar(w))
		weight += w
	}

	return center.DivScalar(weight)
}

func (p Polygon[T]) TranslateToOrigin() (translationVector Vector2[T]) {
	translationVector = p.Center2()
	for i := range p.Points {
		p.Points[i].Sub(translationVector)
	}

	return translationVector
}

// https://stackoverflow.com/questions/471962/how-do-i-efficiently-determine-if-a-polygon-is-convex-non-convex-or-complex
func (p Polygon[T]) IsConvex() bool {
	l := len(p.Points) - 1
	m := l - 1
	j, k := 0, 0
	wasPositive := false
	for i := range p.Points {
		if i == m {
			j = i + 1
			k = 0
		} else if i == l {
			j = 0
			k = 1
		} else {
			j = i + 1
			k = i + 2
		}
		dx1 := p.Points[j].X - p.Points[i].X
		dy1 := p.Points[j].Y - p.Points[i].Y
		dx2 := p.Points[k].X - p.Points[j].X
		dy2 := p.Points[k].Y - p.Points[j].Y
		zcrossproduct := dx1*dy2 - dy1*dx2
		positive := zcrossproduct >= 0
		if i == 0 {
			wasPositive = positive
		} else {
			if (wasPositive && !positive) || (!wasPositive && positive) {
				return false
			}
		}
	}

	return true
}

func TransformToAxis[T Float](p Polygon[T], axis Vector2[T]) (min, max T, minp, maxp Vector2[T]) {
	min = T(math.MaxFloat32)
	max = T(-math.MaxFloat32)

	for _, p1 := range p.Points {
		v := p1.X*axis.X + p1.Y*axis.Y
		if v < min {
			minp = p1
			min = v
		}
		if v > max {
			maxp = p1
			max = v
		}
	}

	return
}

func (p Polygon[T]) ExtendSimple(e T) Polygon[T] {
	ps := make([]Vector2[T], 0, len(p.Points))

	var (
		p0, p1, p2 Vector2[T]
		l          = len(p.Points)
	)
	for i := range p.Points {
		if i == 0 {
			p0 = p.Points[l-1]
			p1 = p.Points[i]
			p2 = p.Points[i+1]
		} else if i == l-1 {
			p0 = p.Points[i-1]
			p1 = p.Points[i]
			p2 = p.Points[0]
		} else {
			p0 = p.Points[i-1]
			p1 = p.Points[i]
			p2 = p.Points[i+1]
		}

		n1 := GetNormal(p1.Sub(p0))
		n2 := GetNormal(p2.Sub(p1))

		//t1 := n1.Clone().Add(n2)
		//t2 := t1.Clone().MultiplyScalar(e)
		//t3 := p1.Clone().Add(t2)

		//fmt.Printf("t1: %v, t2: %v, t3: %v\n", t1, t2, t3)

		ps = append(ps, p1.Add(n1.Add(n2).Normalize().MultiplyScalar(e)))
	}

	return Polygon[T]{
		Points: ps,
	}
}

/*func (p *Polygon[T]) Transform(t *math32.Matrix4) *Polygon {
	newPoints := make([]math32.Vector2, 0, len(p.Points))
	for _, pp := range p.Points {
		newPoints = append(newPoints, math32.Vector2{
			X: pp.X*t[0] + t[1] + pp.Y*t[2] + t[3],
			Y: pp.X*t[2] + t[3] + pp.Y*t[10] + t[11],
		})
	}

	return &Polygon{
		Points: newPoints,
	}
}*/

func (p Polygon[T]) CCWPoints() {
	var ccw T
	last := len(p.Points) - 1
	var p0, p1, p2 Vector2[T]
	for i := range p.Points {
		if i == last {
			p0 = p.Points[i]
			p1 = p.Points[0]
			p2 = p.Points[1]
		} else if i == last-1 {
			p0 = p.Points[i]
			p1 = p.Points[i+1]
			p2 = p.Points[0]
		} else {
			p0 = p.Points[i]
			p1 = p.Points[i+1]
			p2 = p.Points[i+2]
		}

		ccw += CCW2D(p0, p1, p2)
	}

	// Polygon is already counterclockwise
	if ccw > 0 {
		return
	}

	// Reverse
	for i := len(p.Points)/2 - 1; i >= 0; i-- {
		opp := len(p.Points) - 1 - i
		p.Points[i], p.Points[opp] = p.Points[opp], p.Points[i]
	}
}

/*
*
https://web.archive.org/web/20100405070507/http://valis.cs.uiuc.edu/~sariel/research/CG/compgeom/msg00831.html
Let 'vertices' be an array of N pairs (x,y), indexed from 0
Let 'area' = 0.0
for i = 0 to N-1, do

	Let j = (i+1) mod N
	Let area = area + vertices[i].x * vertices[j].y
	Let area = area - vertices[i].y * vertices[j].x

end for
Return 'area'
*/
func (p Polygon[T]) Area() (area T) {
	for i := range p.Points {
		j := (i + 1) % len(p.Points)
		area += p.Points[i].X * p.Points[j].Y
		area -= p.Points[i].Y * p.Points[j].X
	}

	return area
}
