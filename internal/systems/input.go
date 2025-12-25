package systems

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// InputState represents the current state of all inputs.
type InputState struct {
	// Keyboard
	KeysPressed  map[ebiten.Key]bool
	KeysJustDown map[ebiten.Key]bool
	KeysJustUp   map[ebiten.Key]bool

	// Mouse
	MouseX, MouseY    int
	MouseDX, MouseDY  int // Delta since last frame
	MouseButtons      map[ebiten.MouseButton]bool
	MouseJustPressed  map[ebiten.MouseButton]bool
	MouseJustReleased map[ebiten.MouseButton]bool
	MouseWheelX       float64
	MouseWheelY       float64

	// Touch
	Touches     []Touch
	TouchJustOn []Touch

	// Gamepad
	GamepadIDs     []ebiten.GamepadID
	GamepadButtons map[ebiten.GamepadID]map[ebiten.GamepadButton]bool
	GamepadAxes    map[ebiten.GamepadID]map[int]float64
}

// Touch represents a single touch point.
type Touch struct {
	ID ebiten.TouchID
	X  int
	Y  int
}

// InputManager handles all input processing.
type InputManager struct {
	state    InputState
	prevMX   int
	prevMY   int
	bindings map[string][]ebiten.Key
}

// NewInputManager creates a new input manager.
func NewInputManager() *InputManager {
	return &InputManager{
		state: InputState{
			KeysPressed:       make(map[ebiten.Key]bool),
			KeysJustDown:      make(map[ebiten.Key]bool),
			KeysJustUp:        make(map[ebiten.Key]bool),
			MouseButtons:      make(map[ebiten.MouseButton]bool),
			MouseJustPressed:  make(map[ebiten.MouseButton]bool),
			MouseJustReleased: make(map[ebiten.MouseButton]bool),
			GamepadButtons:    make(map[ebiten.GamepadID]map[ebiten.GamepadButton]bool),
			GamepadAxes:       make(map[ebiten.GamepadID]map[int]float64),
		},
		bindings: make(map[string][]ebiten.Key),
	}
}

// BindAction binds a named action to one or more keys.
func (m *InputManager) BindAction(name string, keys ...ebiten.Key) {
	m.bindings[name] = keys
}

// IsActionPressed returns true if any key bound to the action is pressed.
func (m *InputManager) IsActionPressed(name string) bool {
	keys, ok := m.bindings[name]
	if !ok {
		return false
	}

	for _, k := range keys {
		if m.state.KeysPressed[k] {
			return true
		}
	}

	return false
}

// IsActionJustPressed returns true if any key bound to the action was just pressed.
func (m *InputManager) IsActionJustPressed(name string) bool {
	keys, ok := m.bindings[name]
	if !ok {
		return false
	}

	for _, k := range keys {
		if m.state.KeysJustDown[k] {
			return true
		}
	}

	return false
}

