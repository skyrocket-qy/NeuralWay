package ai

import (
	"math/rand"
	"slices"
)

// RandomPlayer selects actions randomly for fuzz testing.
type RandomPlayer struct {
	rng *rand.Rand
}

// NewRandomPlayer creates a player that picks random actions.
func NewRandomPlayer(seed int64) *RandomPlayer {
	return &RandomPlayer{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// DecideAction picks a random action from available actions.
func (p *RandomPlayer) DecideAction(state GameState, available []ActionType) ActionType {
	if len(available) == 0 {
		return ActionNone
	}

	return available[p.rng.Intn(len(available))]
}

// WeightedRandomPlayer selects actions with weighted probabilities.
type WeightedRandomPlayer struct {
	rng     *rand.Rand
	weights map[ActionType]float64
}

// NewWeightedRandomPlayer creates a player with action weights.
func NewWeightedRandomPlayer(seed int64, weights map[ActionType]float64) *WeightedRandomPlayer {
	return &WeightedRandomPlayer{
		rng:     rand.New(rand.NewSource(seed)),
		weights: weights,
	}
}

// DecideAction picks an action based on weights.
func (p *WeightedRandomPlayer) DecideAction(state GameState, available []ActionType) ActionType {
	if len(available) == 0 {
		return ActionNone
	}

	// Calculate total weight for available actions
	totalWeight := 0.0

	for _, a := range available {
		if w, ok := p.weights[a]; ok {
			totalWeight += w
		} else {
			totalWeight += 1.0 // Default weight
		}
	}

	// Pick based on weight
	pick := p.rng.Float64() * totalWeight
	cumulative := 0.0

	for _, a := range available {
		w := 1.0
		if pw, ok := p.weights[a]; ok {
			w = pw
		}

		cumulative += w
		if pick <= cumulative {
			return a
		}
	}

	return available[len(available)-1]
}

// StrategyPlayer uses a behavior tree for decision making.
type StrategyPlayer struct {
	tree       *BehaviorTree
	actionMap  map[string]ActionType
	lastAction ActionType
}

// NewStrategyPlayer creates a player with a behavior tree strategy.
func NewStrategyPlayer(tree *BehaviorTree) *StrategyPlayer {
	return &StrategyPlayer{
		tree: tree,
		actionMap: map[string]ActionType{
			"move_up":    ActionMoveUp,
			"move_down":  ActionMoveDown,
			"move_left":  ActionMoveLeft,
			"move_right": ActionMoveRight,
			"attack":     ActionAttack,
			"jump":       ActionJump,
		},
	}
}

// DecideAction runs the behavior tree to pick an action.
func (p *StrategyPlayer) DecideAction(state GameState, available []ActionType) ActionType {
	// For now, simplified - in full implementation would run BT
	// and extract action from blackboard

	// Simple chase behavior: move toward center if far
	centerX := 450.0
	centerY := 350.0

	dx := centerX - state.PlayerPos[0]
	dy := centerY - state.PlayerPos[1]

	if abs(dx) > abs(dy) {
		if dx > 0 && contains(available, ActionMoveRight) {
			return ActionMoveRight
		} else if dx < 0 && contains(available, ActionMoveLeft) {
			return ActionMoveLeft
		}
	} else {
		if dy > 0 && contains(available, ActionMoveDown) {
			return ActionMoveDown
		} else if dy < 0 && contains(available, ActionMoveUp) {
			return ActionMoveUp
		}
	}

	if len(available) > 0 {
		return available[0]
	}

	return ActionNone
}

// ReplayPlayer replays a recorded action sequence.
type ReplayPlayer struct {
	actions []ActionType
	index   int
}

// NewReplayPlayer creates a player that replays actions.
func NewReplayPlayer(actions []ActionType) *ReplayPlayer {
	return &ReplayPlayer{
		actions: actions,
		index:   0,
	}
}

// DecideAction returns the next action in the sequence.
func (p *ReplayPlayer) DecideAction(state GameState, available []ActionType) ActionType {
	if p.index >= len(p.actions) {
		return ActionNone
	}

	action := p.actions[p.index]
	p.index++

	return action
}

// Reset restarts the replay from beginning.
func (p *ReplayPlayer) Reset() {
	p.index = 0
}

// Helper functions.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}

	return x
}

func contains(slice []ActionType, item ActionType) bool {
	return slices.Contains(slice, item)
}
