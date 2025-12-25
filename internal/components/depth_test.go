package components

import (
	"testing"
)

// TestZIndex tests ZIndex component.
func TestZIndex(t *testing.T) {
	tests := []struct {
		name  string
		value int
	}{
		{"default", 0},
		{"foreground", 100},
		{"background", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := ZIndex{Value: tt.value}
			if z.Value != tt.value {
				t.Errorf("ZIndex.Value = %v, want %v", z.Value, tt.value)
			}
		})
	}
}

// TestElevation tests Elevation component.
func TestElevation(t *testing.T) {
	t.Run("jumping entity", func(t *testing.T) {
		e := Elevation{
			Height:   0,
			Ground:   100,
			Velocity: 10,
			Gravity:  0.5,
		}

		// Simulate one frame of jumping
		e.Height += e.Velocity
		e.Velocity -= e.Gravity

		if e.Height != 10 {
			t.Errorf("Elevation.Height = %v, want 10", e.Height)
		}

		if e.Velocity != 9.5 {
			t.Errorf("Elevation.Velocity = %v, want 9.5", e.Velocity)
		}
	})

	t.Run("grounded entity", func(t *testing.T) {
		e := Elevation{
			Height:   0,
			Ground:   100,
			Velocity: 0,
			Gravity:  1,
		}

		if e.Height != 0 {
			t.Error("Grounded entity should have height 0")
		}
	})
}

// TestShadow tests Shadow component.
func TestShadow(t *testing.T) {
	t.Run("NewShadow creates default shadow", func(t *testing.T) {
		s := NewShadow()

		if !s.Enabled {
			t.Error("NewShadow should be enabled")
		}

		if s.OffsetX != 0 {
			t.Errorf("NewShadow.OffsetX = %v, want 0", s.OffsetX)
		}

		if s.OffsetY != 8 {
			t.Errorf("NewShadow.OffsetY = %v, want 8", s.OffsetY)
		}

		if s.ScaleX != 1.0 {
			t.Errorf("NewShadow.ScaleX = %v, want 1.0", s.ScaleX)
		}

		if s.ScaleY != 0.5 {
			t.Errorf("NewShadow.ScaleY = %v, want 0.5", s.ScaleY)
		}

		if s.Opacity != 0.3 {
			t.Errorf("NewShadow.Opacity = %v, want 0.3", s.Opacity)
		}
	})

	t.Run("shadow opacity clamping", func(t *testing.T) {
		s := Shadow{Enabled: true, Opacity: 0.5}
		if s.Opacity < 0 || s.Opacity > 1 {
			t.Errorf("Shadow opacity %v should be between 0 and 1", s.Opacity)
		}
	})
}
