package gomath

import "fmt"

// Matrix3 is a column-major 3x3 matrix representing a LINEAR transform in 3D:
// rotation, scale and shear about the origin, with no translation. The element
// in row r, column c is stored at index c*3+r, so the layout is:
//
//	m[0] m[3] m[6]
//	m[1] m[4] m[7]
//	m[2] m[5] m[8]
//
// # Matrix3 versus Matrix
//
// Both are [9]T and both are column-major, but they mean different things and
// are not interchangeable:
//
//   - Matrix is a 2D AFFINE transform. Its columns 0 and 1 are a 2D basis and
//     column 2 is a translation, so it moves points around a plane.
//   - Matrix3 is a 3D LINEAR basis. All three columns are basis vectors and
//     there is no translation anywhere in it.
//
// Matrix3 exists mainly to carry the normal matrix (see Matrix4.NormalMatrix):
// under non-uniform scale, surface normals must be transformed by the
// inverse-transpose of the upper-left 3x3, not by the transform itself.
//
// Receivers follow the same rule as Matrix4: methods that read the matrix to
// produce a different kind of value take a pointer receiver, methods that
// produce a new Matrix3 or that appear in chained expressions take a value
// receiver. See the Matrix4 doc comment for the measurements behind that split.
type Matrix3[T Float] [9]T

// NewMatrix3 returns the 3x3 identity matrix.
func NewMatrix3[T Float]() Matrix3[T] {
	return Identity3[T]()
}

