package engine

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// TouchState represents the state of a single touch.
type TouchState struct {
	ID                 ebiten.TouchID
	StartX, StartY     int
	CurrentX, CurrentY int
	StartTime          time.Time
	IsActive           bool
}

// TouchManager handles multi-touch input.
type TouchManager struct {
	touches     map[ebiten.TouchID]*TouchState
	touchIDs    []ebiten.TouchID
	prevTouches map[ebiten.TouchID]*TouchState
}

// NewTouchManager creates a new touch manager.
func NewTouchManager() *TouchManager {
	return &TouchManager{
		touches:     make(map[ebiten.TouchID]*TouchState),
		touchIDs:    make([]ebiten.TouchID, 0),
		prevTouches: make(map[ebiten.TouchID]*TouchState),
	}
}

// Update should be called each frame to update touch states.
func (t *TouchManager) Update() {
	// Store previous states
	t.prevTouches = make(map[ebiten.TouchID]*TouchState)
	for id, state := range t.touches {
		stateCopy := *state
		t.prevTouches[id] = &stateCopy
	}

	// Get current touches
	t.touchIDs = inpututil.AppendJustPressedTouchIDs(t.touchIDs[:0])

	// Add new touches
	for _, id := range t.touchIDs {
		x, y := ebiten.TouchPosition(id)
		t.touches[id] = &TouchState{
			ID:        id,
			StartX:    x,
			StartY:    y,
			CurrentX:  x,
			CurrentY:  y,
			StartTime: time.Now(),
			IsActive:  true,
		}
	}

	// Update existing touches
	for id, state := range t.touches {
		if state.IsActive {
			x, y := ebiten.TouchPosition(id)
			if x == 0 && y == 0 && state.CurrentX != 0 {
				// Touch ended
				state.IsActive = false
			} else {
				state.CurrentX = x
				state.CurrentY = y
			}
		}
	}

	// Clean up ended touches
	for id, state := range t.touches {
		if !state.IsActive && time.Since(state.StartTime) > 500*time.Millisecond {
			delete(t.touches, id)
		}
	}
}

// IsTouching returns true if there are active touches.
func (t *TouchManager) IsTouching() bool {
	for _, state := range t.touches {
		if state.IsActive {
			return true
		}
	}

	return false
}

// TouchCount returns the number of active touches.
func (t *TouchManager) TouchCount() int {
	count := 0

	for _, state := range t.touches {
		if state.IsActive {
			count++
		}
	}

	return count
}

// GetTouch returns a specific touch state.
func (t *TouchManager) GetTouch(id ebiten.TouchID) *TouchState {
	return t.touches[id]
}

// GetAllTouches returns all active touch states.
func (t *TouchManager) GetAllTouches() []*TouchState {
	result := make([]*TouchState, 0)

	for _, state := range t.touches {
		if state.IsActive {
			result = append(result, state)
		}
	}

	return result
}

// GetFirstTouch returns the first active touch, or nil if none.
func (t *TouchManager) GetFirstTouch() *TouchState {
	for _, state := range t.touches {
		if state.IsActive {
			return state
		}
	}

	return nil
}

// GestureType identifies the type of gesture.
type GestureType int

const (
	GestureTap GestureType = iota
	GestureDoubleTap
	GestureSwipeLeft
	GestureSwipeRight
	GestureSwipeUp
	GestureSwipeDown
	GesturePinch
	GestureDrag
)

// Gesture represents a detected gesture.
type Gesture struct {
	Type     GestureType
	X, Y     int     // Position (center or start)
	DX, DY   float64 // Direction or delta
	Scale    float64 // For pinch (1.0 = no change)
	Velocity float64 // Movement speed
}

// GestureRecognizer detects gestures from touch input.
type GestureRecognizer struct {
	touchMgr        *TouchManager
	handlers        map[GestureType]func(Gesture)
	lastTapTime     time.Time
	lastTapX        int
	lastTapY        int
	swipeThreshold  float64
	tapMaxDuration  time.Duration
	doubleTapWindow time.Duration
}

// NewGestureRecognizer creates a new gesture recognizer.
func NewGestureRecognizer(touchMgr *TouchManager) *GestureRecognizer {
	return &GestureRecognizer{
		touchMgr:        touchMgr,
		handlers:        make(map[GestureType]func(Gesture)),
		swipeThreshold:  50, // Minimum pixels for swipe
		tapMaxDuration:  200 * time.Millisecond,
		doubleTapWindow: 300 * time.Millisecond,
	}
}

// OnGesture registers a handler for a gesture type.
func (g *GestureRecognizer) OnGesture(gestureType GestureType, handler func(Gesture)) {
	g.handlers[gestureType] = handler
}

