package game

import (
	"math"
	"math/rand"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// Camera represents a 2D camera with follow, zoom, and shake.
type Camera struct {
	X, Y         float64 // Camera position (center)
	TargetX      float64 // Follow target X
	TargetY      float64 // Follow target Y
	Zoom         float64 // Zoom level (1.0 = normal)
	MinZoom      float64
	MaxZoom      float64
	ScreenWidth  float64
	ScreenHeight float64

	// Smooth follow
	Smoothing float64 // 0 = instant, 1 = very slow
	Deadzone  float64 // Area where camera doesn't move

	// Bounds (0 = unlimited)
	MinX, MaxX float64
	MinY, MaxY float64

	// Shake
	TraumaAmount float64 // Current trauma (0-1)
	TraumaDecay  float64 // Trauma decay per second
	ShakeOffset  struct{ X, Y float64 }
	MaxShakeX    float64 // Max shake offset
	MaxShakeY    float64
	ShakeFreq    float64 // Shake oscillation frequency
	shakeTime    float64
}

// NewCamera creates a camera with defaults.
func NewCamera(screenWidth, screenHeight float64) *Camera {
	return &Camera{
		Zoom:         1.0,
		MinZoom:      0.25,
		MaxZoom:      4.0,
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
		Smoothing:    0.1,
		TraumaDecay:  1.0,
		MaxShakeX:    10,
		MaxShakeY:    10,
		ShakeFreq:    30,
	}
}

// SetTarget sets the follow target position.
func (c *Camera) SetTarget(x, y float64) {
	c.TargetX = x
	c.TargetY = y
}

// FollowEntity follows an entity with a Position component.
func (c *Camera) FollowEntity(world *ecs.World, entity ecs.Entity, filter *ecs.Filter1[components.Position]) {
	query := filter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			pos := query.Get()
			c.SetTarget(pos.X, pos.Y)

			return
		}
	}
}

// Update updates the camera position and effects.
func (c *Camera) Update(dt float64) {
	// Smooth follow
	if c.Smoothing > 0 {
		lerpFactor := 1.0 - math.Pow(c.Smoothing, dt*60)

		// Check deadzone
		dx := c.TargetX - c.X
		dy := c.TargetY - c.Y

		if math.Abs(dx) > c.Deadzone {
			c.X += dx * lerpFactor
		}

		if math.Abs(dy) > c.Deadzone {
			c.Y += dy * lerpFactor
		}
	} else {
		c.X = c.TargetX
		c.Y = c.TargetY
	}

	// Apply bounds
	c.clampToBounds()

	// Update shake
	c.updateShake(dt)
}

// clampToBounds keeps camera within bounds.
func (c *Camera) clampToBounds() {
	halfW := c.ScreenWidth / 2 / c.Zoom
	halfH := c.ScreenHeight / 2 / c.Zoom

	if c.MinX != 0 || c.MaxX != 0 {
		if c.X-halfW < c.MinX {
			c.X = c.MinX + halfW
		}

		if c.X+halfW > c.MaxX {
			c.X = c.MaxX - halfW
		}
	}

	if c.MinY != 0 || c.MaxY != 0 {
		if c.Y-halfH < c.MinY {
			c.Y = c.MinY + halfH
		}

		if c.Y+halfH > c.MaxY {
			c.Y = c.MaxY - halfH
		}
	}
}

// updateShake applies screen shake.
func (c *Camera) updateShake(dt float64) {
	if c.TraumaAmount <= 0 {
		c.ShakeOffset.X = 0
		c.ShakeOffset.Y = 0

		return
	}

	// Decay trauma
	c.TraumaAmount -= c.TraumaDecay * dt
	if c.TraumaAmount < 0 {
		c.TraumaAmount = 0
	}

	// Calculate shake (trauma^2 for better feel)
	shake := c.TraumaAmount * c.TraumaAmount
	c.shakeTime += dt * c.ShakeFreq

	// Use perlin-like noise (simplified)
	c.ShakeOffset.X = c.MaxShakeX * shake * (rand.Float64()*2 - 1)
	c.ShakeOffset.Y = c.MaxShakeY * shake * (rand.Float64()*2 - 1)
}

