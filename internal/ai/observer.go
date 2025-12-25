package ai

import (
	"time"
)

// Observation represents a single frame of game state during QA testing.
type Observation struct {
	Tick      int64      `json:"tick"`
	Timestamp time.Time  `json:"timestamp"`
	State     GameState  `json:"state"`
	Action    ActionType `json:"action"`
	Metrics   QAMetrics  `json:"metrics"`
}

// QAMetrics contains performance and health metrics.
type QAMetrics struct {
	EntityCount  int     `json:"entity_count"`
	FPS          float64 `json:"fps,omitempty"`
	MemoryMB     float64 `json:"memory_mb,omitempty"`
	UpdateTimeMs float64 `json:"update_time_ms,omitempty"`
}

// Observer records game state history for anomaly detection.
type Observer struct {
	history    []Observation
	maxHistory int

	// Callbacks
	OnObservation  func(obs Observation)
	OnAnomalyFound func(anomaly Anomaly)

	// Internal tracking
	lastState    GameState
	stuckCounter int
}

// NewObserver creates an observer with the given history limit.
func NewObserver(maxHistory int) *Observer {
	if maxHistory <= 0 {
		maxHistory = 1000
	}

	return &Observer{
		history:    make([]Observation, 0, maxHistory),
		maxHistory: maxHistory,
	}
}

// Record adds an observation to history.
func (o *Observer) Record(tick int64, state GameState, action ActionType) Observation {
	obs := Observation{
		Tick:      tick,
		Timestamp: time.Now(),
		State:     state,
		Action:    action,
		Metrics: QAMetrics{
			EntityCount: state.EntityCount,
		},
	}

	// Append and trim if needed
	o.history = append(o.history, obs)
	if len(o.history) > o.maxHistory {
		o.history = o.history[1:]
	}

	// Callback
	if o.OnObservation != nil {
		o.OnObservation(obs)
	}

	// Update tracking
	o.lastState = state

	return obs
}

// History returns all recorded observations.
func (o *Observer) History() []Observation {
	return o.history
}

// LastN returns the most recent N observations.
func (o *Observer) LastN(n int) []Observation {
	if n >= len(o.history) {
		return o.history
	}

	return o.history[len(o.history)-n:]
}

// Clear resets the observation history.
func (o *Observer) Clear() {
	o.history = make([]Observation, 0, o.maxHistory)
	o.stuckCounter = 0
}

// GetStateChanges returns observations where state changed significantly.
func (o *Observer) GetStateChanges() []Observation {
	changes := make([]Observation, 0)

	for i := 1; i < len(o.history); i++ {
		prev := o.history[i-1]
		curr := o.history[i]

		// Detect significant changes
		if curr.State.Score != prev.State.Score ||
			curr.State.PlayerHealth[0] != prev.State.PlayerHealth[0] ||
			curr.State.EntityCount != prev.State.EntityCount {
			changes = append(changes, curr)
		}
	}

	return changes
}

// GetPositionHistory returns player positions over time.
func (o *Observer) GetPositionHistory() [][2]float64 {
	positions := make([][2]float64, len(o.history))
	for i, obs := range o.history {
		positions[i] = obs.State.PlayerPos
	}

	return positions
}

// GetScoreHistory returns score progression over time.
func (o *Observer) GetScoreHistory() []int {
	scores := make([]int, len(o.history))
	for i, obs := range o.history {
		scores[i] = obs.State.Score
	}

	return scores
}

// Stats returns summary statistics of observations.
func (o *Observer) Stats() ObserverStats {
	if len(o.history) == 0 {
		return ObserverStats{}
	}

	first := o.history[0]
	last := o.history[len(o.history)-1]

	maxScore := 0
	minHealth := first.State.PlayerHealth[0]
	maxEntities := 0

	for _, obs := range o.history {
		if obs.State.Score > maxScore {
			maxScore = obs.State.Score
		}

		if obs.State.PlayerHealth[0] < minHealth {
			minHealth = obs.State.PlayerHealth[0]
		}

		if obs.State.EntityCount > maxEntities {
			maxEntities = obs.State.EntityCount
		}
	}

	return ObserverStats{
		TotalTicks:   last.Tick - first.Tick + 1,
		Observations: len(o.history),
		MaxScore:     maxScore,
		MinHealth:    minHealth,
		MaxEntities:  maxEntities,
		FinalScore:   last.State.Score,
	}
}

// ObserverStats contains summary statistics.
type ObserverStats struct {
	TotalTicks   int64
	Observations int
	MaxScore     int
	MinHealth    int
	MaxEntities  int
	FinalScore   int
}
