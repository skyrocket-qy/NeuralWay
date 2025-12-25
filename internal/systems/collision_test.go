package systems

import (
	"testing"
)

// TestSpatialHash tests the SpatialHash data structure.
func TestSpatialHash(t *testing.T) {
	t.Run("NewSpatialHash creates empty hash", func(t *testing.T) {
		sh := NewSpatialHash(64)
		if sh.cellSize != 64 {
			t.Errorf("SpatialHash.cellSize = %v, want 64", sh.cellSize)
		}

		if len(sh.cells) != 0 {
			t.Errorf("SpatialHash.cells should be empty")
		}
	})

	t.Run("Clear removes all entities", func(t *testing.T) {
		sh := NewSpatialHash(64)
		// Manually add some entries using the hash function directly
		// We can't easily create ecs.Entity, so just verify Clear works on the map
		key1 := sh.hash(0, 0)
		key2 := sh.hash(1, 1)
		sh.cells[key1] = nil // Empty slice, just to have entries
		sh.cells[key2] = nil

		sh.Clear()

		if len(sh.cells) != 0 {
			t.Errorf("Clear should remove all cells, got %v", len(sh.cells))
		}
	})

	t.Run("cellCoords calculates correctly", func(t *testing.T) {
		sh := NewSpatialHash(64)

		tests := []struct {
			x, y         float64
			wantX, wantY int
		}{
			{0, 0, 0, 0},
			{32, 32, 0, 0},
			{64, 64, 1, 1},
			{128, 192, 2, 3},
			{-64, -64, -1, -1},
		}

		for _, tt := range tests {
			cx, cy := sh.cellCoords(tt.x, tt.y)
			if cx != tt.wantX || cy != tt.wantY {
				t.Errorf("cellCoords(%v, %v) = (%v, %v), want (%v, %v)",
					tt.x, tt.y, cx, cy, tt.wantX, tt.wantY)
			}
		}
	})

	t.Run("hash creates unique keys", func(t *testing.T) {
		sh := NewSpatialHash(64)

		key1 := sh.hash(0, 0)
		key2 := sh.hash(1, 0)
		key3 := sh.hash(0, 1)
		key4 := sh.hash(1, 1)

		keys := map[int64]bool{key1: true, key2: true, key3: true, key4: true}
		if len(keys) != 4 {
			t.Error("hash should produce unique keys for different coordinates")
		}
	})
}

// TestAABBCollision tests AABB collision detection.
func TestAABBCollision(t *testing.T) {
	tests := []struct {
		name           string
		x1, y1, w1, h1 float64
		x2, y2, w2, h2 float64
		want           bool
	}{
		{"overlapping", 0, 0, 32, 32, 16, 16, 32, 32, true},
		{"touching edge", 0, 0, 32, 32, 32, 0, 32, 32, false},
		{"separate horizontal", 0, 0, 32, 32, 100, 0, 32, 32, false},
		{"separate vertical", 0, 0, 32, 32, 0, 100, 32, 32, false},
		{"contained", 10, 10, 10, 10, 0, 0, 50, 50, true},
		{"same position", 0, 0, 32, 32, 0, 0, 32, 32, true},
		{"partial overlap", 0, 0, 32, 32, 16, 16, 16, 16, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := aabbCollision(tt.x1, tt.y1, tt.w1, tt.h1, tt.x2, tt.y2, tt.w2, tt.h2)
			if got != tt.want {
				t.Errorf("aabbCollision = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPointInAABB tests point-in-AABB detection.
func TestPointInAABB(t *testing.T) {
	tests := []struct {
		name       string
		px, py     float64
		x, y, w, h float64
		want       bool
	}{
		{"inside", 16, 16, 0, 0, 32, 32, true},
		{"outside right", 50, 16, 0, 0, 32, 32, false},
		{"outside left", -10, 16, 0, 0, 32, 32, false},
		{"outside top", 16, -10, 0, 0, 32, 32, false},
		{"outside bottom", 16, 50, 0, 0, 32, 32, false},
		{"on corner", 0, 0, 0, 0, 32, 32, true},
		{"on edge", 32, 16, 0, 0, 32, 32, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PointInAABB(tt.px, tt.py, tt.x, tt.y, tt.w, tt.h)
			if got != tt.want {
				t.Errorf("PointInAABB = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCircleCollision tests circle collision detection.
func TestCircleCollision(t *testing.T) {
	tests := []struct {
		name       string
		x1, y1, r1 float64
		x2, y2, r2 float64
		want       bool
	}{
		{"overlapping", 0, 0, 16, 20, 0, 16, true},
		{"touching", 0, 0, 16, 32, 0, 16, false},
		{"separate", 0, 0, 16, 100, 0, 16, false},
		{"same center", 0, 0, 16, 0, 0, 16, true},
		{"contained", 0, 0, 32, 0, 0, 8, true},
		{"diagonal overlap", 0, 0, 20, 10, 10, 20, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CircleCollision(tt.x1, tt.y1, tt.r1, tt.x2, tt.y2, tt.r2)
			if got != tt.want {
				t.Errorf("CircleCollision = %v, want %v", got, tt.want)
			}
		})
	}
}
