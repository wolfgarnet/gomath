package gomath

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatrix_Translation(t *testing.T) {
	m := Matrix[float32]{}
	m.Identity()
	m.Translation(1, 1)
	v := Vector2[float32]{}

	v2 := v.ApplyMatrix(m)

	spew.Dump(v2)
}

func TestMatrix_Rotation(t *testing.T) {
	m := Matrix[float32]{}
	m.Identity()
	m.Rotation(-3.14 / 2)
	v := Vector2[float32]{1, 0}

	v2 := v.ApplyMatrix(m)

	spew.Dump(v2)
}

func TestMatrix_SetPosition(t *testing.T) {
	m := Matrix[float32]{}
	m.Identity()
	spew.Dump(m)
	m.SetPosition(1, 2)
	spew.Dump(m)
}

func TestMatrix_SetDirection(t *testing.T) {
	tests := map[string]struct {
		direction Vector2[float32]
	}{
		"right": {
			direction: Vector2[float32]{1, 0},
		},
		"up": {
			direction: Vector2[float32]{0, 1},
		},
		"up left": {
			direction: Vector2[float32]{-1, 1},
		},
		"down right": {
			direction: Vector2[float32]{1, -1},
		},
		"up right": {
			direction: Vector2[float32]{1, 1},
		},
		"left": {
			direction: Vector2[float32]{-1, 0},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewMatrix[float32]()
			normalizedTestDirection := test.direction.Normalize()
			m.SetDirection(normalizedTestDirection)

			direction := m.GetDirection()
			assert.Truef(t, assertAlmostEqualVectors(normalizedTestDirection, direction, 0.0001), "Expected %v, got %v", normalizedTestDirection, direction)
		})
	}
}
