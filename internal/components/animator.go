package components

import (
	"github.com/hajimehoshi/ebiten/v2"
	"td/internal/assets"
)

// Animator is a component that controls sprite animation.
type Animator struct {
	AnimationSet   *assets.AnimationSet
	CurrentAnim    string
	CurrentFrame   int
	FrameTimer     float64
	Speed          float64 // Animation speed multiplier
	Playing        bool
	OnAnimationEnd func(name string)
}

// NewAnimator creates an animator component.
func NewAnimator(animSet *assets.AnimationSet) *Animator {
	return &Animator{
		AnimationSet: animSet,
		Speed:        1.0,
		Playing:      true,
	}
}

// Play starts or switches to an animation.
func (a *Animator) Play(name string) {
	if a.CurrentAnim == name {
		return
	}

	a.CurrentAnim = name
	a.CurrentFrame = 0
	a.FrameTimer = 0
	a.Playing = true
}

// Stop pauses the animation.
func (a *Animator) Stop() {
	a.Playing = false
}

// Resume continues the animation.
func (a *Animator) Resume() {
	a.Playing = true
}

// Reset restarts the current animation from the beginning.
func (a *Animator) Reset() {
	a.CurrentFrame = 0
	a.FrameTimer = 0
}

// GetCurrentSprite returns the current frame's sprite image.
func (a *Animator) GetCurrentSprite() *ebiten.Image {
	if a.AnimationSet == nil || a.AnimationSet.Sheet == nil {
		return nil
	}

	anim := a.AnimationSet.Get(a.CurrentAnim)
	if anim == nil || len(anim.Frames) == 0 {
		return nil
	}

	frameIdx := anim.Frames[a.CurrentFrame]

	return a.AnimationSet.Sheet.Frame(frameIdx)
}
