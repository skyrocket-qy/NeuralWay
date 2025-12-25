package render3d

import "math"

// Vector3 represents a 3D vector.
type Vector3 struct {
	X, Y, Z float32
}

// NewVector3 creates a new 3D vector.
func NewVector3(x, y, z float32) Vector3 {
	return Vector3{X: x, Y: y, Z: z}
}

// Zero returns a zero vector.
func (v Vector3) Zero() Vector3 {
	return Vector3{}
}

// One returns a unit vector.
func (v Vector3) One() Vector3 {
	return Vector3{X: 1, Y: 1, Z: 1}
}

// Add adds two vectors.
func (v Vector3) Add(other Vector3) Vector3 {
	return Vector3{X: v.X + other.X, Y: v.Y + other.Y, Z: v.Z + other.Z}
}

// Sub subtracts two vectors.
func (v Vector3) Sub(other Vector3) Vector3 {
	return Vector3{X: v.X - other.X, Y: v.Y - other.Y, Z: v.Z - other.Z}
}

// Scale multiplies the vector by a scalar.
func (v Vector3) Scale(s float32) Vector3 {
	return Vector3{X: v.X * s, Y: v.Y * s, Z: v.Z * s}
}

// Length returns the magnitude of the vector.
func (v Vector3) Length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
}

// Normalize returns a unit vector.
func (v Vector3) Normalize() Vector3 {
	length := v.Length()
	if length == 0 {
		return v
	}

	return v.Scale(1 / length)
}

// Dot returns the dot product.
func (v Vector3) Dot(other Vector3) float32 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Cross returns the cross product.
func (v Vector3) Cross(other Vector3) Vector3 {
	return Vector3{
		X: v.Y*other.Z - v.Z*other.Y,
		Y: v.Z*other.X - v.X*other.Z,
		Z: v.X*other.Y - v.Y*other.X,
	}
}

// Distance returns the distance between two vectors.
func (v Vector3) Distance(other Vector3) float32 {
	return v.Sub(other).Length()
}

// Lerp linearly interpolates between two vectors.
func (v Vector3) Lerp(other Vector3, t float32) Vector3 {
	return Vector3{
		X: v.X + (other.X-v.X)*t,
		Y: v.Y + (other.Y-v.Y)*t,
		Z: v.Z + (other.Z-v.Z)*t,
	}
}

// Matrix4 represents a 4x4 transformation matrix.
type Matrix4 [16]float32

// Identity returns an identity matrix.
func Identity() Matrix4 {
	return Matrix4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// Translation creates a translation matrix.
func Translation(x, y, z float32) Matrix4 {
	m := Identity()
	m[12] = x
	m[13] = y
	m[14] = z

	return m
}

// Scaling creates a scaling matrix.
func Scaling(x, y, z float32) Matrix4 {
	return Matrix4{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

// RotationY creates a rotation matrix around the Y axis.
func RotationY(angle float32) Matrix4 {
	c := float32(math.Cos(float64(angle)))
	s := float32(math.Sin(float64(angle)))

	return Matrix4{
		c, 0, s, 0,
		0, 1, 0, 0,
		-s, 0, c, 0,
		0, 0, 0, 1,
	}
}

// Multiply multiplies two matrices.
func (m Matrix4) Multiply(other Matrix4) Matrix4 {
	var result Matrix4

	for i := range 4 {
		for j := range 4 {
			result[i*4+j] = m[i*4+0]*other[0*4+j] +
				m[i*4+1]*other[1*4+j] +
				m[i*4+2]*other[2*4+j] +
				m[i*4+3]*other[3*4+j]
		}
	}

	return result
}

// TransformVector applies the matrix to a vector.
func (m Matrix4) TransformVector(v Vector3) Vector3 {
	return Vector3{
		X: m[0]*v.X + m[4]*v.Y + m[8]*v.Z + m[12],
		Y: m[1]*v.X + m[5]*v.Y + m[9]*v.Z + m[13],
		Z: m[2]*v.X + m[6]*v.Y + m[10]*v.Z + m[14],
	}
}

// Quaternion represents a rotation quaternion.
type Quaternion struct {
	X, Y, Z, W float32
}

// QuaternionIdentity returns an identity quaternion.
func QuaternionIdentity() Quaternion {
	return Quaternion{W: 1}
}

// FromEuler creates a quaternion from Euler angles (in radians).
func FromEuler(pitch, yaw, roll float32) Quaternion {
	cy := float32(math.Cos(float64(yaw * 0.5)))
	sy := float32(math.Sin(float64(yaw * 0.5)))
	cp := float32(math.Cos(float64(pitch * 0.5)))
	sp := float32(math.Sin(float64(pitch * 0.5)))
	cr := float32(math.Cos(float64(roll * 0.5)))
	sr := float32(math.Sin(float64(roll * 0.5)))

	return Quaternion{
		W: cr*cp*cy + sr*sp*sy,
		X: sr*cp*cy - cr*sp*sy,
		Y: cr*sp*cy + sr*cp*sy,
		Z: cr*cp*sy - sr*sp*cy,
	}
}
