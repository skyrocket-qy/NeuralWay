package systems

import (
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// GameStateSystem manages game phase transitions.
type GameStateSystem struct {
	state            *components.GameState
	livesFilter      *ecs.Filter1[components.Lives]
	winFilter        *ecs.Filter1[components.WinCondition]
	checkpointMgr    *components.CheckpointManager
	onPhaseChange    func(from, to components.GamePhase)
	onGameOver       func()
	onVictory        func()
	pauseRequested   bool
	unpauseRequested bool
}

// NewGameStateSystem creates a game state system.
func NewGameStateSystem(world *ecs.World) *GameStateSystem {
	state := components.NewGameState()

	return &GameStateSystem{
		state:       &state,
		livesFilter: ecs.NewFilter1[components.Lives](world),
		winFilter:   ecs.NewFilter1[components.WinCondition](world),
	}
}

// SetCheckpointManager sets the checkpoint manager.
func (s *GameStateSystem) SetCheckpointManager(mgr *components.CheckpointManager) {
	s.checkpointMgr = mgr
}

// SetOnPhaseChange sets the phase change callback.
func (s *GameStateSystem) SetOnPhaseChange(fn func(from, to components.GamePhase)) {
	s.onPhaseChange = fn
}

// SetOnGameOver sets the game over callback.
func (s *GameStateSystem) SetOnGameOver(fn func()) {
	s.onGameOver = fn
}

// SetOnVictory sets the victory callback.
func (s *GameStateSystem) SetOnVictory(fn func()) {
	s.onVictory = fn
}

// GetState returns the current game state.
func (s *GameStateSystem) GetState() *components.GameState {
	return s.state
}

// GetPhase returns the current phase.
func (s *GameStateSystem) GetPhase() components.GamePhase {
	return s.state.Phase
}

// SetPhase changes the game phase.
func (s *GameStateSystem) SetPhase(phase components.GamePhase) {
	if s.state.Phase == phase {
		return
	}

	oldPhase := s.state.Phase
	s.state.SetPhase(phase)

	if s.onPhaseChange != nil {
		s.onPhaseChange(oldPhase, phase)
	}
}

// StartGame transitions to playing state.
func (s *GameStateSystem) StartGame() {
	s.SetPhase(components.PhasePlaying)
}

// RequestPause requests a pause on next update.
func (s *GameStateSystem) RequestPause() {
	if s.state.Phase == components.PhasePlaying {
		s.pauseRequested = true
	}
}

// RequestUnpause requests an unpause on next update.
func (s *GameStateSystem) RequestUnpause() {
	if s.state.Phase == components.PhasePaused {
		s.unpauseRequested = true
	}
}

// TogglePause toggles between playing and paused.
func (s *GameStateSystem) TogglePause() {
	switch s.state.Phase {
	case components.PhasePlaying:
		s.RequestPause()
	case components.PhasePaused:
		s.RequestUnpause()
	}
}

// IsPaused returns true if the game is paused.
func (s *GameStateSystem) IsPaused() bool {
	return s.state.IsPaused()
}

// IsPlaying returns true if actively playing.
func (s *GameStateSystem) IsPlaying() bool {
	return s.state.IsPlaying()
}

// LoseLife decrements a life and checks for game over.
func (s *GameStateSystem) LoseLife(world *ecs.World, entity ecs.Entity) bool {
	query := s.livesFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			lives := query.Get()

			stillAlive := lives.Lose()
			if !stillAlive {
				s.TriggerGameOver()
			}

			return stillAlive
		}
	}

	return false
}

// GetLives returns current and max lives for an entity.
func (s *GameStateSystem) GetLives(world *ecs.World, entity ecs.Entity) (current, maxVal int) {
	query := s.livesFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			lives := query.Get()

			return lives.Current, lives.Max
		}
	}

	return 0, 0
}

// AddLife adds a life to an entity.
func (s *GameStateSystem) AddLife(world *ecs.World, entity ecs.Entity) {
	query := s.livesFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			lives := query.Get()
			lives.Add(1)

			return
		}
	}
}

// TriggerGameOver transitions to game over state.
func (s *GameStateSystem) TriggerGameOver() {
	s.SetPhase(components.PhaseGameOver)

	if s.onGameOver != nil {
		s.onGameOver()
	}
}

// TriggerVictory transitions to victory state.
func (s *GameStateSystem) TriggerVictory() {
	s.SetPhase(components.PhaseVictory)

	if s.onVictory != nil {
		s.onVictory()
	}
}

// CheckWinConditions checks all win conditions.
func (s *GameStateSystem) CheckWinConditions(world *ecs.World) bool {
	allComplete := true
	hasConditions := false

	query := s.winFilter.Query()
	for query.Next() {
		hasConditions = true

		wc := query.Get()
		if !wc.IsComplete() {
			allComplete = false
		}
	}

	if hasConditions && allComplete {
		s.TriggerVictory()

		return true
	}

	return false
}

// AddWinProgress adds progress to a win condition.
func (s *GameStateSystem) AddWinProgress(world *ecs.World, entity ecs.Entity, amount int) {
	query := s.winFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			wc := query.Get()
			wc.AddProgress(amount)

			return
		}
	}
}

// ActivateCheckpoint activates a checkpoint.
func (s *GameStateSystem) ActivateCheckpoint(checkpointID string) {
	if s.checkpointMgr != nil {
		s.checkpointMgr.SetActive(checkpointID, s.state.PlayTime)
	}
}

// GetSpawnPoint returns the current spawn point.
func (s *GameStateSystem) GetSpawnPoint() (x, y float64, ok bool) {
	if s.checkpointMgr != nil {
		return s.checkpointMgr.GetSpawnPoint()
	}

	return 0, 0, false
}

// Update updates game state.
func (s *GameStateSystem) Update(world *ecs.World, dt float64) {
	// Handle pause requests
	if s.pauseRequested {
		s.SetPhase(components.PhasePaused)
		s.pauseRequested = false
	}

	if s.unpauseRequested {
		s.SetPhase(components.PhasePlaying)
		s.unpauseRequested = false
	}

	// Update timers
	s.state.ElapsedTime += dt
	if s.state.Phase == components.PhasePlaying {
		s.state.PlayTime += dt
	}

	// Update win condition timers
	query := s.winFilter.Query()
	for query.Next() {
		wc := query.Get()
		if wc.TimeLimit > 0 {
			wc.TimeRemaining -= dt
			if wc.TimeRemaining <= 0 {
				wc.TimeRemaining = 0
				// Check if this is a survive condition
				if wc.Type == components.WinSurviveTime {
					wc.Completed = true
				}
			}
		}
	}
}

// GetPlayTime returns total play time.
func (s *GameStateSystem) GetPlayTime() float64 {
	return s.state.PlayTime
}

// Reset resets the game state to initial.
func (s *GameStateSystem) Reset() {
	*s.state = components.NewGameState()
}

// TransitionTo starts a transition to another phase.
func (s *GameStateSystem) TransitionTo(phase components.GamePhase, duration float64) {
	s.state.TransitionTo = phase
	s.state.TransitionTime = duration
	s.SetPhase(components.PhaseTransition)
}
