package intersection

import "github.com/wolfgarnet/gomath"

/*
https://www.geeksforgeeks.org/check-if-a-point-is-inside-outside-or-on-the-ellipse/
*/

func EllipsoidPoint[T gomath.Number](point gomath.Vector2[T], ellipsoid gomath.Ellipsoid[T]) bool {
	// Normalize the point in the ellipsoid's local space
	normalizedPoint := gomath.Vector2[T]{X: point.X / ellipsoid.Radii.X, Y: point.Y / ellipsoid.Radii.Y}

	// Check if the normalized point is inside the unit circle
	return normalizedPoint.Length() <= 1.0
}

func EllipsoidEllipsoid[T gomath.Number](e1, e2 gomath.Ellipsoid[T]) bool {
	ellipsoid2InEllipsoid1Space := gomath.TransformEllipsoidToEllipsoidSpace(e2, e1)

	// Get the distance between the ellipsoid centers
	distance := e1.Position.DistanceTo(ellipsoid2InEllipsoid1Space.Position)

	// Sum of radii
	sumRadii := e1.Radii.Length() + ellipsoid2InEllipsoid1Space.Radii.Length()

	return distance <= sumRadii
}
