package collider

import "github.com/wolfgarnet/gomath"

func RayCircle[T gomath.Float](rayOriginX T, rayOriginY T, rayDirectionX T, rayDirectionY T, circleCenterX T, circleCenterY T, circleRadius T) (intersects int, p1 gomath.Vector2[T], p2 gomath.Vector2[T]) {

	toCircleX := circleCenterX - rayOriginX
	toCircleY := circleCenterY - rayOriginY
	dot := rayDirectionX*toCircleX + rayDirectionY*toCircleY

	projectionX := rayDirectionX * dot
	projectionY := rayDirectionY * dot

	u2x := toCircleX - projectionX
	u2y := toCircleY - projectionY
	d := gomath.Sqrt(u2x*u2x + u2y*u2y)

	if d > circleRadius {
		return 0, p1, p2
	}

	m := gomath.Sqrt(circleRadius*circleRadius - d*d)

	if m == 0 {
		p1.X = rayOriginX + projectionX
		p1.Y = rayOriginY + projectionY
		return 1, p1, p2
	}

	p1.X = rayOriginX + projectionX + (rayDirectionX * m)
	p1.Y = rayOriginY + projectionY + (rayDirectionY * m)
	p2.X = rayOriginX + projectionX - (rayDirectionX * m)
	p2.Y = rayOriginY + projectionY - (rayDirectionY * m)

	return 2, p1, p2
}