// Identity3 returns the 3x3 identity matrix.
func Identity3[T Float]() Matrix3[T] {
	return Matrix3[T]{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
}

// NewMatrix3From builds a matrix from elements supplied row by row, so the
// arguments read in the same visual order as the matrix is written on paper
// while the storage stays column-major. This mirrors Matrix.Set and
// NewMatrix4From.
func NewMatrix3From[T Float](
	n11, n12, n13 T,
	n21, n22, n23 T,
	n31, n32, n33 T,
) Matrix3[T] {
	return Matrix3[T]{
		n11, n21, n31,
		n12, n22, n32,
		n13, n23, n33,
	}
}

// Basis3 builds a matrix whose columns are the given basis vectors.
func Basis3[T Float](x, y, z Vector3[T]) Matrix3[T] {
	return Matrix3[T]{
		x.X, x.Y, x.Z,
		y.X, y.Y, y.Z,
		z.X, z.Y, z.Z,
	}
}

// At returns the element at row r, column c.
func (m *Matrix3[T]) At(r, c int) T {
	return m[c*3+r]
}

// Column returns the c'th basis column.
func (m *Matrix3[T]) Column(c int) Vector3[T] {
	return Vector3[T]{X: m[c*3], Y: m[c*3+1], Z: m[c*3+2]}
}

// Mul returns the matrix product m*b as a new matrix. Transforms compose
// right-to-left: (m*b) applies b first, then m.
func (m Matrix3[T]) Mul(b Matrix3[T]) Matrix3[T] {
	var out Matrix3[T]
	for c := 0; c < 3; c++ {
		for r := 0; r < 3; r++ {
			out[c*3+r] = m[r]*b[c*3] +
				m[3+r]*b[c*3+1] +
				m[6+r]*b[c*3+2]
		}
	}

	return out
}

// MultiplyScalar scales every element by s.
func (m Matrix3[T]) MultiplyScalar(s T) Matrix3[T] {
	for i := range m {
		m[i] *= s
	}

	return m
}

// Transpose returns the transpose of m: element (r, c) becomes (c, r).
func (m Matrix3[T]) Transpose() Matrix3[T] {
	return Matrix3[T]{
		m[0], m[3], m[6],
		m[1], m[4], m[7],
		m[2], m[5], m[8],
	}
}

// TransformVector3 applies the linear transform to v. There is no translation
// in a Matrix3, so points and directions transform identically.
func (m *Matrix3[T]) TransformVector3(v Vector3[T]) Vector3[T] {
	return Vector3[T]{
		X: m[0]*v.X + m[3]*v.Y + m[6]*v.Z,
		Y: m[1]*v.X + m[4]*v.Y + m[7]*v.Z,
		Z: m[2]*v.X + m[5]*v.Y + m[8]*v.Z,
	}
}

// Determinant returns the determinant of m, which is zero exactly when m is
// singular.
func (m Matrix3[T]) Determinant() T {
	return m[0]*(m[4]*m[8]-m[7]*m[5]) -
		m[3]*(m[1]*m[8]-m[7]*m[2]) +
		m[6]*(m[1]*m[5]-m[4]*m[2])
}

// Inverse returns the inverse of m and true, or the identity and false if m is
// singular.
func (m Matrix3[T]) Inverse() (Matrix3[T], bool) {
	determinant := m.Determinant()
	if determinant == 0 {
		return Identity3[T](), false
	}

	inverseDeterminant := 1 / determinant

	// The adjugate, written column by column.
	return Matrix3[T]{
		(m[4]*m[8] - m[7]*m[5]) * inverseDeterminant,
		(m[7]*m[2] - m[1]*m[8]) * inverseDeterminant,
		(m[1]*m[5] - m[4]*m[2]) * inverseDeterminant,

		(m[6]*m[5] - m[3]*m[8]) * inverseDeterminant,
		(m[0]*m[8] - m[6]*m[2]) * inverseDeterminant,
		(m[3]*m[2] - m[0]*m[5]) * inverseDeterminant,

		(m[3]*m[7] - m[6]*m[4]) * inverseDeterminant,
		(m[6]*m[1] - m[0]*m[7]) * inverseDeterminant,
		(m[0]*m[4] - m[3]*m[1]) * inverseDeterminant,
	}, true
}

// Floats32 returns the elements as a [9]float32 in the same column-major order,
// ready to upload with glUniformMatrix3fv(..., transpose=false, ...).
func (m *Matrix3[T]) Floats32() [9]float32 {
	var out [9]float32
	for i := range m {
		out[i] = float32(m[i])
	}

	return out
}

// Equals reports whether every element of m equals the corresponding element of
// b exactly. Prefer ApproxEquals after any rotation or inversion.
func (m Matrix3[T]) Equals(b Matrix3[T]) bool {
	return m == b
}

// ApproxEquals reports whether every element of m is within tol of the
// corresponding element of b.
func (m Matrix3[T]) ApproxEquals(b Matrix3[T], tol T) bool {
	for i := range m {
		if Abs(m[i]-b[i]) > tol {
			return false
		}
	}

	return true
}

// String renders the matrix in row-major form for debugging, so it reads the way
// it would be written on paper.
func (m Matrix3[T]) String() string {
	return fmt.Sprintf("Matrix3[[%v %v %v] [%v %v %v] [%v %v %v]]",
		m[0], m[3], m[6],
		m[1], m[4], m[7],
		m[2], m[5], m[8])
}

// Basis returns the upper-left 3x3 of m: the rotation/scale/shear part, with the
// translation discarded.
func (m *Matrix4[T]) Basis() Matrix3[T] {
	return Matrix3[T]{
		m[0], m[1], m[2],
		m[4], m[5], m[6],
		m[8], m[9], m[10],
	}
}

// NormalMatrix returns the basis to transform surface normals by: the
// inverse-transpose of the upper-left 3x3.
//
// Transforming a normal by the model matrix itself is only correct under
// rotation and uniform scale. Under non-uniform scale it shears the normal off
// the surface — squashing a sphere along Y tilts its normals the wrong way — so
// lighting goes visibly wrong. The inverse-transpose keeps normals
// perpendicular to the transformed surface.
//
// If the basis is singular (an axis has been scaled to zero) there is no valid
// normal basis and the identity is returned along with false.
func (m *Matrix4[T]) NormalMatrix() (Matrix3[T], bool) {
	basis := m.Basis()

	inverse, ok := basis.Inverse()
	if !ok {
		return Identity3[T](), false
	}

	return inverse.Transpose(), true
}
