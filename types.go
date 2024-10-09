package gomath

type CenteredVolume[T Number] interface {
	GetCenter() Vector2[T]
}

type SimpleVolume[T Number] interface {
	CenteredVolume[T]

	SetCenter(Vector2[T])
	GetHalfWidth() T
	GetHalfHeight() T
	SetHalfWidth(T)
	SetHalfHeight(T)

	Scale(T)
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
	}

	return nil
}
