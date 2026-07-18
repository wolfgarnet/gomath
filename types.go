package gomath

type CenteredVolume[T Number] interface {
	GetCenter() Vector2[T]
}

// SimpleVolume is a read-only bounding-volume view: a centre plus half-extents.
// It carries no mutators, so it is satisfied by both value and pointer forms of
// the concrete volumes (Circle, AABB). To transform a volume, use the concrete
// type's functional helpers (Scaled, WithCenter, Translate), which return a new
// value.
type SimpleVolume[T Number] interface {
	CenteredVolume[T]

	GetHalfWidth() T
	GetHalfHeight() T
}

func CloneSimpleVolume[T Number](v SimpleVolume[T]) SimpleVolume[T] {
	switch v2 := v.(type) {
	case *Circle[T]:
		return &Circle[T]{
			Center: v2.Center,
			Radius: v2.Radius,
		}

	case *AABB[T]:
		return &AABB[T]{
			Center:     v2.Center,
			HalfWidth:  v2.HalfWidth,
			HalfHeight: v2.HalfHeight,
		}

	case *Ellipsoid[T]:
		return &Ellipsoid[T]{
			Position: v2.Position,
			Radii:    v2.Radii,
		}

	case *SquaredAABB2D[T]:
		return &SquaredAABB2D[T]{
			Center:   v2.Center,
			HalfSize: v2.HalfSize,
		}
	}

	return nil
}
