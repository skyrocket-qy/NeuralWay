package systems

import (
	"td/internal/assets"
	"td/internal/components"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
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

// AnimationSystem updates all animators and their sprites.
type AnimationSystem struct {
	animFilter *ecs.Filter2[Animator, components.Sprite]
	deltaTime  float64
}

// NewAnimationSystem creates an animation system.
func NewAnimationSystem(world *ecs.World) *AnimationSystem {
	return &AnimationSystem{
		animFilter: ecs.NewFilter2[Animator, components.Sprite](world),
		deltaTime:  1.0 / 60.0, // Default 60 FPS
	}
}

// SetDeltaTime sets the delta time for this frame.
func (s *AnimationSystem) SetDeltaTime(dt float64) {
	s.deltaTime = dt
}

// Update advances all animations.
func (s *AnimationSystem) Update(world *ecs.World) {
	query := s.animFilter.Query()
	for query.Next() {
		anim, sprite := query.Get()

		if !anim.Playing || anim.AnimationSet == nil {
			continue
		}

		animation := anim.AnimationSet.Get(anim.CurrentAnim)
		if animation == nil || len(animation.Frames) == 0 {
			continue
		}

		// Update timer
		anim.FrameTimer += s.deltaTime * anim.Speed

		// Check for frame advance
		if anim.FrameTimer >= animation.Duration {
			anim.FrameTimer -= animation.Duration
			anim.CurrentFrame++

			// Handle animation end
			if anim.CurrentFrame >= len(animation.Frames) {
				if animation.Loop {
					anim.CurrentFrame = 0
				} else {
					anim.CurrentFrame = len(animation.Frames) - 1
					anim.Playing = false
					if anim.OnAnimationEnd != nil {
						anim.OnAnimationEnd(anim.CurrentAnim)
					}
				}
			}
		}

		// Update sprite image
		if img := anim.GetCurrentSprite(); img != nil {
			sprite.Image = img
		}
	}
}

// FlipBook is a simpler animation system without AnimationSet.
// It just cycles through a slice of images.
type FlipBook struct {
	Frames       []*ebiten.Image
	CurrentFrame int
	FrameTimer   float64
	Duration     float64 // Duration per frame
	Loop         bool
	Playing      bool
}

// NewFlipBook creates a simple flipbook animation.
func NewFlipBook(frames []*ebiten.Image, duration float64, loop bool) *FlipBook {
	return &FlipBook{
		Frames:   frames,
		Duration: duration,
		Loop:     loop,
		Playing:  true,
	}
}

// Update advances the flipbook animation.
func (f *FlipBook) Update(dt float64) {
	if !f.Playing || len(f.Frames) == 0 {
		return
	}

	f.FrameTimer += dt
	if f.FrameTimer >= f.Duration {
		f.FrameTimer -= f.Duration
		f.CurrentFrame++

		if f.CurrentFrame >= len(f.Frames) {
			if f.Loop {
				f.CurrentFrame = 0
			} else {
				f.CurrentFrame = len(f.Frames) - 1
				f.Playing = false
			}
		}
	}
}

// CurrentImage returns the current frame image.
func (f *FlipBook) CurrentImage() *ebiten.Image {
	if f.CurrentFrame < 0 || f.CurrentFrame >= len(f.Frames) {
		return nil
	}
	return f.Frames[f.CurrentFrame]
}