// Update polls all input devices and updates state.
func (m *InputManager) Update() {
	// Clear just-pressed/released maps
	for k := range m.state.KeysJustDown {
		delete(m.state.KeysJustDown, k)
	}

	for k := range m.state.KeysJustUp {
		delete(m.state.KeysJustUp, k)
	}

	for b := range m.state.MouseJustPressed {
		delete(m.state.MouseJustPressed, b)
	}

	for b := range m.state.MouseJustReleased {
		delete(m.state.MouseJustReleased, b)
	}

	// Update keyboard
	for k := range m.state.KeysPressed {
		if !ebiten.IsKeyPressed(k) {
			m.state.KeysJustUp[k] = true
			delete(m.state.KeysPressed, k)
		}
	}

	for _, k := range inpututil.AppendPressedKeys(nil) {
		if !m.state.KeysPressed[k] {
			m.state.KeysJustDown[k] = true
		}

		m.state.KeysPressed[k] = true
	}

	// Update mouse position and delta
	mx, my := ebiten.CursorPosition()
	m.state.MouseDX = mx - m.prevMX
	m.state.MouseDY = my - m.prevMY
	m.prevMX, m.prevMY = mx, my
	m.state.MouseX, m.state.MouseY = mx, my

	// Update mouse wheel
	wx, wy := ebiten.Wheel()
	m.state.MouseWheelX = wx
	m.state.MouseWheelY = wy

	// Update mouse buttons
	buttons := []ebiten.MouseButton{
		ebiten.MouseButtonLeft,
		ebiten.MouseButtonRight,
		ebiten.MouseButtonMiddle,
	}
	for _, b := range buttons {
		pressed := ebiten.IsMouseButtonPressed(b)
		wasPressed := m.state.MouseButtons[b]

		if pressed && !wasPressed {
			m.state.MouseJustPressed[b] = true
		}

		if !pressed && wasPressed {
			m.state.MouseJustReleased[b] = true
		}

		m.state.MouseButtons[b] = pressed
	}

	// Update touches
	m.state.Touches = m.state.Touches[:0]
	m.state.TouchJustOn = m.state.TouchJustOn[:0]

	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	for _, id := range touchIDs {
		x, y := ebiten.TouchPosition(id)
		m.state.TouchJustOn = append(m.state.TouchJustOn, Touch{ID: id, X: x, Y: y})
	}

	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		m.state.Touches = append(m.state.Touches, Touch{ID: id, X: x, Y: y})
	}

	// Update gamepads
	m.state.GamepadIDs = ebiten.AppendGamepadIDs(m.state.GamepadIDs[:0])
	for _, gid := range m.state.GamepadIDs {
		if m.state.GamepadButtons[gid] == nil {
			m.state.GamepadButtons[gid] = make(map[ebiten.GamepadButton]bool)
			m.state.GamepadAxes[gid] = make(map[int]float64)
		}

		// Update buttons
		for b := range ebiten.GamepadButtonMax {
			m.state.GamepadButtons[gid][b] = ebiten.IsGamepadButtonPressed(gid, b)
		}

		// Update axes
		axisCount := ebiten.GamepadAxisCount(gid)
		for a := range axisCount {
			m.state.GamepadAxes[gid][a] = ebiten.GamepadAxisValue(gid, a)
		}
	}
}

// State returns the current input state.
func (m *InputManager) State() *InputState {
	return &m.state
}

// IsKeyPressed returns true if the key is currently held.
func (m *InputManager) IsKeyPressed(key ebiten.Key) bool {
	return m.state.KeysPressed[key]
}

// IsKeyJustPressed returns true if the key was just pressed this frame.
func (m *InputManager) IsKeyJustPressed(key ebiten.Key) bool {
	return m.state.KeysJustDown[key]
}

// IsMouseButtonPressed returns true if the mouse button is held.
func (m *InputManager) IsMouseButtonPressed(button ebiten.MouseButton) bool {
	return m.state.MouseButtons[button]
}

// IsMouseButtonJustPressed returns true if the button was just pressed.
func (m *InputManager) IsMouseButtonJustPressed(button ebiten.MouseButton) bool {
	return m.state.MouseJustPressed[button]
}

// MousePosition returns the current mouse position.
func (m *InputManager) MousePosition() (int, int) {
	return m.state.MouseX, m.state.MouseY
}

// MouseDelta returns the mouse movement since last frame.
func (m *InputManager) MouseDelta() (int, int) {
	return m.state.MouseDX, m.state.MouseDY
}

// GetAxis returns a 1D axis value from key bindings (-1, 0, or 1).
func (m *InputManager) GetAxis(negative, positive ebiten.Key) float64 {
	var val float64
	if m.state.KeysPressed[negative] {
		val -= 1
	}

	if m.state.KeysPressed[positive] {
		val += 1
	}

	return val
}

// GetVector returns a 2D movement vector from WASD or arrow keys.
func (m *InputManager) GetVector() (float64, float64) {
	x := m.GetAxis(ebiten.KeyA, ebiten.KeyD) + m.GetAxis(ebiten.KeyLeft, ebiten.KeyRight)
	y := m.GetAxis(ebiten.KeyW, ebiten.KeyS) + m.GetAxis(ebiten.KeyUp, ebiten.KeyDown)

	// Clamp to -1, 1
	if x > 1 {
		x = 1
	} else if x < -1 {
		x = -1
	}

	if y > 1 {
		y = 1
	} else if y < -1 {
		y = -1
	}

	return x, y
}
