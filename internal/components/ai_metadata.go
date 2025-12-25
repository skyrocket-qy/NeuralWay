package components

import "slices"

// AIMetadata provides semantic metadata for visual entities.
// This helps LLMs understand what game elements represent.
type AIMetadata struct {
	// EntityType categorizes what this entity represents.
	// Common values: "player", "enemy", "projectile", "ui", "effect", "item", "npc"
	EntityType string

	// Description is a human-readable description for LLM context.
	// Example: "Player character - sword-wielding hero"
	Description string

	// VisualState describes the current visual state.
	// Common values: "idle", "moving", "attacking", "damaged", "dead", "hidden"
	VisualState string

	// Tags are filterable labels for grouping.
	// Examples: ["hostile", "animated", "collidable", "interactable"]
	Tags []string

	// ActiveEffects lists currently active visual effects.
	// Examples: ["flash_damage", "glow_powerup", "outline_selected"]
	ActiveEffects []string
}

// NewAIMetadata creates basic AI metadata.
func NewAIMetadata(entityType, description string) AIMetadata {
	return AIMetadata{
		EntityType:    entityType,
		Description:   description,
		VisualState:   "idle",
		Tags:          []string{},
		ActiveEffects: []string{},
	}
}

// NewPlayerMetadata creates metadata for player entities.
func NewPlayerMetadata(description string) AIMetadata {
	return AIMetadata{
		EntityType:    "player",
		Description:   description,
		VisualState:   "idle",
		Tags:          []string{"controllable", "collidable"},
		ActiveEffects: []string{},
	}
}

// NewEnemyMetadata creates metadata for enemy entities.
func NewEnemyMetadata(description string) AIMetadata {
	return AIMetadata{
		EntityType:    "enemy",
		Description:   description,
		VisualState:   "idle",
		Tags:          []string{"hostile", "collidable", "animated"},
		ActiveEffects: []string{},
	}
}

// NewProjectileMetadata creates metadata for projectile entities.
func NewProjectileMetadata(description string) AIMetadata {
	return AIMetadata{
		EntityType:    "projectile",
		Description:   description,
		VisualState:   "active",
		Tags:          []string{"collidable", "temporary"},
		ActiveEffects: []string{},
	}
}

// NewUIMetadata creates metadata for UI elements.
func NewUIMetadata(description string) AIMetadata {
	return AIMetadata{
		EntityType:    "ui",
		Description:   description,
		VisualState:   "visible",
		Tags:          []string{"interactable"},
		ActiveEffects: []string{},
	}
}

// NewEffectMetadata creates metadata for visual effects.
func NewEffectMetadata(effectType, description string) AIMetadata {
	return AIMetadata{
		EntityType:    "effect",
		Description:   description,
		VisualState:   "playing",
		Tags:          []string{"temporary", effectType},
		ActiveEffects: []string{},
	}
}

// NewItemMetadata creates metadata for collectible items.
func NewItemMetadata(description string) AIMetadata {
	return AIMetadata{
		EntityType:    "item",
		Description:   description,
		VisualState:   "idle",
		Tags:          []string{"collectible", "collidable"},
		ActiveEffects: []string{},
	}
}

// SetState updates the visual state.
func (m *AIMetadata) SetState(state string) {
	m.VisualState = state
}

// AddTag adds a tag if not already present.
func (m *AIMetadata) AddTag(tag string) {
	if slices.Contains(m.Tags, tag) {
		return
	}

	m.Tags = append(m.Tags, tag)
}

// RemoveTag removes a tag.
func (m *AIMetadata) RemoveTag(tag string) {
	for i, t := range m.Tags {
		if t == tag {
			m.Tags = append(m.Tags[:i], m.Tags[i+1:]...)

			return
		}
	}
}

// HasTag checks if a tag is present.
func (m *AIMetadata) HasTag(tag string) bool {
	return slices.Contains(m.Tags, tag)
}

// AddEffect adds an active effect.
func (m *AIMetadata) AddEffect(effect string) {
	if slices.Contains(m.ActiveEffects, effect) {
		return
	}

	m.ActiveEffects = append(m.ActiveEffects, effect)
}

// RemoveEffect removes an active effect.
func (m *AIMetadata) RemoveEffect(effect string) {
	for i, e := range m.ActiveEffects {
		if e == effect {
			m.ActiveEffects = append(m.ActiveEffects[:i], m.ActiveEffects[i+1:]...)

			return
		}
	}
}
