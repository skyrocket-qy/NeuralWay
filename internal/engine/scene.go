package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Scene represents a game scene (menu, gameplay, pause, etc.)
type Scene interface {
	// Load is called when entering the scene.
	Load() error
	// Unload is called when leaving the scene.
	Unload()
	// Update is called every frame.
	Update() error
	// Draw is called every frame to render.
	Draw(screen *ebiten.Image)
}

// SceneManager handles scene transitions.
type SceneManager struct {
	current      Scene
	next         Scene
	transition   Transition
	inTransition bool
}

// Transition defines how scenes switch.
type Transition interface {
	// Start initializes the transition.
	Start(from, to Scene)
	// Update advances the transition. Returns true when complete.
	Update() bool
	// Draw renders the transition effect.
	Draw(screen *ebiten.Image)
}

// NewSceneManager creates a scene manager.
func NewSceneManager() *SceneManager {
	return &SceneManager{}
}

// SetScene immediately switches to a new scene.
func (m *SceneManager) SetScene(scene Scene) error {
	if m.current != nil {
		m.current.Unload()
	}
	m.current = scene
	if m.current != nil {
		return m.current.Load()
	}
	return nil
}

// TransitionTo starts a transition to a new scene.
func (m *SceneManager) TransitionTo(scene Scene, transition Transition) {
	m.next = scene
	m.transition = transition
	m.inTransition = true
	if transition != nil {
		transition.Start(m.current, m.next)
	}
}

// Update updates the current scene or transition.
func (m *SceneManager) Update() error {
	if m.inTransition && m.transition != nil {
		if m.transition.Update() {
			// Transition complete
			if m.current != nil {
				m.current.Unload()
			}
			m.current = m.next
			m.next = nil
			if m.current != nil {
				if err := m.current.Load(); err != nil {
					return err
				}
			}
			m.inTransition = false
			m.transition = nil
		}
	} else if m.current != nil {
		return m.current.Update()
	}
	return nil
}

// Draw renders the current scene or transition.
func (m *SceneManager) Draw(screen *ebiten.Image) {
	if m.inTransition && m.transition != nil {
		m.transition.Draw(screen)
	} else if m.current != nil {
		m.current.Draw(screen)
	}
}

// Current returns the current scene.
func (m *SceneManager) Current() Scene {
	return m.current
}

// FadeTransition fades between scenes.
type FadeTransition struct {
	from      Scene
	to        Scene
	duration  float64
	elapsed   float64
	fadeOut   bool
	alpha     float64
	fadeColor *ebiten.Image
}

// NewFadeTransition creates a fade transition.
func NewFadeTransition(duration float64) *FadeTransition {
	return &FadeTransition{
		duration: duration,
	}
}

// Start initializes the fade transition.
func (t *FadeTransition) Start(from, to Scene) {
	t.from = from
	t.to = to
	t.elapsed = 0
	t.fadeOut = true
	t.alpha = 0
}

// Update advances the fade.
func (t *FadeTransition) Update() bool {
	t.elapsed += 1.0 / 60.0 // Assume 60 FPS

	halfDuration := t.duration / 2

	if t.elapsed < halfDuration {
		// Fading out
		t.alpha = t.elapsed / halfDuration
	} else if t.elapsed < t.duration {
		// Fading in
		t.fadeOut = false
		t.alpha = 1.0 - (t.elapsed-halfDuration)/halfDuration
	} else {
		return true // Complete
	}

	return false
}

// Draw renders the fade effect.
func (t *FadeTransition) Draw(screen *ebiten.Image) {
	// Draw the appropriate scene
	if t.fadeOut && t.from != nil {
		t.from.Draw(screen)
	} else if !t.fadeOut && t.to != nil {
		t.to.Draw(screen)
	}

	// Draw fade overlay
	if t.fadeColor == nil {
		bounds := screen.Bounds()
		t.fadeColor = ebiten.NewImage(bounds.Dx(), bounds.Dy())
	}

	t.fadeColor.Fill(color.RGBA{R: 0, G: 0, B: 0, A: uint8(t.alpha * 255)})
	screen.DrawImage(t.fadeColor, nil)
}

// BaseScene provides a basic scene implementation.
type BaseScene struct {
	Name string
}

// Load does nothing by default.
func (s *BaseScene) Load() error { return nil }

// Unload does nothing by default.
func (s *BaseScene) Unload() {}

// Update does nothing by default.
func (s *BaseScene) Update() error { return nil }

// Draw does nothing by default.
func (s *BaseScene) Draw(screen *ebiten.Image) {}
