package render3d

import (
	"github.com/mlange-42/ark/ecs"
)

// System3D is the interface for 3D ECS systems.
type System3D interface {
	// Init initializes the system with the world.
	Init(world *ecs.World)
	// Update updates the system logic.
	Update(world *ecs.World, dt float32)
	// Draw renders the system (called within 3D mode).
	Draw(world *ecs.World)
}

// Game3D is the main 3D game loop wrapper.
type Game3D struct {
	World    *ecs.World
	Systems  []System3D
	Camera   *OrbitCamera
	Title    string
	Width    int
	Height   int
	FPS      int
	Running  bool
	OnUpdate func(dt float32)
	OnDraw   func()
}

// NewGame3D creates a new 3D game.
func NewGame3D(title string, width, height int) *Game3D {
	world := ecs.NewWorld()

	return &Game3D{
		World:   &world,
		Systems: make([]System3D, 0),
		Camera:  NewOrbitCamera(20, Vector3{}),
		Title:   title,
		Width:   width,
		Height:  height,
		FPS:     60,
		Running: true,
	}
}

// AddSystem adds a system to the game.
func (g *Game3D) AddSystem(system System3D) {
	system.Init(g.World)
	g.Systems = append(g.Systems, system)
}

// GetCamera returns the camera for control.
func (g *Game3D) GetCamera() *OrbitCamera {
	return g.Camera
}

// SetCamera sets a custom camera.
func (g *Game3D) SetCamera(camera *OrbitCamera) {
	g.Camera = camera
}

// Update runs all system updates.
func (g *Game3D) Update(dt float32) {
	for _, sys := range g.Systems {
		sys.Update(g.World, dt)
	}

	if g.OnUpdate != nil {
		g.OnUpdate(dt)
	}
}

// Draw runs all system draws.
func (g *Game3D) Draw() {
	for _, sys := range g.Systems {
		sys.Draw(g.World)
	}

	if g.OnDraw != nil {
		g.OnDraw()
	}
}

// Run starts the game loop.
// Note: This is a placeholder. Actual raylib-go integration requires
// importing github.com/gen2brain/raylib-go/raylib and using:
//
//	rl.InitWindow, rl.BeginDrawing, rl.BeginMode3D, etc.
func (g *Game3D) Run() {
	// Initialize all systems
	for _, sys := range g.Systems {
		sys.Init(g.World)
	}

	// Placeholder game loop (in real impl, use raylib-go)
	// This allows the structure to compile without raylib-go dependency
	g.Running = true

	// Example raylib-go loop structure:
	// rl.InitWindow(int32(g.Width), int32(g.Height), g.Title)
	// rl.SetTargetFPS(int32(g.FPS))
	// for !rl.WindowShouldClose() && g.Running {
	//     dt := rl.GetFrameTime()
	//     g.Update(dt)
	//     rl.BeginDrawing()
	//     rl.ClearBackground(rl.RayWhite)
	//     rl.BeginMode3D(g.Camera.ToRaylibCamera())
	//     g.Draw()
	//     rl.EndMode3D()
	//     rl.EndDrawing()
	// }
	// rl.CloseWindow()
}

// Stop stops the game loop.
func (g *Game3D) Stop() {
	g.Running = false
}
