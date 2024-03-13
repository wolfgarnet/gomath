package gomath

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"math"
)

const Epsilon = 0.00001

type Number interface {
	constraints.Signed | constraints.Float
}

type Float interface {
	constraints.Float
}

type Vector2[T Number] struct {
	X T
	Y T
}

func NewVector2[T Number](x T, y T) Vector2[T] {
	return Vector2[T]{
		X: x,
		Y: y,
	}
}

func (v Vector2[T]) Set(x T, y T) Vector2[T] {
	v.X = x
	v.Y = y

	return v
}

// GetCenter is implemented just to satisfy the CenteredVolume interface
func (v Vector2[T]) GetCenter() Vector2[T] {
	return v
}

func (v Vector2[T]) Add(v2 Vector2[T]) Vector2[T] {
	return Vector2[T]{
		X: v.X + v2.X,
		Y: v.Y + v2.Y,
	}
}

func (v Vector2[T]) Component(i int) T {
	switch i {
	case 0:
		return v.X
	case 1:
		return v.Y
	default:
		panic("out of range")
	}
}

func (v Vector2[T]) AddScalar(scalar T) Vector2[T] {
	return Vector2[T]{
		X: v.X + scalar,
		Y: v.Y + scalar,
	}
}

func (v Vector2[T]) AddScalars(x T, y T) Vector2[T] {
	return Vector2[T]{
		X: v.X + x,
		Y: v.Y + y,
	}
}

// Sub subtracts the coordinates of the given Vector2 from the coordinates of the current Vector2
func (v Vector2[T]) Sub(v2 Vector2[T]) Vector2[T] {
	return Vector2[T]{
		X: v.X - v2.X,
		Y: v.Y - v2.Y,
	}
}

func (v Vector2[T]) SubScalar(scalar T) Vector2[T] {
	return Vector2[T]{
		X: v.X - scalar,
		Y: v.Y - scalar,
	}
}

func (v Vector2[T]) SubScalars(x T, y T) Vector2[T] {
	return Vector2[T]{
		X: v.X - x,
		Y: v.Y - y,
	}
}

// Mul multiplies the current vector by another vector element-wise and returns the result as a new Vector2.
func (v Vector2[T]) Mul(v2 Vector2[T]) Vector2[T] {
	return Vector2[T]{
		X: v.X * v2.X,
		Y: v.Y * v2.Y,
	}
}

// MultiplyScalar multiplies the current vector by a scalar value.
// It returns a new vector with the result of the multiplication.
func (v Vector2[T]) MultiplyScalar(scalar T) Vector2[T] {
	return Vector2[T]{
		X: v.X * scalar,
		Y: v.Y * scalar,
	}
}

func (v Vector2[T]) MultiplyScalars(x T, y T) Vector2[T] {
	return Vector2[T]{
		X: v.X * x,
		Y: v.Y * y,
	}
}

func (v Vector2[T]) Div(v2 Vector2[T]) Vector2[T] {
	return Vector2[T]{
		X: v.X / v2.X,
		Y: v.Y / v2.Y,
	}
}

func (v Vector2[T]) DivScalar(scalar T) Vector2[T] {
	return Vector2[T]{X: v.X / scalar, Y: v.Y / scalar}
}

func (v Vector2[T]) DivScalars(x T, y T) Vector2[T] {
	return Vector2[T]{
		X: v.X / x,
		Y: v.Y / y,
	}
}

// Distance calculates the Euclidean distance between two Vector2 points.
func (v Vector2[T]) Distance(v2 Vector2[T]) float64 {
	dx := v.X - v2.X
	dy := v.Y - v2.Y
	return math.Sqrt(float64(dx*dx + dy*dy))
}

// Dot returns the dot product of the given vector and `v2`.
func (v Vector2[T]) Dot(v2 Vector2[T]) T {
	return v.X*v2.X + v.Y*v2.Y
}

func (v Vector2[T]) Lerp(v2 Vector2[T], t T) Vector2[T] {
	return Vector2[T]{
		v.X + (v2.X-v.X)*t,
		v.Y + (v2.Y-v.Y)*t,
	}
}

func (v Vector2[T]) Length64() float64 {
	return math.Sqrt(float64(v.X)*float64(v.X) + float64(v.Y)*float64(v.Y))
}

func (v Vector2[T]) Length() T {
	return T(math.Sqrt(float64(v.X)*float64(v.X) + float64(v.Y)*float64(v.Y)))
}

func (v Vector2[T]) LengthSq() T {
	//return T(math.Sqrt(float64(v.X)*float64(v.X) + float64(v.Y)*float64(v.Y)))
	return v.X*v.X + v.Y*v.Y
}

func (v Vector2[T]) Length32() float32 {
	return float32(math.Sqrt(float64(v.X)*float64(v.X) + float64(v.Y)*float64(v.Y)))
}

func (v Vector2[T]) SetLength(l T) Vector2[T] {
	oldLength := v.Length()
	if oldLength != 0 && l != oldLength {
		v.MultiplyScalar(l / oldLength)
	}

	return v
}

