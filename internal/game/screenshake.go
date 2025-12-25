package game

import (
	"math/rand"
)

// ScreenShake provides standalone screen shake functionality.
// This is a simpler alternative to the camera's built-in shake.
type ScreenShake struct {
	Trauma      float64 // Current trauma amount (0-1)
	TraumaDecay float64 // How fast trauma decays per second
	MaxOffsetX  float64 // Maximum X offset
	MaxOffsetY  float64 // Maximum Y offset
	MaxRotation float64 // Maximum rotation in radians

	// Output values (read these for rendering)
	OffsetX  float64
	OffsetY  float64
	Rotation float64

	// Internal
	time      float64
	frequency float64
}

// NewScreenShake creates a screen shake controller.
func NewScreenShake() *ScreenShake {
	return &ScreenShake{
		TraumaDecay: 1.0,
		MaxOffsetX:  15,
		MaxOffsetY:  15,
		MaxRotation: 0.05,
		frequency:   30,
	}
}

// AddTrauma adds trauma (clamped to 0-1).
func (s *ScreenShake) AddTrauma(amount float64) {
	s.Trauma += amount
	if s.Trauma > 1 {
		s.Trauma = 1
	}

	if s.Trauma < 0 {
		s.Trauma = 0
	}
}

// Shake is a convenience method to add trauma.
func (s *ScreenShake) Shake(intensity float64) {
	s.AddTrauma(intensity)
}

// ShakeSmall adds a small shake (0.2 trauma).
func (s *ScreenShake) ShakeSmall() {
	s.AddTrauma(0.2)
}

// ShakeMedium adds a medium shake (0.5 trauma).
func (s *ScreenShake) ShakeMedium() {
	s.AddTrauma(0.5)
}

// ShakeLarge adds a large shake (0.8 trauma).
func (s *ScreenShake) ShakeLarge() {
	s.AddTrauma(0.8)
}

// ShakeMax adds maximum shake (1.0 trauma).
func (s *ScreenShake) ShakeMax() {
	s.AddTrauma(1.0)
}

// Update updates the shake effect.
func (s *ScreenShake) Update(dt float64) {
	if s.Trauma <= 0 {
		s.OffsetX = 0
		s.OffsetY = 0
		s.Rotation = 0

		return
	}

	// Decay trauma
	s.Trauma -= s.TraumaDecay * dt
	if s.Trauma < 0 {
		s.Trauma = 0
	}

	// Calculate shake amount (trauma^2 for smoother feel)
	shake := s.Trauma * s.Trauma

	s.time += dt * s.frequency

	// Use noise-like random for each axis
	// In a real implementation, you'd use Perlin noise
	s.OffsetX = s.MaxOffsetX * shake * (rand.Float64()*2 - 1)
	s.OffsetY = s.MaxOffsetY * shake * (rand.Float64()*2 - 1)
	s.Rotation = s.MaxRotation * shake * (rand.Float64()*2 - 1)
}

// IsActive returns true if shake is currently active.
func (s *ScreenShake) IsActive() bool {
	return s.Trauma > 0.01
}

// Reset clears all shake.
func (s *ScreenShake) Reset() {
	s.Trauma = 0
	s.OffsetX = 0
	s.OffsetY = 0
	s.Rotation = 0
}

// GetOffset returns current X, Y offset.
func (s *ScreenShake) GetOffset() (x, y float64) {
	return s.OffsetX, s.OffsetY
}

// GetTransform returns offset and rotation for rendering.
func (s *ScreenShake) GetTransform() (x, y, rotation float64) {
	return s.OffsetX, s.OffsetY, s.Rotation
}

// SetDecay sets how fast trauma decays.
func (s *ScreenShake) SetDecay(decay float64) {
	s.TraumaDecay = decay
}

// SetMaxOffset sets maximum shake offset.
func (s *ScreenShake) SetMaxOffset(x, y float64) {
	s.MaxOffsetX = x
	s.MaxOffsetY = y
}

// SetFrequency sets the shake frequency.
func (s *ScreenShake) SetFrequency(freq float64) {
	s.frequency = freq
}

// PresetHit applies a hit/impact shake preset.
func (s *ScreenShake) PresetHit() {
	s.MaxOffsetX = 8
	s.MaxOffsetY = 8
	s.MaxRotation = 0.02
	s.TraumaDecay = 2.0
	s.Shake(0.4)
}

// PresetExplosion applies an explosion shake preset.
func (s *ScreenShake) PresetExplosion() {
	s.MaxOffsetX = 20
	s.MaxOffsetY = 20
	s.MaxRotation = 0.08
	s.TraumaDecay = 0.8
	s.Shake(0.9)
}

// PresetRumble applies a continuous rumble preset.
func (s *ScreenShake) PresetRumble() {
	s.MaxOffsetX = 3
	s.MaxOffsetY = 3
	s.MaxRotation = 0.01
	s.TraumaDecay = 0.5
	s.Shake(0.3)
}

// PresetCritical applies a critical hit shake preset.
func (s *ScreenShake) PresetCritical() {
	s.MaxOffsetX = 12
	s.MaxOffsetY = 12
	s.MaxRotation = 0.04
	s.TraumaDecay = 1.5
	s.Shake(0.6)
}
