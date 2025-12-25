package systems_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
	"td/internal/assets"
	"td/internal/components"
	"td/internal/systems"
)

func TestAnimationSystem(t *testing.T) {
	// Setup world
	world := ecs.NewWorld()
	sys := systems.NewAnimationSystem(&world)

	// Create dummy animation set
	img := ebiten.NewImage(64, 32) // 2 frames of 32x32
	sheet := assets.NewSpriteSheet(img, 32, 32)
	animSet := assets.NewAnimationSet(sheet)
	animSet.Add("idle", []int{0, 1}, 0.1, true)

	// Create entity using Mapper
	mapper := ecs.NewMap2[components.Animator, components.Sprite](&world)

	animator := components.NewAnimator(animSet)
	animator.Play("idle")

	sprite := components.NewSprite(nil)

	// Ensure we pass correct types
	// NewAnimator returns *Animator
	// NewSprite returns Sprite, so we pass &sprite
	mapper.NewEntity(animator, &sprite)

	// Initial state
	sys.SetDeltaTime(0)
	sys.Update(&world)
	// Sprite image is updated in place via ECS query

	// Query back to verify
	query := ecs.NewFilter2[components.Animator, components.Sprite](&world).Query()
	if query.Next() {
		// _, s := query.Get()
		// Note: in Ark ECS, query.Get() returns pointers to live components
		_, s := query.Get()
		if s.Image == nil {
			t.Error("Sprite should have image after first update")
		}
	} else {
		t.Fatal("Entity not found")
	}

	// Advance time (half frame)
	sys.SetDeltaTime(0.05)
	sys.Update(&world)

	// Verify frame 0
	query = ecs.NewFilter2[components.Animator, components.Sprite](&world).Query()
	if query.Next() {
		a, _ := query.Get()
		if a.CurrentFrame != 0 {
			t.Errorf("Should be frame 0, got %d", a.CurrentFrame)
		}
	}

	// Advance time (next frame)
	sys.SetDeltaTime(0.06) // Total > 0.1
	sys.Update(&world)

	// Verify frame 1
	query = ecs.NewFilter2[components.Animator, components.Sprite](&world).Query()
	if query.Next() {
		a, _ := query.Get()
		if a.CurrentFrame != 1 {
			t.Errorf("Should be frame 1, got %d", a.CurrentFrame)
		}
	}
}