func (v Vector2[T]) Normalize() Vector2[T] {
	m := v.Length()

	if float64(m) > Epsilon {
		return v.DivScalar(m)
	} else {
		return v
	}
}

func (v Vector2[T]) Reflect(v2 Vector2[T]) Vector2[T] {
	factor := -T(2.0) * v.Dot(v2)
	return NewVector2(
		factor*v.X+v2.X,
		factor*v.Y+v2.Y,
	)
}

func (v Vector2[T]) Equals(v2 Vector2[T]) bool {
	return v.X == v2.X && v.Y == v2.Y
}

func (v Vector2[T]) AngleTo(v2 Vector2[T]) T {
	theta := v.Dot(v2) / v.Length() * v2.Length()
	return Acos(Clamp[T](theta, T(-1.0), 1))
}

func (v Vector2[T]) AngleToWithSign(v2 Vector2[T]) T {
	theta := v.AngleTo(v2)
	cross := v.X*v2.Y - v.Y*v2.X

	if cross < 0 {
		theta = -theta
	}

	return theta
}

// DistanceTo returns the distance to the given vector
func (v Vector2[T]) DistanceTo(v2 Vector2[T]) T {

	return Sqrt(v.DistanceToSquared(v2))
}

// DistanceToSquared returns the squared distance to the given vector
func (v Vector2[T]) DistanceToSquared(v2 Vector2[T]) T {

	dx := v.X - v2.X
	dy := v.Y - v2.Y
	return dx*dx + dy*dy
}

func (v Vector2[T]) ApplyMatrix(m Matrix[T]) Vector2[T] {
	x := v.X
	y := v.Y
	v.X = m[0]*x + m[3]*y + m[6]
	v.Y = m[1]*x + m[4]*y + m[7]

	return v
}

const (
	pi180 = 180 / math.Pi
)

func RotateVectorAngle[T Float](v Vector2[T], phi T) (r Vector2[T]) {
	return v.Rotate(phi / pi180)
}

// Rotate will rotate counterclockwise for positive phi (radian)
func (v Vector2[T]) Rotate(phi T) (r Vector2[T]) {
	r.X = v.X*Cos(phi) - v.Y*Sin(phi)
	r.Y = v.X*Sin(phi) + v.Y*Cos(phi)

	return r
}

func (v Vector2[T]) ToVector3(z T) Vector3[T] {
	return Vector3[T]{
		X: v.X,
		Y: v.Y,
		Z: z,
	}
}

// FindNormalVector returns a normalized normal vector which is perpendicular to the line segment given by the two vectors.
func FindNormalVector[T Number](p1, p2 Vector2[T]) Vector2[T] {
	p3, p4 := FindNormal(p1, p2)
	return p4.Sub(p3).Normalize()
}

// FindNormal returns a normal vector which is perpendicular to the line segment given by the two vectors as two vector endpoints.
func FindNormal[T Number](p1, p2 Vector2[T]) (Vector2[T], Vector2[T]) {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y

	return Vector2[T]{X: -dy, Y: dx}, Vector2[T]{X: dy, Y: -dx}
}

func GetNormal[T Number](v Vector2[T]) Vector2[T] {
	return NewVector2(v.Y, -v.X).Normalize()
}

// Three points are a counter-clockwise turn if ccw > 0, clockwise if
// ccw < 0, and collinear if ccw = 0 because ccw is a determinant that
// gives twice the signed  area of the triangle formed by p1, p2 and p3.
func CCW[T Number](p1, p2, p3 Vector2[T]) T {
	return (p2.X-p1.X)*(p3.Y-p1.Y) - (p2.Y-p1.Y)*(p3.X-p1.X)
}

func IsLeft[T Number](p1, p2, p3 Vector2[T]) T {
	return CCW(p1, p2, p3)
}

func CCWOrigin[T Number](p2, p3 Vector2[T]) T {
	return (p2.X-0)*(p3.Y-0) - (p2.Y-0)*(p3.X-0)
}

func (v Vector2[T]) String() string {
	return fmt.Sprintf("Vector2(%v, %v)", v.X, v.Y)
}

func (v Vector2[T]) Limit(limit T) Vector2[T] {
	l := v.Length()
	if l > limit {
		return v.Normalize().MultiplyScalar(limit)
	}

	return v
}

func (v Vector2[T]) DirectionTo(other Vector2[T]) Vector2[T] {
	return other.Sub(v).Normalize()
}

func (v Vector2[T]) IsZero() bool {
	return v.X == 0 && v.Y == 0
}

func (v Vector2[T]) ScaleToLength(length T) Vector2[T] {
	if length == 0 || v.IsZero() {
		return v
	}

	oldLength := v.Length()
	fraction := length / oldLength

	return v.MultiplyScalar(fraction)
}

func VectorAs[T1 Number, T2 Number](v Vector2[T1]) Vector2[T2] {
	return Vector2[T2]{
		X: T2(v.X),
		Y: T2(v.Y),
	}
}
