package ai

import (
	"testing"
)

// MockGameAdapter implements GameAdapter for testing.
type MockGameAdapter struct {
	name      string
	tick      int64
	score     int
	playerPos [2]float64
	health    [2]int
	entities  int
	gameOver  bool
}

func NewMockGameAdapter() *MockGameAdapter {
	return &MockGameAdapter{
		name:      "MockGame",
		health:    [2]int{100, 100},
		playerPos: [2]float64{100, 100},
	}
}

func (m *MockGameAdapter) Name() string     { return m.name }
func (m *MockGameAdapter) GetScore() int    { return m.score }
func (m *MockGameAdapter) IsGameOver() bool { return m.gameOver }

func (m *MockGameAdapter) GetState() GameState {
	return GameState{
		Tick:         m.tick,
		Score:        m.score,
		PlayerPos:    m.playerPos,
		PlayerHealth: m.health,
		EntityCount:  m.entities,
	}
}

func (m *MockGameAdapter) AvailableActions() []ActionType {
	return []ActionType{ActionMoveUp, ActionMoveDown, ActionMoveLeft, ActionMoveRight}
}

func (m *MockGameAdapter) PerformAction(action ActionType) error {
	switch action {
	case ActionMoveUp:
		m.playerPos[1] -= 5
	case ActionMoveDown:
		m.playerPos[1] += 5
	case ActionMoveLeft:
		m.playerPos[0] -= 5
	case ActionMoveRight:
		m.playerPos[0] += 5
	}

	return nil
}

func (m *MockGameAdapter) Step() error {
	m.tick++
	m.score += 10

	return nil
}

func (m *MockGameAdapter) Reset() error {
	m.tick = 0
	m.score = 0
	m.playerPos = [2]float64{100, 100}
	m.health = [2]int{100, 100}
	m.gameOver = false

	return nil
}

// TestObserver tests the Observer component.
func TestObserver(t *testing.T) {
	t.Run("NewObserver creates observer with max history", func(t *testing.T) {
		obs := NewObserver(100)
		if obs.maxHistory != 100 {
			t.Errorf("maxHistory = %d, want 100", obs.maxHistory)
		}
	})

	t.Run("Record adds observations", func(t *testing.T) {
		obs := NewObserver(100)
		state := GameState{Tick: 1, Score: 50}
		obs.Record(1, state, ActionMoveUp)

		if len(obs.History()) != 1 {
			t.Errorf("History length = %d, want 1", len(obs.History()))
		}
	})

	t.Run("History respects maxHistory", func(t *testing.T) {
		obs := NewObserver(5)
		state := GameState{}

		for i := range 10 {
			state.Tick = int64(i)
			obs.Record(int64(i), state, ActionNone)
		}

		if len(obs.History()) != 5 {
			t.Errorf("History should be capped at 5, got %d", len(obs.History()))
		}
	})

	t.Run("LastN returns recent observations", func(t *testing.T) {
		obs := NewObserver(100)
		state := GameState{}

		for i := range 20 {
			state.Tick = int64(i)
			obs.Record(int64(i), state, ActionNone)
		}

		last5 := obs.LastN(5)
		if len(last5) != 5 {
			t.Errorf("LastN(5) returned %d, want 5", len(last5))
		}

		if last5[0].Tick != 15 {
			t.Errorf("LastN(5)[0].Tick = %d, want 15", last5[0].Tick)
		}
	})
}

