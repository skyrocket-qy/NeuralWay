package systems

import (
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// CooldownReadyEvent represents an ability coming off cooldown.
type CooldownReadyEvent struct {
	Entity    ecs.Entity
	AbilityID string
}

// CooldownSystem manages ability cooldowns.
type CooldownSystem struct {
	cooldownFilter *ecs.Filter1[components.Cooldown]
	manaFilter     *ecs.Filter1[components.Mana]
	readyQueue     []CooldownReadyEvent
	onReady        func(CooldownReadyEvent)
}

// NewCooldownSystem creates a cooldown system.
func NewCooldownSystem(world *ecs.World) *CooldownSystem {
	return &CooldownSystem{
		cooldownFilter: ecs.NewFilter1[components.Cooldown](world),
		manaFilter:     ecs.NewFilter1[components.Mana](world),
		readyQueue:     make([]CooldownReadyEvent, 0),
	}
}

// SetOnReady sets the callback for when abilities become ready.
func (s *CooldownSystem) SetOnReady(fn func(CooldownReadyEvent)) {
	s.onReady = fn
}

// Update updates all cooldowns and mana regeneration.
func (s *CooldownSystem) Update(world *ecs.World, dt float64) {
	s.readyQueue = s.readyQueue[:0]

	// Update cooldowns
	query := s.cooldownFilter.Query()
	for query.Next() {
		cd := query.Get()
		entity := query.Entity()

		wasOnCooldown := !cd.IsReady()

		// Update main cooldown
		if cd.Remaining > 0 {
			cd.Remaining -= dt
			if cd.Remaining < 0 {
				cd.Remaining = 0
			}
		}

		// Update charge regeneration
		if cd.MaxCharges > 1 && cd.Charges < cd.MaxCharges {
			cd.ChargeTimer += dt
			if cd.ChargeTimer >= cd.ChargeTime {
				cd.ChargeTimer -= cd.ChargeTime
				cd.Charges++
			}
		}

		// Check if just became ready
		if wasOnCooldown && cd.IsReady() {
			event := CooldownReadyEvent{Entity: entity}

			s.readyQueue = append(s.readyQueue, event)
			if s.onReady != nil {
				s.onReady(event)
			}
		}
	}

	// Update mana regeneration
	manaQuery := s.manaFilter.Query()
	for manaQuery.Next() {
		mana := manaQuery.Get()

		// Update delay timer
		if mana.DelayTimer > 0 {
			mana.DelayTimer -= dt

			continue
		}

		// Regenerate mana
		mana.Current += mana.Regen * dt
		if mana.Current > mana.Max {
			mana.Current = mana.Max
		}
	}
}

// IsReady checks if an entity's cooldown is ready.
func (s *CooldownSystem) IsReady(world *ecs.World, entity ecs.Entity) bool {
	query := s.cooldownFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			cd := query.Get()

			return cd.IsReady()
		}
	}

	return false
}

// UseAbility attempts to use an ability with cooldown and mana cost.
func (s *CooldownSystem) UseAbility(world *ecs.World, entity ecs.Entity, manaCost float64) bool {
	var (
		cd   *components.Cooldown
		mana *components.Mana
	)

	// Get cooldown
	cdQuery := s.cooldownFilter.Query()
	for cdQuery.Next() {
		e := cdQuery.Entity()
		if e == entity {
			cd = cdQuery.Get()

			break
		}
	}

	// Get mana
	manaQuery := s.manaFilter.Query()
	for manaQuery.Next() {
		e := manaQuery.Entity()
		if e == entity {
			mana = manaQuery.Get()

			break
		}
	}

	// Check cooldown
	if cd != nil && !cd.IsReady() {
		return false
	}

	// Check mana
	if mana != nil && manaCost > 0 && !mana.Use(manaCost) {
		return false
	}

	// Trigger cooldown
	if cd != nil {
		cd.Use()
	}

	return true
}

// GetCooldownProgress returns progress through cooldown (0.0 = just used, 1.0 = ready).
func (s *CooldownSystem) GetCooldownProgress(world *ecs.World, entity ecs.Entity) float64 {
	query := s.cooldownFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			cd := query.Get()
			if cd.Duration == 0 || cd.Remaining == 0 {
				return 1.0
			}

			return 1.0 - (cd.Remaining / cd.Duration)
		}
	}

	return 1.0
}

// GetMana returns current and max mana for an entity.
func (s *CooldownSystem) GetMana(world *ecs.World, entity ecs.Entity) (current, maxVal float64) {
	query := s.manaFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			mana := query.Get()

			return mana.Current, mana.Max
		}
	}

	return 0, 0
}

// RestoreMana adds mana to an entity.
func (s *CooldownSystem) RestoreMana(world *ecs.World, entity ecs.Entity, amount float64) {
	query := s.manaFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			mana := query.Get()

			mana.Current += amount
			if mana.Current > mana.Max {
				mana.Current = mana.Max
			}

			return
		}
	}
}

// ReduceCooldown reduces remaining cooldown time.
func (s *CooldownSystem) ReduceCooldown(world *ecs.World, entity ecs.Entity, amount float64) {
	query := s.cooldownFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			cd := query.Get()

			cd.Remaining -= amount
			if cd.Remaining < 0 {
				cd.Remaining = 0
			}

			return
		}
	}
}

// ResetCooldown resets cooldown to ready state.
func (s *CooldownSystem) ResetCooldown(world *ecs.World, entity ecs.Entity) {
	query := s.cooldownFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			cd := query.Get()
			cd.Remaining = 0
			cd.Charges = cd.MaxCharges

			return
		}
	}
}
