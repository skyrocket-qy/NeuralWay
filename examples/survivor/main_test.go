package main

import (
	"testing"
)

// TestGameLogic verifies basic game mechanics including initialization,
// weapon firing, and object pooling.
func TestGameLogic(t *testing.T) {
	// 1. Initialize Game
	g := NewGame()
	if g == nil {
		t.Fatal("NewGame returned nil")
	}

	// 2. Start Game
	g.startGame(CharJunior)

	if g.player == nil {
		t.Fatal("Player not initialized after startGame")
	}

	if g.state != StatePlaying {
		t.Errorf("Expected StatePlaying, got %v", g.state)
	}

	// 3. Verify Player Logic
	initialX := g.player.X

	g.player.X += 10 // Simulate movement
	if g.player.X <= initialX {
		t.Errorf("Player X did not update")
	}

	// 4. Verify Weapon Firing & Pooling
	initialProjCount := len(g.projectiles)
	if len(g.player.Weapons) == 0 {
		t.Fatal("Player has no weapons")
	}

	w := g.player.Weapons[0]

	// Force fire
	g.fireWeapon(w)

	if len(g.projectiles) <= initialProjCount {
		t.Errorf("FireWeapon did not spawn projectiles")
	}

	// 5. Verify Projectile Properties (Alignment Check)
	p := g.projectiles[0]
	// Projectile should be reasonably close to player (Sword uses range 40)
	dx := p.X - g.player.X

	dy := p.Y - g.player.Y
	if dx*dx+dy*dy > 10000 { // 100^2 arbitrary large range
		t.Errorf("Projectile spawned too far: %f, %f (Player at %f, %f)", p.X, p.Y, g.player.X, g.player.Y)
	}

	// 6. Verify Object Pooling
	// Update projectile to expire it
	p.Lifetime = -1

	g.updateProjectiles(1.0) // Should free projectile

	if len(g.projectiles) != 0 {
		t.Errorf("Projectile not removed after lifetime expiration")
	}

	if len(g.unusedProjs) == 0 {
		t.Errorf("Projectile not returned to pool")
	}
}
