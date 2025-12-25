package components

import "testing"

func TestNewAIMetadata(t *testing.T) {
	meta := NewAIMetadata("custom", "A custom entity")

	if meta.EntityType != "custom" {
		t.Errorf("EntityType mismatch: got %q, want %q", meta.EntityType, "custom")
	}

	if meta.Description != "A custom entity" {
		t.Errorf("Description mismatch: got %q", meta.Description)
	}

	if meta.VisualState != "idle" {
		t.Errorf("Default VisualState should be 'idle', got %q", meta.VisualState)
	}
}

func TestNewPlayerMetadata(t *testing.T) {
	meta := NewPlayerMetadata("Hero character")

	if meta.EntityType != "player" {
		t.Errorf("EntityType should be 'player', got %q", meta.EntityType)
	}

	if !meta.HasTag("controllable") {
		t.Error("Player should have 'controllable' tag")
	}

	if !meta.HasTag("collidable") {
		t.Error("Player should have 'collidable' tag")
	}
}

func TestNewEnemyMetadata(t *testing.T) {
	meta := NewEnemyMetadata("Goblin warrior")

	if meta.EntityType != "enemy" {
		t.Errorf("EntityType should be 'enemy', got %q", meta.EntityType)
	}

	if !meta.HasTag("hostile") {
		t.Error("Enemy should have 'hostile' tag")
	}
}

func TestAIMetadataTagOperations(t *testing.T) {
	meta := NewAIMetadata("test", "Test entity")

	// Add tag
	meta.AddTag("special")

	if !meta.HasTag("special") {
		t.Error("Tag 'special' should be present after AddTag")
	}

	// Add duplicate (should not duplicate)
	meta.AddTag("special")

	count := 0

	for _, tag := range meta.Tags {
		if tag == "special" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("Tag should not duplicate, found %d occurrences", count)
	}

	// Remove tag
	meta.RemoveTag("special")

	if meta.HasTag("special") {
		t.Error("Tag 'special' should be removed")
	}
}

func TestAIMetadataEffectOperations(t *testing.T) {
	meta := NewAIMetadata("test", "Test entity")

	// Add effect
	meta.AddEffect("glow")

	if len(meta.ActiveEffects) != 1 || meta.ActiveEffects[0] != "glow" {
		t.Error("Effect not added correctly")
	}

	// Remove effect
	meta.RemoveEffect("glow")

	if len(meta.ActiveEffects) != 0 {
		t.Error("Effect should be removed")
	}
}

func TestAIMetadataSetState(t *testing.T) {
	meta := NewAIMetadata("test", "Test entity")

	meta.SetState("attacking")

	if meta.VisualState != "attacking" {
		t.Errorf("VisualState should be 'attacking', got %q", meta.VisualState)
	}
}
