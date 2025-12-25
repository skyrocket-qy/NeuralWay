package components

// GamePhase represents the current game phase.
type GamePhase int

const (
	PhaseTitle GamePhase = iota
	PhaseLoading
	PhasePlaying
	PhasePaused
	PhaseGameOver
	PhaseVictory
	PhaseTransition
)

// String returns the phase name.
func (p GamePhase) String() string {
	names := []string{"Title", "Loading", "Playing", "Paused", "GameOver", "Victory", "Transition"}
	if int(p) < len(names) {
		return names[p]
	}

	return "Unknown"
}

// GameState represents the overall game state.
type GameState struct {
	Phase          GamePhase
	PreviousPhase  GamePhase
	ElapsedTime    float64 // Total time in current phase
	PlayTime       float64 // Total play time (excludes pause)
	TransitionTo   GamePhase
	TransitionTime float64
}

// NewGameState creates an initial game state.
func NewGameState() GameState {
	return GameState{
		Phase: PhaseTitle,
	}
}

// SetPhase changes the game phase.
func (gs *GameState) SetPhase(phase GamePhase) {
	gs.PreviousPhase = gs.Phase
	gs.Phase = phase
	gs.ElapsedTime = 0
}

// IsPaused returns true if the game is paused.
func (gs *GameState) IsPaused() bool {
	return gs.Phase == PhasePaused
}

// IsPlaying returns true if actively playing.
func (gs *GameState) IsPlaying() bool {
	return gs.Phase == PhasePlaying
}

// Lives represents player lives/attempts.
type Lives struct {
	Current      int
	Max          int
	Infinite     bool // Ignore current/max if true
	ExtraEnabled bool // Whether extra lives can be gained
}

// NewLives creates a lives component.
func NewLives(maxVal int) Lives {
	return Lives{
		Current:      maxVal,
		Max:          maxVal,
		ExtraEnabled: true,
	}
}

// Lose decrements a life, returns true if still alive.
func (l *Lives) Lose() bool {
	if l.Infinite {
		return true
	}

	l.Current--

	return l.Current > 0
}

// Add adds lives up to max (or beyond if extraEnabled).
func (l *Lives) Add(amount int) {
	l.Current += amount
	if !l.ExtraEnabled && l.Current > l.Max {
		l.Current = l.Max
	}
}

// IsGameOver returns true if no lives remain.
func (l *Lives) IsGameOver() bool {
	return !l.Infinite && l.Current <= 0
}

// Checkpoint represents a save point.
type Checkpoint struct {
	ID        string
	X, Y      float64 // Position to respawn
	Activated bool
	Timestamp float64 // When activated
	LevelID   string  // Level this checkpoint belongs to
}

// NewCheckpoint creates a checkpoint.
func NewCheckpoint(id string, x, y float64) Checkpoint {
	return Checkpoint{
		ID: id,
		X:  x,
		Y:  y,
	}
}

// Activate marks the checkpoint as active.
func (c *Checkpoint) Activate(timestamp float64) {
	c.Activated = true
	c.Timestamp = timestamp
}

// CheckpointManager tracks all checkpoints.
type CheckpointManager struct {
	Checkpoints map[string]*Checkpoint
	ActiveID    string
}

// NewCheckpointManager creates a checkpoint manager.
func NewCheckpointManager() CheckpointManager {
	return CheckpointManager{
		Checkpoints: make(map[string]*Checkpoint),
	}
}

// Add registers a checkpoint.
func (cm *CheckpointManager) Add(checkpoint Checkpoint) {
	cm.Checkpoints[checkpoint.ID] = &checkpoint
}

// SetActive sets the current active checkpoint.
func (cm *CheckpointManager) SetActive(id string, timestamp float64) {
	if cp, ok := cm.Checkpoints[id]; ok {
		cp.Activate(timestamp)

		cm.ActiveID = id
	}
}

// GetSpawnPoint returns the position of the active checkpoint.
func (cm *CheckpointManager) GetSpawnPoint() (float64, float64, bool) {
	if cp, ok := cm.Checkpoints[cm.ActiveID]; ok && cp.Activated {
		return cp.X, cp.Y, true
	}

	return 0, 0, false
}

// WinConditionType represents how victory is determined.
type WinConditionType string

const (
	WinKillAllEnemies  WinConditionType = "kill_all"
	WinReachGoal       WinConditionType = "reach_goal"
	WinSurviveTime     WinConditionType = "survive"
	WinCollectItems    WinConditionType = "collect"
	WinDefendObjective WinConditionType = "defend"
	WinScore           WinConditionType = "score"
	WinBossDefeated    WinConditionType = "boss"
)

