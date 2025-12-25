package components

// ZIndex provides manual depth control for 2.5D rendering.
type ZIndex struct {
	Value int // Higher values render on top
}

// Elevation represents vertical height in 2.5D space.
type Elevation struct {
	Height   float64 // Current height above ground
	Ground   float64 // Ground level Y position
	Velocity float64 // Vertical velocity (for jumping)
	Gravity  float64 // Gravity strength
}

// Shadow component for rendering blob shadows.
type Shadow struct {
	Enabled bool
	OffsetX float64
	OffsetY float64
	ScaleX  float64
	ScaleY  float64
	Opacity float64 // 0.0 to 1.0
}

// NewShadow creates a default shadow.
func NewShadow() Shadow {
	return Shadow{
		Enabled: true,
		OffsetX: 0,
		OffsetY: 8,
		ScaleX:  1.0,
		ScaleY:  0.5,
		Opacity: 0.3,
	}
}
