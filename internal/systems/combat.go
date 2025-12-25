package systems

import (
	"math/rand"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// AttackEvent represents an attack action.
type AttackEvent struct {
	Attacker   ecs.Entity
	Target     ecs.Entity
	Damage     float64
	DamageType components.DamageType
	IsCrit     bool
	Timestamp  float64
}

// CombatSystem handles combat between entities.
type CombatSystem struct {
	combatFilter *ecs.Filter1[components.Combat]
	critFilter   *ecs.Filter1[components.CriticalHit]
	buffFilter   *ecs.Filter1[components.BuffContainer]
	attackQueue  []AttackEvent
	onAttack     func(AttackEvent)
	healthSystem *HealthSystem
	currentTime  float64
}

// NewCombatSystem creates a combat system.
func NewCombatSystem(world *ecs.World, healthSystem *HealthSystem) *CombatSystem {
	return &CombatSystem{
		combatFilter: ecs.NewFilter1[components.Combat](world),
		critFilter:   ecs.NewFilter1[components.CriticalHit](world),
		buffFilter:   ecs.NewFilter1[components.BuffContainer](world),
		attackQueue:  make([]AttackEvent, 0),
		healthSystem: healthSystem,
	}
}

// SetOnAttack sets the attack callback.
func (s *CombatSystem) SetOnAttack(fn func(AttackEvent)) {
	s.onAttack = fn
}

// CanAttack checks if an entity can attack.
func (s *CombatSystem) CanAttack(world *ecs.World, attacker ecs.Entity) bool {
	query := s.combatFilter.Query()
	for query.Next() {
		entity := query.Entity()
		if entity != attacker {
			continue
		}

		combat := query.Get()
		if !combat.CanAttack {
			return false
		}

		attackCooldown := 1.0 / combat.AttackSpeed

		return s.currentTime-combat.LastAttackAt >= attackCooldown
	}

	return false
}

// Attack performs an attack from attacker to target.
func (s *CombatSystem) Attack(world *ecs.World, attacker, target ecs.Entity) bool {
	if !s.CanAttack(world, attacker) {
		return false
	}

	var (
		attackerCombat *components.Combat
		attackerCrit   *components.CriticalHit
		attackerBuffs  *components.BuffContainer
	)

	// Get attacker components
	query := s.combatFilter.Query()
	for query.Next() {
		entity := query.Entity()
		if entity == attacker {
			attackerCombat = query.Get()

			break
		}
	}

	if attackerCombat == nil {
		return false
	}

	// Get crit component if exists
	critQuery := s.critFilter.Query()
	for critQuery.Next() {
		entity := critQuery.Entity()
		if entity == attacker {
			attackerCrit = critQuery.Get()

			break
		}
	}

	// Get buff container if exists
	buffQuery := s.buffFilter.Query()
	for buffQuery.Next() {
		entity := buffQuery.Entity()
		if entity == attacker {
			attackerBuffs = buffQuery.Get()

			break
		}
	}

	// Calculate damage
	baseDamage := attackerCombat.AttackPower

	// Apply buffs
	if attackerBuffs != nil {
		add, mult := attackerBuffs.GetModifier("attack")
		baseDamage = (baseDamage + add) * mult
	}

	// Check for crit
	isCrit := false

	if attackerCrit != nil {
		if attackerCrit.Guaranteed || rand.Float64() < attackerCrit.Chance {
			baseDamage *= attackerCrit.Multiplier
			isCrit = true
			attackerCrit.Guaranteed = false
		}
	}

	// Update attack timestamp
	attackerCombat.LastAttackAt = s.currentTime

	// Queue damage
	event := AttackEvent{
		Attacker:   attacker,
		Target:     target,
		Damage:     baseDamage,
		DamageType: attackerCombat.DamageType,
		IsCrit:     isCrit,
		Timestamp:  s.currentTime,
	}

	if s.healthSystem != nil {
		s.healthSystem.QueueDamage(target, attacker, baseDamage, attackerCombat.DamageType, isCrit)
	}

	if s.onAttack != nil {
		s.onAttack(event)
	}

	return true
}

// GetCombatStats returns combat stats for an entity.
func (s *CombatSystem) GetCombatStats(world *ecs.World, entity ecs.Entity) *components.Combat {
	query := s.combatFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			return query.Get()
		}
	}

	return nil
}

// SetTime updates the current time for cooldown calculations.
func (s *CombatSystem) SetTime(time float64) {
	s.currentTime = time
}

// Update updates combat state.
func (s *CombatSystem) Update(world *ecs.World, dt float64) {
	s.currentTime += dt

	// Update buff durations
	query := s.buffFilter.Query()
	for query.Next() {
		buffs := query.Get()
		for i := range buffs.Buffs {
			buffs.Buffs[i].Duration -= dt
		}

		buffs.RemoveExpired()
	}
}

// ApplyBuff applies a buff to an entity.
func (s *CombatSystem) ApplyBuff(world *ecs.World, target ecs.Entity, buff components.Buff) bool {
	query := s.buffFilter.Query()
	for query.Next() {
		entity := query.Entity()
		if entity == target {
			buffs := query.Get()
			buffs.AddBuff(buff)

			return true
		}
	}

	return false
}

// CalculateDamage calculates final damage with defense.
func CalculateDamage(baseDamage, defense float64) float64 {
	// Simple damage reduction formula
	reduction := defense / (defense + 100)

	return baseDamage * (1 - reduction)
}

// CalculateDamageWithPenetration calculates damage with defense penetration.
func CalculateDamageWithPenetration(baseDamage, defense, penetration float64) float64 {
	effectiveDefense := defense * (1 - penetration)
	if effectiveDefense < 0 {
		effectiveDefense = 0
	}

	return CalculateDamage(baseDamage, effectiveDefense)
}
