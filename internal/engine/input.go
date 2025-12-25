package engine

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

// InputType identifies the type of input binding.
type InputType int

const (
	InputKeyboard InputType = iota
	InputGamepad
	InputMouse
)

// InputBinding represents a single input binding.
type InputBinding struct {
	Type      InputType
	Key       ebiten.Key
	Button    ebiten.GamepadButton
	MouseBtn  ebiten.MouseButton
	GamepadID ebiten.GamepadID
}

// InputManager handles action-based input mapping.
type InputManager struct {
	actions     map[string][]InputBinding
	prevKeys    map[ebiten.Key]bool
	prevButtons map[ebiten.GamepadButton]bool
	prevMouse   map[ebiten.MouseButton]bool
}

// NewInputManager creates a new input manager.
func NewInputManager() *InputManager {
	return &InputManager{
		actions:     make(map[string][]InputBinding),
		prevKeys:    make(map[ebiten.Key]bool),
		prevButtons: make(map[ebiten.GamepadButton]bool),
		prevMouse:   make(map[ebiten.MouseButton]bool),
	}
}

// BindKey binds a keyboard key to an action.
func (m *InputManager) BindKey(action string, key ebiten.Key) {
	m.actions[action] = append(m.actions[action], InputBinding{
		Type: InputKeyboard,
		Key:  key,
	})
}

// BindGamepadButton binds a gamepad button to an action.
func (m *InputManager) BindGamepadButton(
	action string,
	button ebiten.GamepadButton,
	gamepadID ebiten.GamepadID,
) {
	m.actions[action] = append(m.actions[action], InputBinding{
		Type:      InputGamepad,
		Button:    button,
		GamepadID: gamepadID,
	})
}

// BindMouseButton binds a mouse button to an action.
func (m *InputManager) BindMouseButton(action string, button ebiten.MouseButton) {
	m.actions[action] = append(m.actions[action], InputBinding{
		Type:     InputMouse,
		MouseBtn: button,
	})
}

// IsActionPressed returns true if any binding for the action is currently pressed.
func (m *InputManager) IsActionPressed(action string) bool {
	bindings, ok := m.actions[action]
	if !ok {
		return false
	}

	for _, b := range bindings {
		switch b.Type {
		case InputKeyboard:
			if ebiten.IsKeyPressed(b.Key) {
				return true
			}
		case InputGamepad:
			if ebiten.IsGamepadButtonPressed(b.GamepadID, b.Button) {
				return true
			}
		case InputMouse:
			if ebiten.IsMouseButtonPressed(b.MouseBtn) {
				return true
			}
		}
	}

	return false
}

// IsActionJustPressed returns true if any binding was just pressed this frame.
func (m *InputManager) IsActionJustPressed(action string) bool {
	bindings, ok := m.actions[action]
	if !ok {
		return false
	}

	for _, b := range bindings {
		switch b.Type {
		case InputKeyboard:
			if ebiten.IsKeyPressed(b.Key) && !m.prevKeys[b.Key] {
				return true
			}
		case InputGamepad:
			if ebiten.IsGamepadButtonPressed(b.GamepadID, b.Button) && !m.prevButtons[b.Button] {
				return true
			}
		case InputMouse:
			if ebiten.IsMouseButtonPressed(b.MouseBtn) && !m.prevMouse[b.MouseBtn] {
				return true
			}
		}
	}

	return false
}

// Update should be called at the end of each frame to track previous state.
func (m *InputManager) Update() {
	// Update previous key states
	for action, bindings := range m.actions {
		_ = action

		for _, b := range bindings {
			switch b.Type {
			case InputKeyboard:
				m.prevKeys[b.Key] = ebiten.IsKeyPressed(b.Key)
			case InputGamepad:
				m.prevButtons[b.Button] = ebiten.IsGamepadButtonPressed(b.GamepadID, b.Button)
			case InputMouse:
				m.prevMouse[b.MouseBtn] = ebiten.IsMouseButtonPressed(b.MouseBtn)
			}
		}
	}
}

// GetAxis returns a value from -1 to 1 based on negative and positive bindings.
// Useful for movement (e.g., left/right or up/down).
func (m *InputManager) GetAxis(negAction, posAction string) float64 {
	var value float64
	if m.IsActionPressed(negAction) {
		value -= 1
	}

	if m.IsActionPressed(posAction) {
		value += 1
	}

	return value
}

// GetGamepadAxis returns a gamepad stick axis value (-1 to 1).
func (m *InputManager) GetGamepadAxis(gamepadID ebiten.GamepadID, axis int) float64 {
	ids := ebiten.AppendGamepadIDs(nil)
	if slices.Contains(ids, gamepadID) {
		return ebiten.GamepadAxisValue(gamepadID, axis)
	}

	return 0
}

// IsGamepadConnected returns true if the gamepad is connected.
func (m *InputManager) IsGamepadConnected(gamepadID ebiten.GamepadID) bool {
	ids := ebiten.AppendGamepadIDs(nil)

	return slices.Contains(ids, gamepadID)
}

// GetConnectedGamepads returns all connected gamepad IDs.
func (m *InputManager) GetConnectedGamepads() []ebiten.GamepadID {
	return ebiten.AppendGamepadIDs(nil)
}

// ClearBindings removes all bindings for an action.
func (m *InputManager) ClearBindings(action string) {
	delete(m.actions, action)
}

// ClearAllBindings removes all bindings.
func (m *InputManager) ClearAllBindings() {
	m.actions = make(map[string][]InputBinding)
}
