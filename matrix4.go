package gomath

import (
	"fmt"
	"math"
)

// Matrix4 is a column-major 4x4 matrix representing a 3D affine or projective
// transform. The element in row r, column c is stored at index c*4+r, so the
// layout is:
//
//	m[0] m[4] m[ 8] m[12]
//	m[1] m[5] m[ 9] m[13]
//	m[2] m[6] m[10] m[14]
//	m[3] m[7] m[11] m[15]
//
// Columns 0, 1 and 2 hold the linear (rotation/scale/shear) basis and column 3
// (indices 12, 13, 14) holds the translation. This matches Matrix's 2D layout
// (c*3+r) and is also the memory order OpenGL expects, so Floats32 can be handed
// to glUniformMatrix4fv with transpose=false and no repacking.
//
// The package convention is right-handed with Y up, and the projection builders
// (PerspectiveMatrix4, OrthographicMatrix4, FrustumMatrix4) map the view volume
// to the OpenGL clip space where the near plane is z=-1 and the far plane z=+1.
//
// Unlike Matrix, Matrix4 is constrained to floating-point T: every meaningful
// operation on a 3D transform (inversion, projection, rotation, normalisation of
// the basis) divides or goes through trigonometry, so an integer instantiation
// has no valid use and is rejected at compile time rather than silently
// truncating.
//
// # Receivers
//
// Matrix4 deliberately mixes receiver kinds, and the rule is:
//
//   - Methods that read the matrix to produce a *different* kind of value, and
//     that get called in loops — TransformPoint, TransformDirection,
//     TransformVector4, At, GetPosition, GetScale, Floats32 — take a POINTER
//     receiver.
//   - Methods that produce a new Matrix4, or that appear inside chained
//     expressions — Mul, Transpose, Inverse, Translated, Scaled, RotatedAxis,
//     WithPosition, Equals, ApproxEquals, String — take a VALUE receiver.
//
// The package-wide "value receivers, no mutation" convention was written for
// small primitives (Vector2 is 16 bytes and passes in registers). Matrix4 is
// 128 bytes at float64 and is an array, so a value receiver copies the whole
// thing on every call. That is free when the arithmetic dominates — Mul
// benchmarks identically either way, 35.17ns against 35.16ns — but it is not
// free when little work is done per byte copied: transforming 4096 points is
// 1.54x faster through a pointer, and reading the translation out of a slice of
// matrices is 1.78x faster, because a value receiver must touch every cache
// line of the matrix to read three floats.
//
// None of the pointer-receiver methods mutate; the pointer is purely to avoid
// the copy. The value receivers stay because Go requires addressability for a
// pointer-receiver call, so m.Mul(b).Translated(...) would stop compiling.
//
// See matrix4_additions.go for the builders and combinators.
type Matrix4[T Float] [16]T

// NewMatrix4 returns the 4x4 identity matrix.
func NewMatrix4[T Float]() Matrix4[T] {
	return Identity4[T]()
}