// TestAnomalyDetector tests anomaly detection.
func TestAnomalyDetector(t *testing.T) {
	t.Run("NewAnomalyDetector creates with defaults", func(t *testing.T) {
		d := NewAnomalyDetector()
		if d.StuckThreshold != 120 {
			t.Errorf("StuckThreshold = %d, want 120", d.StuckThreshold)
		}
	})

	t.Run("DetectStuck finds stuck player", func(t *testing.T) {
		d := NewAnomalyDetector()
		d.StuckThreshold = 5 // Lower threshold for test

		history := make([]Observation, 10)
		for i := range history {
			history[i] = Observation{
				Tick: int64(i),
				State: GameState{
					PlayerPos: [2]float64{100, 100}, // Same position
				},
			}
		}

		anomalies := d.Analyze(history)
		found := false

		for _, a := range anomalies {
			if a.Type == AnomalyStuck {
				found = true

				break
			}
		}

		if !found {
			t.Error("Should detect stuck player")
		}
	})

	t.Run("DetectEntityLeak finds too many entities", func(t *testing.T) {
		d := NewAnomalyDetector()
		d.EntityLeakThreshold = 50

		// Need 10+ entries for entity leak detection
		history := make([]Observation, 15)
		for i := range history {
			history[i] = Observation{
				Tick:  int64(i),
				State: GameState{EntityCount: 100}, // Over threshold
			}
		}

		anomalies := d.Analyze(history)
		found := false

		for _, a := range anomalies {
			if a.Type == AnomalyEntityLeak {
				found = true

				break
			}
		}

		if !found {
			t.Error("Should detect entity leak")
		}
	})

	t.Run("DetectScoreRegression finds decreasing score", func(t *testing.T) {
		d := NewAnomalyDetector()

		history := []Observation{
			{Tick: 1, State: GameState{Score: 100}},
			{Tick: 2, State: GameState{Score: 50}}, // Score decreased
		}

		anomalies := d.Analyze(history)
		found := false

		for _, a := range anomalies {
			if a.Type == AnomalyScoreRegression {
				found = true

				break
			}
		}

		if !found {
			t.Error("Should detect score regression")
		}
	})
}

// TestPlayers tests player strategies.
func TestPlayers(t *testing.T) {
	t.Run("RandomPlayer picks from available", func(t *testing.T) {
		p := NewRandomPlayer(42)
		state := GameState{}
		available := []ActionType{ActionMoveUp, ActionMoveDown}

		action := p.DecideAction(state, available)
		if action != ActionMoveUp && action != ActionMoveDown {
			t.Errorf("RandomPlayer picked invalid action: %s", action)
		}
	})

	t.Run("RandomPlayer returns None for empty", func(t *testing.T) {
		p := NewRandomPlayer(42)
		state := GameState{}

		action := p.DecideAction(state, []ActionType{})
		if action != ActionNone {
			t.Errorf("Should return ActionNone for empty available, got %s", action)
		}
	})

	t.Run("ReplayPlayer replays sequence", func(t *testing.T) {
		actions := []ActionType{ActionMoveUp, ActionMoveDown, ActionMoveLeft}
		p := NewReplayPlayer(actions)
		state := GameState{}
		available := []ActionType{ActionMoveUp, ActionMoveDown, ActionMoveLeft, ActionMoveRight}

		for i, expected := range actions {
			got := p.DecideAction(state, available)
			if got != expected {
				t.Errorf("Step %d: got %s, want %s", i, got, expected)
			}
		}
	})
}

// TestQASession tests the session orchestrator.
func TestQASession(t *testing.T) {
	t.Run("NewQASession creates session", func(t *testing.T) {
		adapter := NewMockGameAdapter()
		session := NewQASession(adapter)

		if session.adapter == nil {
			t.Error("adapter should not be nil")
		}

		if session.observer == nil {
			t.Error("observer should not be nil")
		}
	})

	t.Run("Run executes game runs", func(t *testing.T) {
		adapter := NewMockGameAdapter()
		session := NewQASession(adapter)
		session.SetConfig(SessionConfig{
			Runs:     2,
			MaxTicks: 100,
		})
		session.SetPlayer(NewRandomPlayer(42))

		report := session.Run()

		if len(report.Runs) != 2 {
			t.Errorf("Should have 2 runs, got %d", len(report.Runs))
		}

		if report.GameName != "MockGame" {
			t.Errorf("GameName = %s, want MockGame", report.GameName)
		}
	})

	t.Run("Report generates markdown", func(t *testing.T) {
		adapter := NewMockGameAdapter()
		session := NewQASession(adapter)
		session.SetConfig(SessionConfig{
			Runs:     1,
			MaxTicks: 10,
		})

		report := session.Run()
		md := report.GenerateMarkdown()

		if len(md) == 0 {
			t.Error("Markdown report should not be empty")
		}

		if !contains2(md, "QA Test Report") {
			t.Error("Markdown should contain title")
		}
	})
}

func contains2(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && (s[:len(substr)] == substr || contains2(s[1:], substr)))
}
