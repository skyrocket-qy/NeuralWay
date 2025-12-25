package systems

import (
	"math"
	"sort"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// Aggro component for entities with threat tables.
type Aggro struct {
	Threats         map[ecs.Entity]float64 // Entity -> threat value
	CurrentTarget   ecs.Entity
	AggroRange      float64 // Range to detect enemies
	LeashRange      float64 // Range before returning to spawn
	LeashX, LeashY  float64 // Spawn/leash point
	Leashed         bool    // Currently returning to spawn
	ThreatDecay     float64 // Threat decay per second
	MaxThreatMemory int     // Max entities to track
}

// NewAggro creates an aggro component.
func NewAggro(aggroRange, leashRange float64) Aggro {
	return Aggro{
		Threats:         make(map[ecs.Entity]float64),
		AggroRange:      aggroRange,
		LeashRange:      leashRange,
		ThreatDecay:     1.0,
		MaxThreatMemory: 10,
	}
}

// ThreatModifier represents modifiers to threat generation.
type ThreatModifier struct {
	DamageMult  float64 // Multiplier for damage threat
	HealingMult float64 // Multiplier for healing threat
	TauntMult   float64 // Multiplier for taunt effects
	BaseThreat  float64 // Flat threat per second
}

// NewThreatModifier creates default threat modifiers.
func NewThreatModifier() ThreatModifier {
	return ThreatModifier{
		DamageMult:  1.0,
		HealingMult: 0.5,
		TauntMult:   1.5,
	}
}

// AggroEvent represents a threat-generating event.
type AggroEvent struct {
	Source  ecs.Entity
	Target  ecs.Entity // Entity with Aggro component
	Amount  float64
	IsTaunt bool
}

// AggroSystem manages enemy targeting based on threat.
type AggroSystem struct {
	aggroFilter    *ecs.Filter2[components.Position, Aggro]
	targetFilter   *ecs.Filter1[components.Position]
	eventQueue     []AggroEvent
	onTargetChange func(entity, oldTarget, newTarget ecs.Entity)
	threatMods     map[ecs.Entity]ThreatModifier
}

// NewAggroSystem creates an aggro system.
func NewAggroSystem(world *ecs.World) *AggroSystem {
	return &AggroSystem{
		aggroFilter:  ecs.NewFilter2[components.Position, Aggro](world),
		targetFilter: ecs.NewFilter1[components.Position](world),
		eventQueue:   make([]AggroEvent, 0),
		threatMods:   make(map[ecs.Entity]ThreatModifier),
	}
}

// SetOnTargetChange sets the target change callback.
func (s *AggroSystem) SetOnTargetChange(fn func(entity, oldTarget, newTarget ecs.Entity)) {
	s.onTargetChange = fn
}

// SetThreatModifier sets threat modifiers for an entity.
func (s *AggroSystem) SetThreatModifier(entity ecs.Entity, mod ThreatModifier) {
	s.threatMods[entity] = mod
}

// AddThreat queues a threat event.
func (s *AggroSystem) AddThreat(source, target ecs.Entity, amount float64, isTaunt bool) {
	s.eventQueue = append(s.eventQueue, AggroEvent{
		Source:  source,
		Target:  target,
		Amount:  amount,
		IsTaunt: isTaunt,
	})
}

// AddDamageThreat adds threat from damage dealt.
func (s *AggroSystem) AddDamageThreat(source, target ecs.Entity, damage float64) {
	mult := 1.0
	if mod, ok := s.threatMods[source]; ok {
		mult = mod.DamageMult
	}

	s.AddThreat(source, target, damage*mult, false)
}

// AddHealingThreat adds threat from healing done.
func (s *AggroSystem) AddHealingThreat(source, target ecs.Entity, healing float64) {
	mult := 0.5
	if mod, ok := s.threatMods[source]; ok {
		mult = mod.HealingMult
	}

	s.AddThreat(source, target, healing*mult, false)
}

// Taunt forces an entity to target the taunter.
func (s *AggroSystem) Taunt(taunter, target ecs.Entity) {
	s.AddThreat(taunter, target, 9999, true)
}

// Update processes threat events and updates targets.
func (s *AggroSystem) Update(world *ecs.World, dt float64) {
	// Build position lookup
	positions := make(map[ecs.Entity]*components.Position)

	posQuery := s.targetFilter.Query()
	for posQuery.Next() {
		positions[posQuery.Entity()] = posQuery.Get()
	}

	// Process events
	for _, event := range s.eventQueue {
		s.processEvent(event)
	}

	s.eventQueue = s.eventQueue[:0]

	// Update each entity with aggro
	query := s.aggroFilter.Query()
	for query.Next() {
		pos, aggro := query.Get()
		entity := query.Entity()

		// Decay threat
		for e := range aggro.Threats {
			aggro.Threats[e] -= aggro.ThreatDecay * dt
			if aggro.Threats[e] <= 0 {
				delete(aggro.Threats, e)
			}
		}

		// Check leash
		if aggro.LeashRange > 0 {
			dx := pos.X - aggro.LeashX
			dy := pos.Y - aggro.LeashY
			distFromSpawn := math.Sqrt(dx*dx + dy*dy)

			if distFromSpawn > aggro.LeashRange {
				aggro.Leashed = true
				aggro.CurrentTarget = ecs.Entity{}
				aggro.Threats = make(map[ecs.Entity]float64)

				continue
			}

			if aggro.Leashed && distFromSpawn < aggro.AggroRange/2 {
				aggro.Leashed = false
			}
		}

		if aggro.Leashed {
			continue
		}

		// Scan for new threats in range
		for e, ePos := range positions {
			if e == entity {
				continue
			}

			dx := ePos.X - pos.X
			dy := ePos.Y - pos.Y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= aggro.AggroRange {
				// Add to threat table if not present
				if _, ok := aggro.Threats[e]; !ok {
					aggro.Threats[e] = 1.0 // Base aggro from proximity
				}
			}
		}

		// Remove threats that are too far
		for e := range aggro.Threats {
			if ePos, ok := positions[e]; ok {
				dx := ePos.X - pos.X
				dy := ePos.Y - pos.Y

				dist := math.Sqrt(dx*dx + dy*dy)
				if dist > aggro.LeashRange {
					delete(aggro.Threats, e)
				}
			} else {
				delete(aggro.Threats, e) // Entity no longer exists
			}
		}

		// Limit threat table size
		if len(aggro.Threats) > aggro.MaxThreatMemory {
			s.trimThreatTable(aggro)
		}

		// Select highest threat target
		oldTarget := aggro.CurrentTarget
		newTarget := s.getHighestThreat(aggro)

		if newTarget != oldTarget {
			aggro.CurrentTarget = newTarget
			if s.onTargetChange != nil {
				s.onTargetChange(entity, oldTarget, newTarget)
			}
		}
	}
}

// processEvent applies a threat event.
func (s *AggroSystem) processEvent(event AggroEvent) {
	query := s.aggroFilter.Query()
	for query.Next() {
		if query.Entity() == event.Target {
			_, aggro := query.Get()

			amount := event.Amount
			if event.IsTaunt {
				// Taunt sets threat to max + bonus
				maxThreat := 0.0
				for _, t := range aggro.Threats {
					if t > maxThreat {
						maxThreat = t
					}
				}

				amount = maxThreat + amount
			}

			aggro.Threats[event.Source] += amount

			return
		}
	}
}

// getHighestThreat returns the entity with highest threat.
func (s *AggroSystem) getHighestThreat(aggro *Aggro) ecs.Entity {
	var highest ecs.Entity

	maxThreat := 0.0

	for e, threat := range aggro.Threats {
		if threat > maxThreat {
			maxThreat = threat
			highest = e
		}
	}

	return highest
}

// trimThreatTable removes lowest threats to stay within limit.
func (s *AggroSystem) trimThreatTable(aggro *Aggro) {
	type threatEntry struct {
		entity ecs.Entity
		threat float64
	}

	entries := make([]threatEntry, 0, len(aggro.Threats))
	for e, t := range aggro.Threats {
		entries = append(entries, threatEntry{e, t})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].threat > entries[j].threat
	})

	aggro.Threats = make(map[ecs.Entity]float64)
	for i := 0; i < aggro.MaxThreatMemory && i < len(entries); i++ {
		aggro.Threats[entries[i].entity] = entries[i].threat
	}
}

