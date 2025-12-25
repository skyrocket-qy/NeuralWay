package systems

import (
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// DamageEvent represents damage dealt to an entity.
type DamageEvent struct {
	Target    ecs.Entity
	Source    ecs.Entity
	Amount    float64
	Type      components.DamageType
	IsCrit    bool
	WasLethal bool
}

// HealEvent represents healing received by an entity.
type HealEvent struct {
	Target ecs.Entity
	Source ecs.Entity
	Amount float64
}

// DeathEvent represents an entity death.
type DeathEvent struct {
	Entity ecs.Entity
	Killer ecs.Entity
}

// HealthSystem manages health, damage, and healing.
type HealthSystem struct {
	healthFilter *ecs.Filter1[components.Health]
	damageQueue  []DamageEvent
	healQueue    []HealEvent
	deathQueue   []DeathEvent
	onDamage     func(DamageEvent)
	onHeal       func(HealEvent)
	onDeath      func(DeathEvent)
	resistances  map[ecs.Entity]map[components.DamageType]float64
}

// NewHealthSystem creates a health management system.
func NewHealthSystem(world *ecs.World) *HealthSystem {
	return &HealthSystem{
		healthFilter: ecs.NewFilter1[components.Health](world),
		damageQueue:  make([]DamageEvent, 0),
		healQueue:    make([]HealEvent, 0),
		deathQueue:   make([]DeathEvent, 0),
		resistances:  make(map[ecs.Entity]map[components.DamageType]float64),
	}
}

// SetOnDamage sets the damage callback.
func (s *HealthSystem) SetOnDamage(fn func(DamageEvent)) {
	s.onDamage = fn
}

// SetOnHeal sets the heal callback.
func (s *HealthSystem) SetOnHeal(fn func(HealEvent)) {
	s.onHeal = fn
}

// SetOnDeath sets the death callback.
func (s *HealthSystem) SetOnDeath(fn func(DeathEvent)) {
	s.onDeath = fn
}

// SetResistance sets damage resistance for an entity.
func (s *HealthSystem) SetResistance(
	entity ecs.Entity,
	damageType components.DamageType,
	resistance float64,
) {
	if s.resistances[entity] == nil {
		s.resistances[entity] = make(map[components.DamageType]float64)
	}

	s.resistances[entity][damageType] = resistance
}

// QueueDamage queues damage to be applied on next update.
func (s *HealthSystem) QueueDamage(
	target, source ecs.Entity,
	amount float64,
	damageType components.DamageType,
	isCrit bool,
) {
	s.damageQueue = append(s.damageQueue, DamageEvent{
		Target: target,
		Source: source,
		Amount: amount,
		Type:   damageType,
		IsCrit: isCrit,
	})
}

// QueueHeal queues healing to be applied on next update.
func (s *HealthSystem) QueueHeal(target, source ecs.Entity, amount float64) {
	s.healQueue = append(s.healQueue, HealEvent{
		Target: target,
		Source: source,
		Amount: amount,
	})
}

// GetDeaths returns entities that died since last update.
func (s *HealthSystem) GetDeaths() []DeathEvent {
	return s.deathQueue
}

// ClearDeaths clears the death queue.
func (s *HealthSystem) ClearDeaths() {
	s.deathQueue = s.deathQueue[:0]
}

// Update processes damage, healing, and death.
func (s *HealthSystem) Update(world *ecs.World) {
	s.deathQueue = s.deathQueue[:0]

	// Build entity->health lookup
	healthMap := make(map[ecs.Entity]*components.Health)

	query := s.healthFilter.Query()
	for query.Next() {
		health := query.Get()
		entity := query.Entity()
		healthMap[entity] = health
	}

	// Process damage
	for i := range s.damageQueue {
		event := &s.damageQueue[i]

		health, ok := healthMap[event.Target]
		if !ok {
			continue
		}

		// Apply resistance
		finalDamage := event.Amount
		if res, ok := s.resistances[event.Target]; ok {
			if r, ok := res[event.Type]; ok {
				finalDamage *= (1.0 - r)
			}
		}

		// Apply damage
		intDamage := int(finalDamage)
		health.Current -= intDamage

		if health.Current <= 0 {
			health.Current = 0
			event.WasLethal = true
			s.deathQueue = append(s.deathQueue, DeathEvent{
				Entity: event.Target,
				Killer: event.Source,
			})
		}

		if s.onDamage != nil {
			s.onDamage(*event)
		}
	}

	s.damageQueue = s.damageQueue[:0]

	// Process healing
	for i := range s.healQueue {
		event := &s.healQueue[i]

		health, ok := healthMap[event.Target]
		if !ok {
			continue
		}

		intHeal := int(event.Amount)

		health.Current += intHeal
		if health.Current > health.Max {
			health.Current = health.Max
		}

		if s.onHeal != nil {
			s.onHeal(*event)
		}
	}

	s.healQueue = s.healQueue[:0]

	// Invoke death callbacks
	for _, death := range s.deathQueue {
		if s.onDeath != nil {
			s.onDeath(death)
		}
	}
}

// ApplyDamageImmediate applies damage immediately without queueing.
func (s *HealthSystem) ApplyDamageImmediate(
	world *ecs.World,
	target, source ecs.Entity,
	amount float64,
	damageType components.DamageType,
) bool {
	query := s.healthFilter.Query()
	for query.Next() {
		entity := query.Entity()
		if entity != target {
			continue
		}

		health := query.Get()

		// Apply resistance
		finalDamage := amount

		if res, ok := s.resistances[target]; ok {
			if r, ok := res[damageType]; ok {
				finalDamage *= (1.0 - r)
			}
		}

		health.Current -= int(finalDamage)
		if health.Current < 0 {
			health.Current = 0
		}

		return health.Current <= 0
	}

	return false
}

// HealImmediate applies healing immediately.
func (s *HealthSystem) HealImmediate(world *ecs.World, target ecs.Entity, amount float64) {
	query := s.healthFilter.Query()
	for query.Next() {
		entity := query.Entity()
		if entity != target {
			continue
		}

		health := query.Get()

		health.Current += int(amount)
		if health.Current > health.Max {
			health.Current = health.Max
		}

		return
	}
}

// IsAlive checks if an entity is alive.
func (s *HealthSystem) IsAlive(world *ecs.World, target ecs.Entity) bool {
	query := s.healthFilter.Query()
	for query.Next() {
		entity := query.Entity()
		if entity == target {
			health := query.Get()

			return health.Current > 0
		}
	}

	return false
}
