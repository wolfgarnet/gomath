package gomath

type Ellipsoid[T Number] struct {
	Position Vector2[T]
	Radii    Vector2[T]
}

func NewEllipsoid[T Number](position Vector2[T], radii Vector2[T]) Ellipsoid[T] {
	return Ellipsoid[T]{
		Position: position,
		Radii:    radii,
	}
}

// TransformEllipsoidToEllipsoidSpace Translates the ellipsoid position to the reference ellipsoid's local space
func TransformEllipsoidToEllipsoidSpace[T Number](ellipsoid Ellipsoid[T], referenceEllipsoid Ellipsoid[T]) Ellipsoid[T] {
	translatedPosition := Vector2[T]{
		X: ellipsoid.Position.X - referenceEllipsoid.Position.X,
		Y: ellipsoid.Position.Y - referenceEllipsoid.Position.Y,
	}

	return Ellipsoid[T]{
		Position: translatedPosition,
		Radii:    ellipsoid.Radii,
	}
}

func TransformPointToEllipsoidSpace[T Number](point Vector2[T], ellipsoid Ellipsoid[T]) Vector2[T] {
	return Vector2[T]{
		X: point.X - ellipsoid.Position.X,
		Y: point.Y - ellipsoid.Position.Y,
	}
}

func TransformLineToEllipsoidSpace[T Number](line LineSegment[T], ellipsoid Ellipsoid[T]) LineSegment[T] {
	return LineSegment[T]{
		P1: TransformPointToEllipsoidSpace(line.P1, ellipsoid),
		P2: TransformPointToEllipsoidSpace(line.P2, ellipsoid),
	}
}