// GetCurrentTarget returns the current target for an entity.
func (s *AggroSystem) GetCurrentTarget(world *ecs.World, entity ecs.Entity) ecs.Entity {
	query := s.aggroFilter.Query()
	for query.Next() {
		if query.Entity() == entity {
			_, aggro := query.Get()

			return aggro.CurrentTarget
		}
	}

	return ecs.Entity{}
}

// GetThreatList returns sorted threat list for an entity.
func (s *AggroSystem) GetThreatList(world *ecs.World, entity ecs.Entity) []ecs.Entity {
	query := s.aggroFilter.Query()
	for query.Next() {
		if query.Entity() == entity {
			_, aggro := query.Get()

			type entry struct {
				e ecs.Entity
				t float64
			}

			entries := make([]entry, 0, len(aggro.Threats))
			for e, t := range aggro.Threats {
				entries = append(entries, entry{e, t})
			}

			sort.Slice(entries, func(i, j int) bool {
				return entries[i].t > entries[j].t
			})

			result := make([]ecs.Entity, len(entries))
			for i, e := range entries {
				result[i] = e.e
			}

			return result
		}
	}

	return nil
}

// ClearThreat clears all threat for an entity.
func (s *AggroSystem) ClearThreat(world *ecs.World, entity ecs.Entity) {
	query := s.aggroFilter.Query()
	for query.Next() {
		if query.Entity() == entity {
			_, aggro := query.Get()
			aggro.Threats = make(map[ecs.Entity]float64)
			aggro.CurrentTarget = ecs.Entity{}

			return
		}
	}
}

// DropThreat drops a percentage of threat for a source entity.
func (s *AggroSystem) DropThreat(world *ecs.World, source ecs.Entity, percent float64) {
	query := s.aggroFilter.Query()
	for query.Next() {
		_, aggro := query.Get()
		if threat, ok := aggro.Threats[source]; ok {
			aggro.Threats[source] = threat * (1.0 - percent)
		}
	}
}
