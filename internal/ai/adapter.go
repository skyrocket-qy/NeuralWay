package ai

// GameAdapter is a universal interface for any game to be QA-tested.
// Implement this interface to enable automated testing for your game.
type GameAdapter interface {
	// Name returns the game's identifier.
	Name() string

	// GetState returns the current game state as an observable snapshot.
	GetState() GameState

	// IsGameOver returns true if the current run has ended.
	IsGameOver() bool

	// GetScore returns the current score (or relevant metric).
	GetScore() int

	// AvailableActions returns the list of valid actions in current state.
	AvailableActions() []ActionType

	// PerformAction executes the given action.
	PerformAction(action ActionType) error

	// Step advances the game by one tick/frame.
	Step() error

	// Reset restarts the game to initial state.
	Reset() error
}

// GameState contains observable game state for AI analysis.
type GameState struct {
	Tick         int64          `json:"tick"`
	Score        int            `json:"score"`
	PlayerPos    [2]float64     `json:"player_pos"`
	PlayerHealth [2]int         `json:"player_health"` // [current, max]
	EntityCount  int            `json:"entity_count"`
	CustomData   map[string]any `json:"custom,omitempty"`
}

// ActionType represents a game action that can be performed.
type ActionType string

// Common action types (games can define more).
const (
	ActionNone      ActionType = "none"
	ActionMoveUp    ActionType = "move_up"
	ActionMoveDown  ActionType = "move_down"
	ActionMoveLeft  ActionType = "move_left"
	ActionMoveRight ActionType = "move_right"
	ActionJump      ActionType = "jump"
	ActionAttack    ActionType = "attack"
	ActionUse       ActionType = "use"
	ActionPause     ActionType = "pause"
)

// Player is an interface for AI players that decide actions.
type Player interface {
	// DecideAction chooses an action based on current state.
	DecideAction(state GameState, available []ActionType) ActionType
}
