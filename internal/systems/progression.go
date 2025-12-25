package systems

import (
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// LevelUpEvent represents a level up.
type LevelUpEvent struct {
	Entity   ecs.Entity
	OldLevel int
	NewLevel int
}

// ProgressionSystem handles XP and leveling.
type ProgressionSystem struct {
	expFilter    *ecs.Filter1[components.Experience]
	levelFilter  *ecs.Filter1[components.Level]
	levelUpQueue []LevelUpEvent
	onLevelUp    func(LevelUpEvent)
}

// NewProgressionSystem creates a progression system.
func NewProgressionSystem(world *ecs.World) *ProgressionSystem {
	return &ProgressionSystem{
		expFilter:    ecs.NewFilter1[components.Experience](world),
		levelFilter:  ecs.NewFilter1[components.Level](world),
		levelUpQueue: make([]LevelUpEvent, 0),
	}
}

// SetOnLevelUp sets the level up callback.
func (s *ProgressionSystem) SetOnLevelUp(fn func(LevelUpEvent)) {
	s.onLevelUp = fn
}

// GetLevelUps returns level ups since last clear.
func (s *ProgressionSystem) GetLevelUps() []LevelUpEvent {
	return s.levelUpQueue
}

// ClearLevelUps clears the level up queue.
func (s *ProgressionSystem) ClearLevelUps() {
	s.levelUpQueue = s.levelUpQueue[:0]
}

// AddExperience adds XP to an entity and checks for level up.
func (s *ProgressionSystem) AddExperience(world *ecs.World, entity ecs.Entity, amount int64) {
	var (
		exp   *components.Experience
		level *components.Level
	)

	// Get experience
	expQuery := s.expFilter.Query()
	for expQuery.Next() {
		e := expQuery.Entity()
		if e == entity {
			exp = expQuery.Get()

			break
		}
	}

	if exp == nil {
		return
	}

	// Get level
	levelQuery := s.levelFilter.Query()
	for levelQuery.Next() {
		e := levelQuery.Entity()
		if e == entity {
			level = levelQuery.Get()

			break
		}
	}

	if level == nil {
		return
	}

	// Add XP
	exp.Current += amount
	exp.Total += amount

	// Check for level up
	for exp.Current >= exp.ToNextLevel(level.Current) && level.CanLevelUp() {
		exp.Current -= exp.ToNextLevel(level.Current)
		oldLevel := level.Current
		level.LevelUp()

		event := LevelUpEvent{
			Entity:   entity,
			OldLevel: oldLevel,
			NewLevel: level.Current,
		}
		s.levelUpQueue = append(s.levelUpQueue, event)

		if s.onLevelUp != nil {
			s.onLevelUp(event)
		}
	}
}

// GetLevel returns the current level for an entity.
func (s *ProgressionSystem) GetLevel(world *ecs.World, entity ecs.Entity) int {
	query := s.levelFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			level := query.Get()

			return level.Current
		}
	}

	return 0
}

// GetExperience returns current XP for an entity.
func (s *ProgressionSystem) GetExperience(world *ecs.World, entity ecs.Entity) (current, toNext int64) {
	var (
		exp   *components.Experience
		level *components.Level
	)

	expQuery := s.expFilter.Query()
	for expQuery.Next() {
		e := expQuery.Entity()
		if e == entity {
			exp = expQuery.Get()

			break
		}
	}

	levelQuery := s.levelFilter.Query()
	for levelQuery.Next() {
		e := levelQuery.Entity()
		if e == entity {
			level = levelQuery.Get()

			break
		}
	}

	if exp != nil && level != nil {
		return exp.Current, exp.ToNextLevel(level.Current)
	}

	return 0, 0
}

// GetProgress returns XP progress as a ratio (0.0 - 1.0).
func (s *ProgressionSystem) GetProgress(world *ecs.World, entity ecs.Entity) float64 {
	current, toNext := s.GetExperience(world, entity)
	if toNext == 0 {
		return 0
	}

	return float64(current) / float64(toNext)
}

// SpendSkillPoints attempts to unlock or upgrade a skill node.
func (s *ProgressionSystem) SpendSkillPoints(
	world *ecs.World,
	entity ecs.Entity,
	tree *components.SkillTree,
	nodeID string,
) bool {
	var level *components.Level

	query := s.levelFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			level = query.Get()

			break
		}
	}

	if level == nil {
		return false
	}

	node, ok := tree.Nodes[nodeID]
	if !ok {
		return false
	}

	unlocked := tree.GetUnlocked()
	if !node.CanUnlock(unlocked, level.SkillPoints) {
		return false
	}

	level.SkillPoints -= node.Cost

	node.CurrentRank++
	if node.CurrentRank == 1 {
		node.Unlocked = true
	}

	return true
}

// ResetSkillTree resets all skill points and refunds them.
func (s *ProgressionSystem) ResetSkillTree(
	world *ecs.World,
	entity ecs.Entity,
	tree *components.SkillTree,
) int {
	var level *components.Level

	query := s.levelFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			level = query.Get()

			break
		}
	}

	if level == nil {
		return 0
	}

	refunded := 0
	for _, node := range tree.Nodes {
		refunded += node.CurrentRank * node.Cost
		node.CurrentRank = 0
		node.Unlocked = false
	}

	level.SkillPoints += refunded

	return refunded
}