// WinCondition defines victory requirements.
type WinCondition struct {
	Type          WinConditionType
	Target        int     // Target count/score
	Current       int     // Current progress
	TimeLimit     float64 // Time limit (0 = no limit)
	TimeRemaining float64
	Completed     bool
}

// NewWinCondition creates a win condition.
func NewWinCondition(condType WinConditionType, target int) WinCondition {
	return WinCondition{
		Type:   condType,
		Target: target,
	}
}

// AddProgress adds progress toward the win condition.
func (wc *WinCondition) AddProgress(amount int) {
	wc.Current += amount
	if wc.Current >= wc.Target {
		wc.Completed = true
	}
}

// IsComplete returns true if the condition is met.
func (wc *WinCondition) IsComplete() bool {
	if wc.Completed {
		return true
	}

	switch wc.Type {
	case WinSurviveTime:
		return wc.TimeRemaining <= 0 && wc.TimeLimit > 0
	default:
		return wc.Current >= wc.Target
	}
}

// GetProgress returns current/target as a ratio.
func (wc *WinCondition) GetProgress() float64 {
	if wc.Target == 0 {
		return 0
	}

	return float64(wc.Current) / float64(wc.Target)
}

// BossPhase represents a phase of a boss fight.
type BossPhase struct {
	ID              string
	Name            string
	HealthThreshold float64  // Transition when boss health % below this
	Duration        float64  // Time limit for phase (0 = until threshold)
	Attacks         []string // Available attack patterns
	SpeedModifier   float64  // Movement speed modifier
	DamageModifier  float64  // Damage modifier
	Invulnerable    bool     // Boss can't take damage
	Active          bool
}

// NewBossPhase creates a boss phase.
func NewBossPhase(id, name string, healthThreshold float64) BossPhase {
	return BossPhase{
		ID:              id,
		Name:            name,
		HealthThreshold: healthThreshold,
		Attacks:         make([]string, 0),
		SpeedModifier:   1.0,
		DamageModifier:  1.0,
	}
}

// BossController manages boss phase transitions.
type BossController struct {
	Phases       []BossPhase
	CurrentPhase int
	MaxHealth    float64
	Enraged      bool
	EnrageTimer  float64
}

// NewBossController creates a boss controller.
func NewBossController(maxHealth float64) BossController {
	return BossController{
		Phases:    make([]BossPhase, 0),
		MaxHealth: maxHealth,
	}
}

// AddPhase adds a phase to the boss.
func (bc *BossController) AddPhase(phase BossPhase) {
	bc.Phases = append(bc.Phases, phase)
}

// GetCurrentPhase returns the active phase.
func (bc *BossController) GetCurrentPhase() *BossPhase {
	if bc.CurrentPhase < len(bc.Phases) {
		return &bc.Phases[bc.CurrentPhase]
	}

	return nil
}

// CheckPhaseTransition checks if phase should change based on health.
func (bc *BossController) CheckPhaseTransition(currentHealth float64) bool {
	if bc.CurrentPhase >= len(bc.Phases)-1 {
		return false
	}

	healthPercent := currentHealth / bc.MaxHealth

	nextPhase := bc.Phases[bc.CurrentPhase+1]
	if healthPercent <= nextPhase.HealthThreshold {
		bc.CurrentPhase++
		bc.Phases[bc.CurrentPhase].Active = true

		return true
	}

	return false
}

// Timer represents a countdown or stopwatch timer.
type Timer struct {
	Duration  float64
	Remaining float64
	Elapsed   float64
	Paused    bool
	Looping   bool
	Completed bool
}

// NewTimer creates a countdown timer.
func NewTimer(duration float64) Timer {
	return Timer{
		Duration:  duration,
		Remaining: duration,
	}
}

// Update advances the timer.
func (t *Timer) Update(dt float64) {
	if t.Paused || t.Completed {
		return
	}

	t.Elapsed += dt

	t.Remaining -= dt
	if t.Remaining <= 0 {
		if t.Looping {
			t.Remaining = t.Duration
		} else {
			t.Remaining = 0
			t.Completed = true
		}
	}
}

// Reset restarts the timer.
func (t *Timer) Reset() {
	t.Remaining = t.Duration
	t.Elapsed = 0
	t.Completed = false
}

// GetProgress returns elapsed/duration as a ratio.
func (t *Timer) GetProgress() float64 {
	if t.Duration == 0 {
		return 0
	}

	return t.Elapsed / t.Duration
}
