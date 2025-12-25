package systems

import (
	"td/internal/components"
)

// StatusSystem manages status effect updates.
type StatusSystem struct {
	// In a real ECS these would be queried.
	// Here we provide methods to update components.
}

func NewStatusSystem() *StatusSystem {
	return &StatusSystem{}
}

// UpdateComponent updates a single status component.
// It returns a list of "Ticks" processing if needed, or we can pass a handler?
// Let's make it simple: It manages time.
// dt: delta time in seconds
func (s *StatusSystem) UpdateComponent(sc *components.StatusComponent, dt float64) {
	// 1. Expire old effects
	sc.RemoveExpired(dt)

	// 2. Process Ticks for remaining effects
	for _, e := range sc.Effects {
		if e.Interval > 0 {
			// Process tick - caller should handle DOT damage via UpdateAndApply
			_ = e.UpdateTick(dt)
		}
	}
}

// ApplyDamageTick is a helper if you want to apply DOTs to health directly.
// You would call this in your game loop if you have both components.
type HealthComponent interface {
	TakeDamage(amount int)
}

// UpdateAndApply processes time and applies DOT damage if health is provided.
func (s *StatusSystem) UpdateAndApply(sc *components.StatusComponent, health HealthComponent, dt float64) {
	// 1. Expire
	sc.RemoveExpired(dt)

	// 2. Tick
	for _, e := range sc.Effects {
		if e.Interval > 0 {
			if e.UpdateTick(dt) {
				if health != nil && e.DamagePerTick > 0 {
					health.TakeDamage(e.DamagePerTick)
				}
			}
		}
	}
}
