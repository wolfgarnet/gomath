package gomath

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestMatrix_Translation(t *testing.T) {
	m := Matrix[float32]{}.Identity()
	m.Translation(1, 1)
	v := Vector2[float32]{}

	v2 := v.ApplyMatrix(m)

	spew.Dump(v2)
}

func TestMatrix_Rotation(t *testing.T) {
	m := Matrix[float32]{}.Identity()
	m.Rotation(-3.14 / 2)
	v := Vector2[float32]{1, 0}

	v2 := v.ApplyMatrix(m)

	spew.Dump(v2)
}
