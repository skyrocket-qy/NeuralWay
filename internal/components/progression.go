package components

// Experience tracks XP and level requirements.
type Experience struct {
	Current    int64   // Current XP
	Total      int64   // Total XP ever earned
	BaseToNext int64   // Base XP needed for next level
	Scaling    float64 // XP scaling per level (e.g., 1.5 = 50% more per level)
}

// NewExperience creates an experience tracker.
func NewExperience(baseToNext int64, scaling float64) Experience {
	return Experience{
		BaseToNext: baseToNext,
		Scaling:    scaling,
	}
}

// ToNextLevel returns XP needed for the next level from current level.
func (e *Experience) ToNextLevel(currentLevel int) int64 {
	if e.Scaling <= 1 {
		return e.BaseToNext
	}

	multiplier := 1.0
	for i := 1; i < currentLevel; i++ {
		multiplier *= e.Scaling
	}

	return int64(float64(e.BaseToNext) * multiplier)
}

// Level represents an entity's level and stat scaling.
type Level struct {
	Current      int
	Max          int
	StatPerLevel map[string]float64 // Stats gained per level
	SkillPoints  int                // Unspent skill points
	PointsPerLvl int                // Skill points gained per level
}

// NewLevel creates a level component.
func NewLevel(maxVal int) Level {
	return Level{
		Current:      1,
		Max:          maxVal,
		StatPerLevel: make(map[string]float64),
		PointsPerLvl: 1,
	}
}

// CanLevelUp returns true if entity can gain a level.
func (l *Level) CanLevelUp() bool {
	return l.Max == 0 || l.Current < l.Max
}

// LevelUp increases the level and grants skill points.
func (l *Level) LevelUp() bool {
	if !l.CanLevelUp() {
		return false
	}

	l.Current++
	l.SkillPoints += l.PointsPerLvl

	return true
}

// SkillTreeNode represents a node in a skill tree.
type SkillTreeNode struct {
	ID           string
	Name         string
	Description  string
	Unlocked     bool
	MaxRank      int
	CurrentRank  int
	Cost         int                // Skill points to unlock/upgrade
	Requirements []string           // IDs of prerequisite nodes
	Effects      map[string]float64 // Stat modifications when unlocked
}

// NewSkillTreeNode creates a skill tree node.
func NewSkillTreeNode(id, name string, maxRank, cost int) SkillTreeNode {
	return SkillTreeNode{
		ID:           id,
		Name:         name,
		MaxRank:      maxRank,
		CurrentRank:  0,
		Cost:         cost,
		Requirements: make([]string, 0),
		Effects:      make(map[string]float64),
	}
}

// CanUnlock checks if the node can be unlocked given unlocked nodes.
func (s *SkillTreeNode) CanUnlock(unlockedNodes map[string]bool, skillPoints int) bool {
	if s.CurrentRank >= s.MaxRank {
		return false
	}

	if skillPoints < s.Cost {
		return false
	}

	for _, req := range s.Requirements {
		if !unlockedNodes[req] {
			return false
		}
	}

	return true
}

// SkillTree contains all skill tree nodes for an entity.
type SkillTree struct {
	Nodes map[string]*SkillTreeNode
}

// NewSkillTree creates an empty skill tree.
func NewSkillTree() SkillTree {
	return SkillTree{
		Nodes: make(map[string]*SkillTreeNode),
	}
}

// AddNode adds a node to the skill tree.
func (st *SkillTree) AddNode(node SkillTreeNode) {
	st.Nodes[node.ID] = &node
}

// GetUnlocked returns a map of unlocked node IDs.
func (st *SkillTree) GetUnlocked() map[string]bool {
	unlocked := make(map[string]bool)

	for id, node := range st.Nodes {
		if node.Unlocked || node.CurrentRank > 0 {
			unlocked[id] = true
		}
	}

	return unlocked
}

// GetTotalEffects returns cumulative effects from all unlocked nodes.
func (st *SkillTree) GetTotalEffects() map[string]float64 {
	effects := make(map[string]float64)

	for _, node := range st.Nodes {
		if node.CurrentRank > 0 {
			for stat, value := range node.Effects {
				effects[stat] += value * float64(node.CurrentRank)
			}
		}
	}

	return effects
}

// Achievement represents a trackable achievement.
type Achievement struct {
	ID          string
	Name        string
	Description string
	Unlocked    bool
	UnlockedAt  int64  // Unix timestamp when unlocked
	Progress    int    // Current progress
	Target      int    // Target to unlock
	Hidden      bool   // Hidden until unlocked
	Reward      string // Optional reward ID
}

// NewAchievement creates an achievement.
func NewAchievement(id, name, description string, target int) Achievement {
	return Achievement{
		ID:          id,
		Name:        name,
		Description: description,
		Target:      target,
	}
}

// AddProgress adds progress and returns true if just unlocked.
func (a *Achievement) AddProgress(amount int) bool {
	if a.Unlocked {
		return false
	}

	a.Progress += amount
	if a.Progress >= a.Target {
		a.Unlocked = true

		return true
	}

	return false
}

// AchievementTracker holds all achievements for an entity.
type AchievementTracker struct {
	Achievements map[string]*Achievement
}

// NewAchievementTracker creates an achievement tracker.
func NewAchievementTracker() AchievementTracker {
	return AchievementTracker{
		Achievements: make(map[string]*Achievement),
	}
}

// Add registers an achievement.
func (at *AchievementTracker) Add(achievement Achievement) {
	at.Achievements[achievement.ID] = &achievement
}

// Progress adds progress to an achievement by ID.
func (at *AchievementTracker) Progress(id string, amount int) bool {
	if a, ok := at.Achievements[id]; ok {
		return a.AddProgress(amount)
	}

	return false
}

// GetUnlockedCount returns the number of unlocked achievements.
func (at *AchievementTracker) GetUnlockedCount() int {
	count := 0

	for _, a := range at.Achievements {
		if a.Unlocked {
			count++
		}
	}

	return count
}

// Prestige represents rebirth/prestige progression.
type Prestige struct {
	Level          int                // Current prestige level
	Points         int64              // Prestige currency
	Multipliers    map[string]float64 // Permanent multipliers
	UnlockedPerks  []string           // Unlocked prestige perks
	ResetThreshold int64              // Required progress to prestige
}

// NewPrestige creates a prestige tracker.
func NewPrestige(resetThreshold int64) Prestige {
	return Prestige{
		Multipliers:    make(map[string]float64),
		UnlockedPerks:  make([]string, 0),
		ResetThreshold: resetThreshold,
	}
}

// CanPrestige returns true if prestige is available.
func (p *Prestige) CanPrestige(currentProgress int64) bool {
	return currentProgress >= p.ResetThreshold
}

// DoPrestige performs a prestige reset and grants points.
func (p *Prestige) DoPrestige(currentProgress int64) int64 {
	if !p.CanPrestige(currentProgress) {
		return 0
	}
	// Grant prestige points based on progress
	gained := currentProgress / p.ResetThreshold
	p.Points += gained
	p.Level++

	return gained
}

// Unlockable represents content that can be unlocked.
type Unlockable struct {
	ID          string
	Name        string
	Unlocked    bool
	Requirement string // Description of how to unlock
	Category    string // Type of unlockable (character, level, item, etc.)
}

// NewUnlockable creates an unlockable.
func NewUnlockable(id, name, requirement, category string) Unlockable {
	return Unlockable{
		ID:          id,
		Name:        name,
		Requirement: requirement,
		Category:    category,
	}
}
