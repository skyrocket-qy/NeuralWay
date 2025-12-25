package components

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestPosition tests Position component.
func TestPosition(t *testing.T) {
	tests := []struct {
		name string
		x, y float64
	}{
		{"zero", 0, 0},
		{"positive", 100.5, 200.75},
		{"negative", -50, -100},
		{"mixed", 100, -50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := Position{X: tt.x, Y: tt.y}
			if pos.X != tt.x {
				t.Errorf("Position.X = %v, want %v", pos.X, tt.x)
			}

			if pos.Y != tt.y {
				t.Errorf("Position.Y = %v, want %v", pos.Y, tt.y)
			}
		})
	}
}

// TestVelocity tests Velocity component.
func TestVelocity(t *testing.T) {
	tests := []struct {
		name string
		x, y float64
	}{
		{"zero", 0, 0},
		{"moving right", 5.0, 0},
		{"moving up", 0, -3.0},
		{"diagonal", 2.5, 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vel := Velocity{X: tt.x, Y: tt.y}
			if vel.X != tt.x {
				t.Errorf("Velocity.X = %v, want %v", vel.X, tt.x)
			}

			if vel.Y != tt.y {
				t.Errorf("Velocity.Y = %v, want %v", vel.Y, tt.y)
			}
		})
	}
}

// TestHealth tests Health component.
func TestHealth(t *testing.T) {
	t.Run("NewHealth creates full health", func(t *testing.T) {
		h := NewHealth(100)
		if h.Current != 100 {
			t.Errorf("NewHealth(100).Current = %v, want 100", h.Current)
		}

		if h.Max != 100 {
			t.Errorf("NewHealth(100).Max = %v, want 100", h.Max)
		}
	})

	t.Run("health can be reduced", func(t *testing.T) {
		h := NewHealth(100)

		h.Current -= 30
		if h.Current != 70 {
			t.Errorf("Health.Current = %v, want 70", h.Current)
		}
	})

	t.Run("health can go negative", func(t *testing.T) {
		h := NewHealth(10)

		h.Current -= 20
		if h.Current != -10 {
			t.Errorf("Health.Current = %v, want -10", h.Current)
		}
	})

	t.Run("health can be healed", func(t *testing.T) {
		h := NewHealth(100)
		h.Current = 50

		h.Current += 30
		if h.Current != 80 {
			t.Errorf("Health.Current = %v, want 80", h.Current)
		}
	})
}

// TestCollider tests Collider component.
func TestCollider(t *testing.T) {
	t.Run("basic collider", func(t *testing.T) {
		c := Collider{Width: 32, Height: 32, Layer: 1, Mask: 0xFF}
		if c.Width != 32 {
			t.Errorf("Collider.Width = %v, want 32", c.Width)
		}

		if c.Height != 32 {
			t.Errorf("Collider.Height = %v, want 32", c.Height)
		}
	})

	t.Run("layer masking", func(t *testing.T) {
		// Player on layer 1, collides with enemies (layer 2)
		player := Collider{Width: 32, Height: 32, Layer: 1, Mask: 2}
		enemy := Collider{Width: 32, Height: 32, Layer: 2, Mask: 1}

		// Check if they can collide with each other
		if player.Layer&enemy.Mask == 0 {
			t.Error("Enemy should be able to collide with player")
		}

		if enemy.Layer&player.Mask == 0 {
			t.Error("Player should be able to collide with enemy")
		}
	})
}

// TestSprite tests Sprite component.
func TestSprite(t *testing.T) {
	t.Run("NewSprite creates visible sprite with default scale", func(t *testing.T) {
		// Create a test image
		img := ebiten.NewImage(16, 16)
		s := NewSprite(img)

		if s.Image != img {
			t.Error("NewSprite should store the provided image")
		}

		if s.ScaleX != 1 || s.ScaleY != 1 {
			t.Errorf("NewSprite scale = (%v, %v), want (1, 1)", s.ScaleX, s.ScaleY)
		}

		if !s.Visible {
			t.Error("NewSprite should be visible by default")
		}

		if s.OffsetX != 0 || s.OffsetY != 0 {
			t.Errorf("NewSprite offset = (%v, %v), want (0, 0)", s.OffsetX, s.OffsetY)
		}
	})

	t.Run("NewSprite with nil image", func(t *testing.T) {
		s := NewSprite(nil)
		if s.Image != nil {
			t.Error("NewSprite(nil) should have nil image")
		}
		// Should still have valid defaults
		if s.ScaleX != 1 || s.ScaleY != 1 {
			t.Error("NewSprite(nil) should still have default scale")
		}
	})
}

// TestTag tests Tag component.
func TestTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
	}{
		{"player", "player"},
		{"enemy", "enemy"},
		{"bullet", "bullet"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := Tag{Name: tt.tag}
			if tag.Name != tt.tag {
				t.Errorf("Tag.Name = %v, want %v", tag.Name, tt.tag)
			}
		})
	}
}
