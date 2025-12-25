package render3d

import "image/color"

// Transform3D represents a 3D transformation.
type Transform3D struct {
	Position Vector3
	Rotation Vector3 // Euler angles in radians
	Scale    Vector3
}

// NewTransform3D creates a default transform.
func NewTransform3D() Transform3D {
	return Transform3D{
		Position: Vector3{},
		Rotation: Vector3{},
		Scale:    Vector3{X: 1, Y: 1, Z: 1},
	}
}

// GetMatrix returns the transformation matrix.
func (t Transform3D) GetMatrix() Matrix4 {
	translation := Translation(t.Position.X, t.Position.Y, t.Position.Z)
	rotation := RotationY(t.Rotation.Y) // Simplified: only Y rotation for now
	scale := Scaling(t.Scale.X, t.Scale.Y, t.Scale.Z)

	return translation.Multiply(rotation).Multiply(scale)
}

// Camera3D represents a 3D camera.
type Camera3D struct {
	Position Vector3
	Target   Vector3
	Up       Vector3
	FOV      float32 // Field of view in degrees
	Near     float32 // Near clipping plane
	Far      float32 // Far clipping plane
}

// NewCamera3D creates a default camera.
func NewCamera3D() Camera3D {
	return Camera3D{
		Position: Vector3{X: 0, Y: 10, Z: 10},
		Target:   Vector3{},
		Up:       Vector3{Y: 1},
		FOV:      45,
		Near:     0.1,
		Far:      1000,
	}
}

// LookAt sets the camera to look at a target.
func (c *Camera3D) LookAt(target Vector3) {
	c.Target = target
}

// Model3D represents a 3D model to render.
type Model3D struct {
	Path      string      // Path to model file (OBJ, glTF)
	Color     color.Color // Tint color
	Visible   bool
	Wireframe bool
}

// NewModel3D creates a new model component.
func NewModel3D(path string) Model3D {
	return Model3D{
		Path:    path,
		Color:   color.White,
		Visible: true,
	}
}

// Mesh3D represents a procedural mesh.
type Mesh3D struct {
	Vertices []Vector3
	Indices  []uint32
	Normals  []Vector3
	UVs      []Vector2
	Color    color.Color
	Visible  bool
}

// Vector2 is a 2D vector for UVs.
type Vector2 struct {
	X, Y float32
}

// Light3D represents a light source.
type Light3D struct {
	Type      LightType
	Position  Vector3
	Direction Vector3
	Color     color.Color
	Intensity float32
}

// LightType identifies the type of light.
type LightType int

const (
	LightDirectional LightType = iota
	LightPoint
	LightSpot
)

// NewDirectionalLight creates a directional light.
func NewDirectionalLight(direction Vector3, intensity float32) Light3D {
	return Light3D{
		Type:      LightDirectional,
		Direction: direction.Normalize(),
		Color:     color.White,
		Intensity: intensity,
	}
}

// NewPointLight creates a point light.
func NewPointLight(position Vector3, intensity float32) Light3D {
	return Light3D{
		Type:      LightPoint,
		Position:  position,
		Color:     color.White,
		Intensity: intensity,
	}
}

// Material3D represents a surface material.
type Material3D struct {
	Diffuse   color.Color
	Specular  color.Color
	Shininess float32
	Texture   string // Path to texture
}

// NewMaterial3D creates a default material.
func NewMaterial3D() Material3D {
	return Material3D{
		Diffuse:   color.White,
		Specular:  color.White,
		Shininess: 32,
	}
}