// AddTrauma adds to the current trauma (clamped to 0-1).
func (c *Camera) AddTrauma(amount float64) {
	c.TraumaAmount += amount
	if c.TraumaAmount > 1 {
		c.TraumaAmount = 1
	}
}

// Shake adds trauma (convenience method).
func (c *Camera) Shake(intensity float64) {
	c.AddTrauma(intensity)
}

// SetZoom sets the zoom level (clamped to min/max).
func (c *Camera) SetZoom(zoom float64) {
	c.Zoom = math.Max(c.MinZoom, math.Min(c.MaxZoom, zoom))
}

// ZoomIn increases zoom.
func (c *Camera) ZoomIn(amount float64) {
	c.SetZoom(c.Zoom + amount)
}

// ZoomOut decreases zoom.
func (c *Camera) ZoomOut(amount float64) {
	c.SetZoom(c.Zoom - amount)
}

// SetBounds sets the camera movement bounds.
func (c *Camera) SetBounds(minX, minY, maxX, maxY float64) {
	c.MinX = minX
	c.MinY = minY
	c.MaxX = maxX
	c.MaxY = maxY
}

// GetViewPosition returns the final camera position including shake.
func (c *Camera) GetViewPosition() (x, y float64) {
	return c.X + c.ShakeOffset.X, c.Y + c.ShakeOffset.Y
}

// WorldToScreen converts world coordinates to screen coordinates.
func (c *Camera) WorldToScreen(worldX, worldY float64) (screenX, screenY float64) {
	viewX, viewY := c.GetViewPosition()
	screenX = (worldX-viewX)*c.Zoom + c.ScreenWidth/2
	screenY = (worldY-viewY)*c.Zoom + c.ScreenHeight/2

	return screenX, screenY
}

// ScreenToWorld converts screen coordinates to world coordinates.
func (c *Camera) ScreenToWorld(screenX, screenY float64) (worldX, worldY float64) {
	viewX, viewY := c.GetViewPosition()
	worldX = (screenX-c.ScreenWidth/2)/c.Zoom + viewX
	worldY = (screenY-c.ScreenHeight/2)/c.Zoom + viewY

	return worldX, worldY
}

// GetViewBounds returns the visible world area.
func (c *Camera) GetViewBounds() (minX, minY, maxX, maxY float64) {
	halfW := c.ScreenWidth / 2 / c.Zoom
	halfH := c.ScreenHeight / 2 / c.Zoom
	viewX, viewY := c.GetViewPosition()

	return viewX - halfW, viewY - halfH, viewX + halfW, viewY + halfH
}

// IsVisible returns true if a point is visible.
func (c *Camera) IsVisible(x, y float64) bool {
	minX, minY, maxX, maxY := c.GetViewBounds()

	return x >= minX && x <= maxX && y >= minY && y <= maxY
}

// IsRectVisible returns true if any part of a rectangle is visible.
func (c *Camera) IsRectVisible(x, y, width, height float64) bool {
	minX, minY, maxX, maxY := c.GetViewBounds()

	return x+width >= minX && x <= maxX && y+height >= minY && y <= maxY
}

// LookAt immediately centers the camera on a position.
func (c *Camera) LookAt(x, y float64) {
	c.X = x
	c.Y = y
	c.TargetX = x
	c.TargetY = y
}

// Reset resets camera to default state.
func (c *Camera) Reset() {
	c.X = 0
	c.Y = 0
	c.TargetX = 0
	c.TargetY = 0
	c.Zoom = 1.0
	c.TraumaAmount = 0
	c.ShakeOffset.X = 0
	c.ShakeOffset.Y = 0
}