// Update should be called each frame to detect gestures.
func (g *GestureRecognizer) Update() []Gesture {
	gestures := make([]Gesture, 0)

	// Check for ended touches (potential taps/swipes)
	for id, prevState := range g.touchMgr.prevTouches {
		currentState := g.touchMgr.GetTouch(id)
		if currentState == nil || !currentState.IsActive {
			// Touch ended, analyze the gesture
			if prevState.IsActive {
				gesture := g.analyzeTouchEnd(prevState)
				if gesture != nil {
					gestures = append(gestures, *gesture)
					if handler, ok := g.handlers[gesture.Type]; ok {
						handler(*gesture)
					}
				}
			}
		}
	}

	// Check for ongoing drags
	for _, state := range g.touchMgr.GetAllTouches() {
		dx := float64(state.CurrentX - state.StartX)
		dy := float64(state.CurrentY - state.StartY)
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance > 10 && time.Since(state.StartTime) > 100*time.Millisecond {
			gesture := Gesture{
				Type: GestureDrag,
				X:    state.CurrentX,
				Y:    state.CurrentY,
				DX:   dx,
				DY:   dy,
			}

			gestures = append(gestures, gesture)
			if handler, ok := g.handlers[GestureDrag]; ok {
				handler(gesture)
			}
		}
	}

	// Check for pinch (two-finger gesture)
	touches := g.touchMgr.GetAllTouches()
	if len(touches) >= 2 {
		gesture := g.detectPinch(touches[0], touches[1])
		if gesture != nil {
			gestures = append(gestures, *gesture)
			if handler, ok := g.handlers[GesturePinch]; ok {
				handler(*gesture)
			}
		}
	}

	return gestures
}

// analyzeTouchEnd determines what gesture a completed touch represents.
func (g *GestureRecognizer) analyzeTouchEnd(state *TouchState) *Gesture {
	duration := time.Since(state.StartTime)
	dx := float64(state.CurrentX - state.StartX)
	dy := float64(state.CurrentY - state.StartY)
	distance := math.Sqrt(dx*dx + dy*dy)

	// Check for swipe
	if distance > g.swipeThreshold {
		var gestureType GestureType

		if math.Abs(dx) > math.Abs(dy) {
			if dx > 0 {
				gestureType = GestureSwipeRight
			} else {
				gestureType = GestureSwipeLeft
			}
		} else {
			if dy > 0 {
				gestureType = GestureSwipeDown
			} else {
				gestureType = GestureSwipeUp
			}
		}

		return &Gesture{
			Type:     gestureType,
			X:        state.StartX,
			Y:        state.StartY,
			DX:       dx,
			DY:       dy,
			Velocity: distance / duration.Seconds(),
		}
	}

	// Check for tap
	if duration < g.tapMaxDuration && distance < 20 {
		// Check for double tap
		timeSinceLastTap := time.Since(g.lastTapTime)
		distFromLastTap := math.Sqrt(
			math.Pow(float64(state.CurrentX-g.lastTapX), 2) +
				math.Pow(float64(state.CurrentY-g.lastTapY), 2),
		)

		g.lastTapTime = time.Now()
		g.lastTapX = state.CurrentX
		g.lastTapY = state.CurrentY

		if timeSinceLastTap < g.doubleTapWindow && distFromLastTap < 50 {
			return &Gesture{
				Type: GestureDoubleTap,
				X:    state.CurrentX,
				Y:    state.CurrentY,
			}
		}

		return &Gesture{
			Type: GestureTap,
			X:    state.CurrentX,
			Y:    state.CurrentY,
		}
	}

	return nil
}

// detectPinch detects pinch gesture from two touches.
func (g *GestureRecognizer) detectPinch(t1, t2 *TouchState) *Gesture {
	// Calculate current and starting distances
	currentDist := math.Sqrt(
		math.Pow(float64(t1.CurrentX-t2.CurrentX), 2) +
			math.Pow(float64(t1.CurrentY-t2.CurrentY), 2),
	)
	startDist := math.Sqrt(
		math.Pow(float64(t1.StartX-t2.StartX), 2) +
			math.Pow(float64(t1.StartY-t2.StartY), 2),
	)

	if startDist < 1 {
		return nil
	}

	scale := currentDist / startDist
	centerX := (t1.CurrentX + t2.CurrentX) / 2
	centerY := (t1.CurrentY + t2.CurrentY) / 2

	return &Gesture{
		Type:  GesturePinch,
		X:     centerX,
		Y:     centerY,
		Scale: scale,
	}
}
