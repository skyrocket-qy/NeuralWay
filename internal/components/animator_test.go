package components

import (
	"testing"
)

// TestAnimator tests Animator component.
func TestAnimator(t *testing.T) {
	t.Run("NewAnimator creates default animator", func(t *testing.T) {
		a := NewAnimator(nil)

		if a.Speed != 1.0 {
			t.Errorf("NewAnimator.Speed = %v, want 1.0", a.Speed)
		}

		if !a.Playing {
			t.Error("NewAnimator should be playing by default")
		}

		if a.CurrentFrame != 0 {
			t.Errorf("NewAnimator.CurrentFrame = %v, want 0", a.CurrentFrame)
		}
	})

	t.Run("Play switches animation", func(t *testing.T) {
		a := NewAnimator(nil)
		a.CurrentAnim = "idle"
		a.CurrentFrame = 5
		a.FrameTimer = 0.5

		a.Play("walk")

		if a.CurrentAnim != "walk" {
			t.Errorf("Animator.CurrentAnim = %v, want 'walk'", a.CurrentAnim)
		}

		if a.CurrentFrame != 0 {
			t.Errorf("Animator.CurrentFrame = %v, want 0 after play", a.CurrentFrame)
		}

		if a.FrameTimer != 0 {
			t.Errorf("Animator.FrameTimer = %v, want 0 after play", a.FrameTimer)
		}
	})

	t.Run("Play same animation does nothing", func(t *testing.T) {
		a := NewAnimator(nil)
		a.CurrentAnim = "walk"
		a.CurrentFrame = 3
		a.FrameTimer = 0.5

		a.Play("walk")

		// Should not reset when playing same animation
		if a.CurrentFrame != 3 {
			t.Errorf("Playing same animation reset frame to %v, should stay 3", a.CurrentFrame)
		}
	})

	t.Run("Stop pauses animation", func(t *testing.T) {
		a := NewAnimator(nil)
		a.Stop()

		if a.Playing {
			t.Error("Stop should set Playing to false")
		}
	})

	t.Run("Resume continues animation", func(t *testing.T) {
		a := NewAnimator(nil)
		a.Stop()
		a.Resume()

		if !a.Playing {
			t.Error("Resume should set Playing to true")
		}
	})

	t.Run("Reset restarts current animation", func(t *testing.T) {
		a := NewAnimator(nil)
		a.CurrentFrame = 5
		a.FrameTimer = 0.75

		a.Reset()

		if a.CurrentFrame != 0 {
			t.Errorf("Reset should set CurrentFrame to 0, got %v", a.CurrentFrame)
		}

		if a.FrameTimer != 0 {
			t.Errorf("Reset should set FrameTimer to 0, got %v", a.FrameTimer)
		}
	})

	t.Run("GetCurrentSprite with nil AnimationSet", func(t *testing.T) {
		a := NewAnimator(nil)
		sprite := a.GetCurrentSprite()

		if sprite != nil {
			t.Error("GetCurrentSprite should return nil when AnimationSet is nil")
		}
	})
}
