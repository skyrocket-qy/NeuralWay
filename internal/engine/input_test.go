package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestInputManagerBindKey(t *testing.T) {
	im := NewInputManager()

	im.BindKey("jump", ebiten.KeySpace)
	im.BindKey("jump", ebiten.KeyW)

	bindings, ok := im.actions["jump"]
	if !ok {
		t.Error("jump action should exist")
	}

	if len(bindings) != 2 {
		t.Errorf("jump should have 2 bindings, got %d", len(bindings))
	}
}

func TestInputManagerGetAxis(t *testing.T) {
	im := NewInputManager()

	im.BindKey("left", ebiten.KeyA)
	im.BindKey("right", ebiten.KeyD)

	// Without any keys pressed, axis should be 0
	// (Can't test actual key presses in unit tests)
	_ = im.GetAxis("left", "right")
}

func TestInputManagerClearBindings(t *testing.T) {
	im := NewInputManager()

	im.BindKey("action", ebiten.KeySpace)
	im.ClearBindings("action")

	_, ok := im.actions["action"]
	if ok {
		t.Error("action should be cleared")
	}
}

func TestInputManagerClearAllBindings(t *testing.T) {
	im := NewInputManager()

	im.BindKey("a", ebiten.KeyA)
	im.BindKey("b", ebiten.KeyB)
	im.ClearAllBindings()

	if len(im.actions) != 0 {
		t.Errorf("all bindings should be cleared, got %d", len(im.actions))
	}
}
