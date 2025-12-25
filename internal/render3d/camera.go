package render3d

import "math"

// OrbitCamera provides an orbiting camera controller.
type OrbitCamera struct {
	Distance float32 // Distance from target
	Yaw      float32 // Horizontal angle (radians)
	Pitch    float32 // Vertical angle (radians)
	Target   Vector3 // Look-at target

	// Constraints
	MinDistance float32
	MaxDistance float32
	MinPitch    float32 // Minimum vertical angle
	MaxPitch    float32 // Maximum vertical angle

	// Sensitivity
	RotateSensitivity float32
	ZoomSensitivity   float32
}

// NewOrbitCamera creates a new orbit camera.
func NewOrbitCamera(distance float32, target Vector3) *OrbitCamera {
	return &OrbitCamera{
		Distance:          distance,
		Target:            target,
		MinDistance:       1,
		MaxDistance:       100,
		MinPitch:          -math.Pi/2 + 0.1,
		MaxPitch:          math.Pi/2 - 0.1,
		RotateSensitivity: 0.01,
		ZoomSensitivity:   0.1,
	}
}

// Rotate rotates the camera by delta yaw and pitch.
func (c *OrbitCamera) Rotate(deltaYaw, deltaPitch float32) {
	c.Yaw += deltaYaw * c.RotateSensitivity
	c.Pitch += deltaPitch * c.RotateSensitivity

	// Clamp pitch
	if c.Pitch < c.MinPitch {
		c.Pitch = c.MinPitch
	}

	if c.Pitch > c.MaxPitch {
		c.Pitch = c.MaxPitch
	}
}

// Zoom changes the distance from target.
func (c *OrbitCamera) Zoom(delta float32) {
	c.Distance -= delta * c.ZoomSensitivity

	// Clamp distance
	if c.Distance < c.MinDistance {
		c.Distance = c.MinDistance
	}

	if c.Distance > c.MaxDistance {
		c.Distance = c.MaxDistance
	}
}

// GetPosition returns the camera world position.
func (c *OrbitCamera) GetPosition() Vector3 {
	// Convert spherical to cartesian coordinates
	x := c.Distance * float32(math.Cos(float64(c.Pitch))) * float32(math.Sin(float64(c.Yaw)))
	y := c.Distance * float32(math.Sin(float64(c.Pitch)))
	z := c.Distance * float32(math.Cos(float64(c.Pitch))) * float32(math.Cos(float64(c.Yaw)))

	return c.Target.Add(Vector3{X: x, Y: y, Z: z})
}

// ToCamera3D converts to a Camera3D component.
func (c *OrbitCamera) ToCamera3D() Camera3D {
	return Camera3D{
		Position: c.GetPosition(),
		Target:   c.Target,
		Up:       Vector3{Y: 1},
		FOV:      45,
		Near:     0.1,
		Far:      1000,
	}
}

// FPSCamera provides a first-person camera controller.
type FPSCamera struct {
	Position    Vector3
	Yaw         float32 // Horizontal direction
	Pitch       float32 // Vertical direction
	MoveSpeed   float32
	Sensitivity float32
}

// NewFPSCamera creates a new FPS camera.
func NewFPSCamera(position Vector3) *FPSCamera {
	return &FPSCamera{
		Position:    position,
		MoveSpeed:   10,
		Sensitivity: 0.002,
	}
}

// Look rotates the camera view.
func (c *FPSCamera) Look(deltaYaw, deltaPitch float32) {
	c.Yaw += deltaYaw * c.Sensitivity
	c.Pitch += deltaPitch * c.Sensitivity

	// Clamp pitch to avoid flipping
	maxPitch := float32(math.Pi/2 - 0.1)
	if c.Pitch > maxPitch {
		c.Pitch = maxPitch
	}

	if c.Pitch < -maxPitch {
		c.Pitch = -maxPitch
	}
}

// Move moves the camera in the given direction.
func (c *FPSCamera) Move(forward, right, up, dt float32) {
	// Calculate forward vector
	fwd := Vector3{
		X: float32(math.Sin(float64(c.Yaw))),
		Y: 0,
		Z: float32(math.Cos(float64(c.Yaw))),
	}
	// Calculate right vector
	rgt := Vector3{
		X: float32(math.Cos(float64(c.Yaw))),
		Y: 0,
		Z: -float32(math.Sin(float64(c.Yaw))),
	}

	velocity := fwd.Scale(forward).Add(rgt.Scale(right)).Add(Vector3{Y: up})
	c.Position = c.Position.Add(velocity.Scale(c.MoveSpeed * dt))
}

// GetTarget returns the look-at target.
func (c *FPSCamera) GetTarget() Vector3 {
	forward := Vector3{
		X: float32(math.Cos(float64(c.Pitch)) * math.Sin(float64(c.Yaw))),
		Y: float32(math.Sin(float64(c.Pitch))),
		Z: float32(math.Cos(float64(c.Pitch)) * math.Cos(float64(c.Yaw))),
	}

	return c.Position.Add(forward)
}

// ToCamera3D converts to a Camera3D component.
func (c *FPSCamera) ToCamera3D() Camera3D {
	return Camera3D{
		Position: c.Position,
		Target:   c.GetTarget(),
		Up:       Vector3{Y: 1},
		FOV:      60,
		Near:     0.1,
		Far:      1000,
	}
}
