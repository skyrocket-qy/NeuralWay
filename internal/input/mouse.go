package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// MouseState generic mouse state tracker.
type MouseState struct {
	X, Y           int
	PrevX, PrevY   int
	DeltaX, DeltaY int

	// Button state
	LeftPressed      bool
	RightPressed     bool
	MiddlePressed    bool
	LeftJustPressed  bool
	RightJustPressed bool

	// Wheel
	WheelX, WheelY float64

	// Drag state
	IsDragging bool
	DragStartX int
	DragStartY int

	// Drag threshold (pixels squared)
	DragThreshold float64
}

// NewMouseState creates a new mouse state tracker.
func NewMouseState() *MouseState {
	return &MouseState{
		DragThreshold: 25, // 5 pixels default
	}
}

// Update updates the mouse state. Should be called once per frame.
func (m *MouseState) Update() {
	m.PrevX, m.PrevY = m.X, m.Y
	m.X, m.Y = ebiten.CursorPosition()
	m.DeltaX = m.X - m.PrevX
	m.DeltaY = m.Y - m.PrevY

	// Button state
	m.LeftPressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	m.RightPressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	m.MiddlePressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle)
	m.LeftJustPressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	m.RightJustPressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)

	// Wheel
	m.WheelX, m.WheelY = ebiten.Wheel()

	// Drag detection
	if m.LeftJustPressed {
		m.DragStartX = m.X
		m.DragStartY = m.Y
	}

	if m.LeftPressed {
		dx := m.X - m.DragStartX

		dy := m.Y - m.DragStartY
		if float64(dx*dx+dy*dy) > m.DragThreshold {
			m.IsDragging = true
		}
	} else {
		m.IsDragging = false
	}
}

// IsInRect checks if mouse is in a rectangle.
func (m *MouseState) IsInRect(x, y, w, h float64) bool {
	return float64(m.X) >= x && float64(m.X) <= x+w &&
		float64(m.Y) >= y && float64(m.Y) <= y+h
}

// IsNear checks if mouse is near a point.
func (m *MouseState) IsNear(x, y, radius float64) bool {
	dx := float64(m.X) - x
	dy := float64(m.Y) - y

	return dx*dx+dy*dy <= radius*radius
}
