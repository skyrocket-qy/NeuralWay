// Package debug provides development tools for inspecting ECS entities and components.
package debug

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// Inspector provides an on-screen debug overlay for inspecting ECS entities.
// Toggle with F12 key. Shows entity counts, component breakdown, and position markers.
type Inspector struct {
	world   *ecs.World
	enabled bool

	// Filters for counting different entity types
	posFilter      *ecs.Filter1[components.Position]
	spriteFilter   *ecs.Filter2[components.Position, components.Sprite]
	velocityFilter *ecs.Filter2[components.Position, components.Velocity]
	healthFilter   *ecs.Filter1[components.Health]
	colliderFilter *ecs.Filter1[components.Collider]
	tilemapFilter  *ecs.Filter1[components.Tilemap]

	// Camera offset for world-to-screen conversion
	cameraX, cameraY float64

	// Stats cache (updated each frame when enabled)
	totalEntities int
	positionCount int
	spriteCount   int
	velocityCount int
	healthCount   int
	colliderCount int
	tilemapCount  int
}

// NewInspector creates a new debug inspector for the given ECS world.
func NewInspector(world *ecs.World) *Inspector {
	return &Inspector{
		world:          world,
		enabled:        false,
		posFilter:      ecs.NewFilter1[components.Position](world),
		spriteFilter:   ecs.NewFilter2[components.Position, components.Sprite](world),
		velocityFilter: ecs.NewFilter2[components.Position, components.Velocity](world),
		healthFilter:   ecs.NewFilter1[components.Health](world),
		colliderFilter: ecs.NewFilter1[components.Collider](world),
		tilemapFilter:  ecs.NewFilter1[components.Tilemap](world),
	}
}

// SetCamera sets the camera offset for world-to-screen position conversion.
func (i *Inspector) SetCamera(x, y float64) {
	i.cameraX = x
	i.cameraY = y
}

// Toggle enables or disables the debug overlay.
func (i *Inspector) Toggle() {
	i.enabled = !i.enabled
}

// Enabled returns whether the inspector overlay is currently visible.
func (i *Inspector) Enabled() bool {
	return i.enabled
}

// Update checks for F12 key and updates stats when enabled.
func (i *Inspector) Update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		i.Toggle()
	}

	if !i.enabled {
		return
	}

	// Update stats
	i.updateStats()
}

// updateStats counts entities for each component type.
func (i *Inspector) updateStats() {
	i.positionCount = 0
	i.spriteCount = 0
	i.velocityCount = 0
	i.healthCount = 0
	i.colliderCount = 0
	i.tilemapCount = 0

	// Count positions (total renderable entities)
	query := i.posFilter.Query()
	for query.Next() {
		i.positionCount++
	}

	// Count sprites
	query2 := i.spriteFilter.Query()
	for query2.Next() {
		i.spriteCount++
	}

	// Count velocities (moving entities)
	query3 := i.velocityFilter.Query()
	for query3.Next() {
		i.velocityCount++
	}

	// Count health
	query4 := i.healthFilter.Query()
	for query4.Next() {
		i.healthCount++
	}

	// Count colliders
	query5 := i.colliderFilter.Query()
	for query5.Next() {
		i.colliderCount++
	}

	// Count tilemaps
	query6 := i.tilemapFilter.Query()
	for query6.Next() {
		i.tilemapCount++
	}

	// Total is max of position count (most common component)
	i.totalEntities = max(i.spriteCount, i.positionCount)
}

// Draw renders the debug overlay on screen.
func (i *Inspector) Draw(screen *ebiten.Image) {
	if !i.enabled {
		return
	}

	// Draw background panel
	panelW, panelH := float32(200), float32(160)
	panelX, panelY := float32(screen.Bounds().Dx())-panelW-10, float32(10)

	// Semi-transparent background
	vector.FillRect(screen, panelX, panelY, panelW, panelH, color.RGBA{R: 0, G: 0, B: 0, A: 200}, false)
	vector.StrokeRect(
		screen,
		panelX,
		panelY,
		panelW,
		panelH,
		1,
		color.RGBA{R: 100, G: 255, B: 100, A: 255},
		false,
	)

	// Title
	ebitenutil.DebugPrintAt(screen, "=== DEBUG INSPECTOR ===", int(panelX)+10, int(panelY)+5)

	// Stats
	y := int(panelY) + 25
	lineHeight := 16

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Entities: %d", i.totalEntities), int(panelX)+10, y)
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Position: %d", i.positionCount), int(panelX)+10, y)
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Sprite: %d", i.spriteCount), int(panelX)+10, y)
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Velocity: %d", i.velocityCount), int(panelX)+10, y)
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Health: %d", i.healthCount), int(panelX)+10, y)
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Collider: %d", i.colliderCount), int(panelX)+10, y)
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Tilemap: %d", i.tilemapCount), int(panelX)+10, y)

	// Footer
	ebitenutil.DebugPrintAt(screen, "Press F12 to close", int(panelX)+10, int(panelY)+int(panelH)-18)

	// Draw entity position markers
	i.drawPositionMarkers(screen)
}

// drawPositionMarkers draws small dots at each entity position.
func (i *Inspector) drawPositionMarkers(screen *ebiten.Image) {
	markerColor := color.RGBA{R: 255, G: 255, B: 0, A: 180}

	query := i.posFilter.Query()
	for query.Next() {
		pos := query.Get()

		// Convert world position to screen position
		screenX := float32(pos.X - i.cameraX)
		screenY := float32(pos.Y - i.cameraY)

		// Skip if off-screen
		if screenX < 0 || screenX > float32(screen.Bounds().Dx()) ||
			screenY < 0 || screenY > float32(screen.Bounds().Dy()) {
			continue
		}

		// Draw small marker
		vector.FillCircle(screen, screenX, screenY, 3, markerColor, false)
	}
}
