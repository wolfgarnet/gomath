package gomath

type CenteredVolume[T Number] interface {
	SetCenter(Vector2[T])
	GetCenter() Vector2[T]
	GetHalfWidth() T
	GetHalfHeight() T
	SetHalfWidth(T)
	SetHalfHeight(T)
}

func CloneCenteredVolume[T Number](v CenteredVolume[T]) CenteredVolume[T] {
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
