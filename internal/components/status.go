package components

// StatusType represents the type of ailment or buff.
type StatusType string

const (
	StatusIgnite StatusType = "ignite" // Fire DOT
	StatusChill  StatusType = "chill"  // Slow
	StatusFreeze StatusType = "freeze" // Stun / Stop Action
	StatusShock  StatusType = "shock"  // Increased Damage Taken
	StatusPoison StatusType = "poison" // Stacking DOT (Chaos)
	StatusBleed  StatusType = "bleed"  // Physical DOT
	StatusBuff   StatusType = "buff"   // Generic Buff
)

// StatusEffect represents a single instance of an active effect.
type StatusEffect struct {
	Type        StatusType
	Duration    float64 // Seconds remaining
	MaxDuration float64
	Interval    float64 // Seconds between ticks (0 for non-ticking)
	TickTimer   float64
	Magnitude   float64 // Strength of effect (e.g., 0.3 for 30% slow)
	Stacks      int     // Current stacks (if aggregated)
	Stackable   bool    // If true, multiple instances can exist (PoE Poison)
	SourceID    int     // ID of the entity that applied this (for attribution)

	// OnTick is a callback for handling damage or events.
	// Note: In pure ECS, logic shouldn't be here, but for practical Go engines,
	// having a decoupled data-logic handler or simple flag is often used.
	// We will rely on the System to interpret 'Magnitude' for standard ailments,
	// but keeping the struct simple data is best.
	DamagePerTick int
}

// StatusComponent holds all active status effects on an entity.
type StatusComponent struct {
	// A slice allows multiple instances of the same type (like multiple Poison stacks).
	Effects []*StatusEffect
}

func NewStatusComponent() *StatusComponent {
	return &StatusComponent{
		Effects: make([]*StatusEffect, 0),
	}
}

// AddEffect applies a new status effect.
// behavior:
// - Stackable=true (Poison): Adds a new independent instance.
// - Stackable=false (Ignite): Replaces existing if new Magnitude is higher (or refreshes duration).
func (sc *StatusComponent) AddEffect(effect *StatusEffect) {
	if effect.Stackable {
		sc.Effects = append(sc.Effects, effect)

		return
	}

	// For non-stackable, check if one exists
	for i, e := range sc.Effects {
		if e.Type == effect.Type {
			// Rule: Usually keep highest magnitude
			if effect.Magnitude >= e.Magnitude {
				// Replace with stronger effect
				sc.Effects[i] = effect
			}
			// If incoming is weaker, keep existing (stronger wins)

			return
		}
	}
	// Not found, add it
	sc.Effects = append(sc.Effects, effect)
}

// HasStatus checks if a specific status type is active.
func (sc *StatusComponent) HasStatus(t StatusType) bool {
	for _, e := range sc.Effects {
		if e.Type == t {
			return true
		}
	}

	return false
}

// GetStatusMagnitude returns the strongest magnitude of a status type (e.g. strongest Chill).
func (sc *StatusComponent) GetStatusMagnitude(t StatusType) float64 {
	maxMag := 0.0

	for _, e := range sc.Effects {
		if e.Type == t {
			if e.Magnitude > maxMag {
				maxMag = e.Magnitude
			}
		}
	}

	return maxMag
}

// RemoveExpireEffects cleans up expired effects.
// Returns list of expired effects if needed (e.g., for cleanup visuals).
func (sc *StatusComponent) RemoveExpired(dt float64) {
	active := sc.Effects[:0]
	for _, e := range sc.Effects {
		e.Duration -= dt
		if e.Duration > 0 {
			active = append(active, e)
		}
	}

	sc.Effects = active
}

// UpdateTicks updates tick timers and returns true if a tick occurred for this effect.
func (e *StatusEffect) UpdateTick(dt float64) bool {
	if e.Interval <= 0 {
		return false
	}

	e.TickTimer += dt
	if e.TickTimer >= e.Interval {
		e.TickTimer -= e.Interval

		return true
	}

	return false
}