// Identity4 returns the 4x4 identity matrix.
func Identity4[T Float]() Matrix4[T] {
	return Matrix4[T]{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// NewMatrix4From builds a matrix from elements supplied row by row, starting at
// row 1 column 1, then row 1 column 2, and so forth. The arguments therefore
// read in the same visual order as the matrix is written on paper, while the
// storage stays column-major. This mirrors Matrix.Set.
func NewMatrix4From[T Float](
	n11, n12, n13, n14 T,
	n21, n22, n23, n24 T,
	n31, n32, n33, n34 T,
	n41, n42, n43, n44 T,
) Matrix4[T] {
	return Matrix4[T]{
		n11, n21, n31, n41,
		n12, n22, n32, n42,
		n13, n23, n33, n43,
		n14, n24, n34, n44,
	}
}

// At returns the element at row r, column c.
func (m *Matrix4[T]) At(r, c int) T {
	return m[c*4+r]
}

// Mul returns the matrix product m*b as a new matrix, leaving both operands
// unchanged. Transforms compose right-to-left: (m*b) applies b first, then m.
func (m Matrix4[T]) Mul(b Matrix4[T]) Matrix4[T] {
	var out Matrix4[T]
	for c := 0; c < 4; c++ {
		for r := 0; r < 4; r++ {
			out[c*4+r] = m[r]*b[c*4] +
				m[4+r]*b[c*4+1] +
				m[8+r]*b[c*4+2] +
				m[12+r]*b[c*4+3]
		}
	}

	return out
}

// MultiplyScalar scales every element by s.
func (m Matrix4[T]) MultiplyScalar(s T) Matrix4[T] {
	for i := range m {
		m[i] *= s
	}

	return m
}

// Transpose returns the transpose of m: element (r, c) becomes (c, r).
func (m Matrix4[T]) Transpose() Matrix4[T] {
	var out Matrix4[T]
	for c := 0; c < 4; c++ {
		for r := 0; r < 4; r++ {
			out[r*4+c] = m[c*4+r]
		}
	}

	return out
}

// TransformPoint applies the full transform to v, treating it as the homogeneous
// point (x, y, z, 1) and dividing the result by its W. The divide makes it
// correct for projection matrices as well as affine ones; for an affine matrix W
// is 1 and the divide is a no-op.
func (m *Matrix4[T]) TransformPoint(v Vector3[T]) Vector3[T] {
	x := m[0]*v.X + m[4]*v.Y + m[8]*v.Z + m[12]
	y := m[1]*v.X + m[5]*v.Y + m[9]*v.Z + m[13]
	z := m[2]*v.X + m[6]*v.Y + m[10]*v.Z + m[14]
	w := m[3]*v.X + m[7]*v.Y + m[11]*v.Z + m[15]

	if w != 0 && w != 1 {
		return Vector3[T]{X: x / w, Y: y / w, Z: z / w}
	}

	return Vector3[T]{X: x, Y: y, Z: z}
}

// TransformDirection applies only the linear part of the transform (rotation,
// scale, shear) to v, ignoring translation and the homogeneous divide. Use it
// for directions and velocities. Note that it is NOT correct for normals under
// non-uniform scale: normals need the inverse-transpose basis, which arrives
// with Matrix3 in the next phase.
func (m *Matrix4[T]) TransformDirection(v Vector3[T]) Vector3[T] {
	return Vector3[T]{
		X: m[0]*v.X + m[4]*v.Y + m[8]*v.Z,
		Y: m[1]*v.X + m[5]*v.Y + m[9]*v.Z,
		Z: m[2]*v.X + m[6]*v.Y + m[10]*v.Z,
	}
}

// TransformVector4 returns m*v, treating v as a homogeneous column vector. It
// performs no perspective divide, so the caller keeps the W coordinate; use
// Vector4.PerspectiveDivide to complete a projection.
func (m *Matrix4[T]) TransformVector4(v Vector4[T]) Vector4[T] {
	return Vector4[T]{
		X: m[0]*v.X + m[4]*v.Y + m[8]*v.Z + m[12]*v.W,
		Y: m[1]*v.X + m[5]*v.Y + m[9]*v.Z + m[13]*v.W,
		Z: m[2]*v.X + m[6]*v.Y + m[10]*v.Z + m[14]*v.W,
		W: m[3]*v.X + m[7]*v.Y + m[11]*v.Z + m[15]*v.W,
	}
}

// TransformBox3 returns the axis-aligned box enclosing box after this transform.
//
// The result is generally larger than the transformed box itself: rotating a box
// and re-fitting an axis-aligned one around it cannot be tight. That is the right
// trade for frustum culling, where this runs once per object per frame — the
// alternative is building an OBB3, which costs three square roots to normalise
// the basis, and a slightly looser bound only ever means a few extra draws.
//
// Use NewOBB3FromMatrix4 when tightness matters more than speed.
//
// It lives here rather than on Box3 because Box3 allows any T Number while a
// transform needs T Float, and a method cannot narrow its receiver's constraint.
func (m *Matrix4[T]) TransformBox3(box Box3[T]) Box3[T] {
	// The extent along each world axis is the half-size projected through the
	// absolute value of the basis. No square roots involved.
	half := Vector3[T]{
		X: Abs(m[0])*box.HalfSize.X + Abs(m[4])*box.HalfSize.Y + Abs(m[8])*box.HalfSize.Z,
		Y: Abs(m[1])*box.HalfSize.X + Abs(m[5])*box.HalfSize.Y + Abs(m[9])*box.HalfSize.Z,
		Z: Abs(m[2])*box.HalfSize.X + Abs(m[6])*box.HalfSize.Y + Abs(m[10])*box.HalfSize.Z,
	}

	return Box3[T]{
		Center:   m.TransformPoint(box.Center),
		HalfSize: half,
	}
}

// GetPosition returns the translation stored in column 3.
func (m *Matrix4[T]) GetPosition() Vector3[T] {
	return Vector3[T]{X: m[12], Y: m[13], Z: m[14]}
}

// WithPosition returns a copy of m with its translation column replaced.
func (m Matrix4[T]) WithPosition(position Vector3[T]) Matrix4[T] {
	m[12] = position.X
	m[13] = position.Y
	m[14] = position.Z

	return m
}

// GetScale returns the scale magnitudes encoded in the linear part, as the
// lengths of the three basis columns. The values are always non-negative, so a
// reflection is not distinguished from a positive scale.
func (m *Matrix4[T]) GetScale() Vector3[T] {
	return Vector3[T]{
		X: T(math.Sqrt(float64(m[0]*m[0] + m[1]*m[1] + m[2]*m[2]))),
		Y: T(math.Sqrt(float64(m[4]*m[4] + m[5]*m[5] + m[6]*m[6]))),
		Z: T(math.Sqrt(float64(m[8]*m[8] + m[9]*m[9] + m[10]*m[10]))),
	}
}

// adjugateAndDeterminant returns the adjugate (the transposed cofactor matrix)
// together with the determinant. Inverse is adjugate/determinant, and both
// Determinant and Inverse share this so the cofactors are written once.
func (m Matrix4[T]) adjugateAndDeterminant() (Matrix4[T], T) {
	var adj Matrix4[T]

	adj[0] = m[5]*m[10]*m[15] - m[5]*m[11]*m[14] - m[9]*m[6]*m[15] +
		m[9]*m[7]*m[14] + m[13]*m[6]*m[11] - m[13]*m[7]*m[10]
	adj[4] = -m[4]*m[10]*m[15] + m[4]*m[11]*m[14] + m[8]*m[6]*m[15] -
		m[8]*m[7]*m[14] - m[12]*m[6]*m[11] + m[12]*m[7]*m[10]
	adj[8] = m[4]*m[9]*m[15] - m[4]*m[11]*m[13] - m[8]*m[5]*m[15] +
		m[8]*m[7]*m[13] + m[12]*m[5]*m[11] - m[12]*m[7]*m[9]
	adj[12] = -m[4]*m[9]*m[14] + m[4]*m[10]*m[13] + m[8]*m[5]*m[14] -
		m[8]*m[6]*m[13] - m[12]*m[5]*m[10] + m[12]*m[6]*m[9]

	adj[1] = -m[1]*m[10]*m[15] + m[1]*m[11]*m[14] + m[9]*m[2]*m[15] -
		m[9]*m[3]*m[14] - m[13]*m[2]*m[11] + m[13]*m[3]*m[10]
	adj[5] = m[0]*m[10]*m[15] - m[0]*m[11]*m[14] - m[8]*m[2]*m[15] +
		m[8]*m[3]*m[14] + m[12]*m[2]*m[11] - m[12]*m[3]*m[10]
	adj[9] = -m[0]*m[9]*m[15] + m[0]*m[11]*m[13] + m[8]*m[1]*m[15] -
		m[8]*m[3]*m[13] - m[12]*m[1]*m[11] + m[12]*m[3]*m[9]
	adj[13] = m[0]*m[9]*m[14] - m[0]*m[10]*m[13] - m[8]*m[1]*m[14] +
		m[8]*m[2]*m[13] + m[12]*m[1]*m[10] - m[12]*m[2]*m[9]

	adj[2] = m[1]*m[6]*m[15] - m[1]*m[7]*m[14] - m[5]*m[2]*m[15] +
		m[5]*m[3]*m[14] + m[13]*m[2]*m[7] - m[13]*m[3]*m[6]
	adj[6] = -m[0]*m[6]*m[15] + m[0]*m[7]*m[14] + m[4]*m[2]*m[15] -
		m[4]*m[3]*m[14] - m[12]*m[2]*m[7] + m[12]*m[3]*m[6]
	adj[10] = m[0]*m[5]*m[15] - m[0]*m[7]*m[13] - m[4]*m[1]*m[15] +
		m[4]*m[3]*m[13] + m[12]*m[1]*m[7] - m[12]*m[3]*m[5]
	adj[14] = -m[0]*m[5]*m[14] + m[0]*m[6]*m[13] + m[4]*m[1]*m[14] -
		m[4]*m[2]*m[13] - m[12]*m[1]*m[6] + m[12]*m[2]*m[5]

	adj[3] = -m[1]*m[6]*m[11] + m[1]*m[7]*m[10] + m[5]*m[2]*m[11] -
		m[5]*m[3]*m[10] - m[9]*m[2]*m[7] + m[9]*m[3]*m[6]
	adj[7] = m[0]*m[6]*m[11] - m[0]*m[7]*m[10] - m[4]*m[2]*m[11] +
		m[4]*m[3]*m[10] + m[8]*m[2]*m[7] - m[8]*m[3]*m[6]
	adj[11] = -m[0]*m[5]*m[11] + m[0]*m[7]*m[9] + m[4]*m[1]*m[11] -
		m[4]*m[3]*m[9] - m[8]*m[1]*m[7] + m[8]*m[3]*m[5]
	adj[15] = m[0]*m[5]*m[10] - m[0]*m[6]*m[9] - m[4]*m[1]*m[10] +
		m[4]*m[2]*m[9] + m[8]*m[1]*m[6] - m[8]*m[2]*m[5]

	determinant := m[0]*adj[0] + m[1]*adj[4] + m[2]*adj[8] + m[3]*adj[12]

	return adj, determinant
}

// Determinant returns the determinant of m. It is zero exactly when m is
// singular (non-invertible).
func (m Matrix4[T]) Determinant() T {
	_, determinant := m.adjugateAndDeterminant()

	return determinant
}

// Inverse returns the inverse of m and true, or the identity and false if m is
// singular. The signature mirrors Matrix.Inverse.
func (m Matrix4[T]) Inverse() (Matrix4[T], bool) {
	adjugate, determinant := m.adjugateAndDeterminant()
	if determinant == 0 {
		return Identity4[T](), false
	}

	return adjugate.MultiplyScalar(1 / determinant), true
}

// Floats32 returns the elements as a plain [16]float32 in the same column-major
// order, ready to upload with glUniformMatrix4fv(..., transpose=false, ...)
// without repacking or transposing.
func (m *Matrix4[T]) Floats32() [16]float32 {
	var out [16]float32
	for i := range m {
		out[i] = float32(m[i])
	}

	return out
}

// Equals reports whether every element of m equals the corresponding element of
// b exactly. Prefer ApproxEquals for anything that has been through a rotation
// or an inversion.
func (m Matrix4[T]) Equals(b Matrix4[T]) bool {
	return m == b
}

// ApproxEquals reports whether every element of m is within tol of the
// corresponding element of b.
func (m Matrix4[T]) ApproxEquals(b Matrix4[T], tol T) bool {
	for i := range m {
		if Abs(m[i]-b[i]) > tol {
			return false
		}
	}

	return true
}

// String renders the matrix in row-major form for debugging, so it reads the way
// it would be written on paper.
func (m Matrix4[T]) String() string {
	return fmt.Sprintf("Matrix4[[%v %v %v %v] [%v %v %v %v] [%v %v %v %v] [%v %v %v %v]]",
		m[0], m[4], m[8], m[12],
		m[1], m[5], m[9], m[13],
		m[2], m[6], m[10], m[14],
		m[3], m[7], m[11], m[15])
}
