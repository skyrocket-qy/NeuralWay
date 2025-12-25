package render3d

import (
	"image/color"

	"github.com/mlange-42/ark/ecs"
)

// RenderSystem3D renders 3D entities with Transform3D and Mesh3D components.
type RenderSystem3D struct {
	transforms *ecs.Map1[Transform3D]
	meshes     *ecs.Map1[Mesh3D]
	filter     *ecs.Filter2[Transform3D, Mesh3D]
	world      *ecs.World
}

// NewRenderSystem3D creates a new 3D render system.
func NewRenderSystem3D() *RenderSystem3D {
	return &RenderSystem3D{}
}

// Init initializes the render system.
func (s *RenderSystem3D) Init(world *ecs.World) {
	s.world = world
	s.transforms = ecs.NewMap1[Transform3D](world)
	s.meshes = ecs.NewMap1[Mesh3D](world)
	s.filter = ecs.NewFilter2[Transform3D, Mesh3D](world)
}

// Update is called each frame (no logic needed for rendering).
func (s *RenderSystem3D) Update(world *ecs.World, dt float32) {
	// Render system doesn't need update logic
}

// Draw renders all entities with Transform3D and Mesh3D.
func (s *RenderSystem3D) Draw(world *ecs.World) {
	query := s.filter.Query()
	for query.Next() {
		entity := query.Entity()
		transform := s.transforms.Get(entity)
		mesh := s.meshes.Get(entity)

		if !mesh.Visible {
			continue
		}

		// Draw the mesh at the transform position
		// In real raylib-go implementation:
		// rl.DrawMesh(mesh, transform.GetMatrix())
		_ = transform
		_ = mesh
	}
}

// DrawMesh draws a single mesh (placeholder for raylib-go).
func DrawMesh(mesh Mesh3D, transform Transform3D, col color.Color) {
	// In real implementation with raylib-go:
	// matrix := transform.GetMatrix()
	// rl.DrawMeshInstanced(...)
	// For now, this is a placeholder
}

// DrawCube draws a cube at the given position.
func DrawCube(position Vector3, size float32, col color.Color) {
	// rl.DrawCube(rl.Vector3{...}, size, rl.Color{...})
}

// DrawSphere draws a sphere at the given position.
func DrawSphere(position Vector3, radius float32, col color.Color) {
	// rl.DrawSphere(rl.Vector3{...}, radius, rl.Color{...})
}

// DrawGrid draws a ground grid.
func DrawGrid(slices int, spacing float32) {
	// rl.DrawGrid(int32(slices), spacing)
}

// TransformSystem3D updates entity transforms based on velocity.
type TransformSystem3D struct {
	transforms *ecs.Map1[Transform3D]
	velocities *ecs.Map1[Velocity3D]
	filter     *ecs.Filter2[Transform3D, Velocity3D]
}

// Velocity3D represents 3D movement.
type Velocity3D struct {
	Linear  Vector3
	Angular Vector3 // Rotation speed
}

// NewTransformSystem3D creates a new transform system.
func NewTransformSystem3D() *TransformSystem3D {
	return &TransformSystem3D{}
}

// Init initializes the transform system.
func (s *TransformSystem3D) Init(world *ecs.World) {
	s.transforms = ecs.NewMap1[Transform3D](world)
	s.velocities = ecs.NewMap1[Velocity3D](world)
	s.filter = ecs.NewFilter2[Transform3D, Velocity3D](world)
}

// Update applies velocities to transforms.
func (s *TransformSystem3D) Update(world *ecs.World, dt float32) {
	query := s.filter.Query()
	for query.Next() {
		entity := query.Entity()
		transform := s.transforms.Get(entity)
		velocity := s.velocities.Get(entity)

		// Apply linear velocity
		transform.Position = transform.Position.Add(velocity.Linear.Scale(dt))

		// Apply angular velocity
		transform.Rotation = transform.Rotation.Add(velocity.Angular.Scale(dt))
	}
}

// Draw is not needed for transform system.
func (s *TransformSystem3D) Draw(world *ecs.World) {}

// CameraSystem3D handles camera updates.
type CameraSystem3D struct {
	camera *OrbitCamera
}

// NewCameraSystem3D creates a camera system.
func NewCameraSystem3D(camera *OrbitCamera) *CameraSystem3D {
	return &CameraSystem3D{camera: camera}
}

// Init initializes the camera system.
func (s *CameraSystem3D) Init(world *ecs.World) {}

// Update handles camera input.
func (s *CameraSystem3D) Update(world *ecs.World, dt float32) {
	// In real implementation, read input and update camera:
	// if rl.IsMouseButtonDown(rl.MouseButtonRight) {
	//     delta := rl.GetMouseDelta()
	//     s.camera.Rotate(delta.X, delta.Y)
	// }
	// wheel := rl.GetMouseWheelMove()
	// s.camera.Zoom(wheel)
}

// Draw is not needed for camera system.
func (s *CameraSystem3D) Draw(world *ecs.World) {}

// GetCamera returns the camera.
func (s *CameraSystem3D) GetCamera() *OrbitCamera {
	return s.camera
}
