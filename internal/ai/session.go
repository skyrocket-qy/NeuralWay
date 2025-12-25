package ai

import (
	"fmt"
	"strings"
	"time"
)

// SessionConfig configures a QA testing session.
type SessionConfig struct {
	Runs          int  // Number of game runs to perform
	MaxTicks      int  // Maximum ticks per run
	RecordEvery   int  // Record observation every N ticks (0 = every tick)
	StopOnAnomaly bool // Stop run when anomaly detected
}

// DefaultSessionConfig returns sensible defaults.
func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		Runs:          5,
		MaxTicks:      3600, // 1 minute at 60fps
		RecordEvery:   1,
		StopOnAnomaly: false,
	}
}

// RunResult contains the result of a single game run.
type RunResult struct {
	RunIndex   int           `json:"run_index"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	TotalTicks int64         `json:"total_ticks"`
	FinalScore int           `json:"final_score"`
	GameOver   bool          `json:"game_over"`
	Anomalies  []Anomaly     `json:"anomalies"`
	Stats      ObserverStats `json:"stats"`
}

// QAReport is the final report from a QA session.
type QAReport struct {
	GameName    string        `json:"game_name"`
	SessionTime time.Time     `json:"session_time"`
	Config      SessionConfig `json:"config"`
	Runs        []RunResult   `json:"runs"`

	// Aggregated
	TotalAnomalies int    `json:"total_anomalies"`
	BestScore      int    `json:"best_score"`
	WorstScore     int    `json:"worst_score"`
	AvgScore       int    `json:"avg_score"`
	Conclusion     string `json:"conclusion"`
}

// QASession orchestrates automated QA testing.
type QASession struct {
	adapter  GameAdapter
	player   Player
	observer *Observer
	detector *AnomalyDetector

	config SessionConfig
	runs   []RunResult
}

// NewQASession creates a QA session for the given game.
func NewQASession(adapter GameAdapter) *QASession {
	return &QASession{
		adapter:  adapter,
		observer: NewObserver(10000),
		detector: NewAnomalyDetector(),
		config:   DefaultSessionConfig(),
	}
}

// SetPlayer sets the AI player strategy.
func (s *QASession) SetPlayer(player Player) {
	s.player = player
}

// SetConfig sets the session configuration.
func (s *QASession) SetConfig(config SessionConfig) {
	s.config = config
}

// SetDetector replaces the anomaly detector.
func (s *QASession) SetDetector(detector *AnomalyDetector) {
	s.detector = detector
}

// Run executes the full QA session.
func (s *QASession) Run() QAReport {
	s.runs = make([]RunResult, 0, s.config.Runs)

	for i := 0; i < s.config.Runs; i++ {
		result := s.runSingle(i)
		s.runs = append(s.runs, result)
	}

	return s.generateReport()
}

// runSingle executes a single game run.
func (s *QASession) runSingle(index int) RunResult {
	result := RunResult{
		RunIndex:  index,
		StartTime: time.Now(),
	}

	// Reset game and observer
	s.adapter.Reset()
	s.observer.Clear()

	// Get default player if none set
	player := s.player
	if player == nil {
		player = NewRandomPlayer(time.Now().UnixNano())
	}

	// Run game loop
	var tick int64
	for tick = 0; tick < int64(s.config.MaxTicks); tick++ {
		// Get state
		state := s.adapter.GetState()

		// Decide and perform action
		available := s.adapter.AvailableActions()
		action := player.DecideAction(state, available)
		s.adapter.PerformAction(action)

		// Step game
		s.adapter.Step()

		// Record observation
		if s.config.RecordEvery <= 1 || tick%int64(s.config.RecordEvery) == 0 {
			s.observer.Record(tick, state, action)
		}

		// Check game over
		if s.adapter.IsGameOver() {
			result.GameOver = true

			break
		}

		// Detect anomalies if configured to stop
		if s.config.StopOnAnomaly && tick%100 == 0 {
			anomalies := s.detector.Analyze(s.observer.History())
			if len(anomalies) > 0 {
				result.Anomalies = anomalies

				break
			}
		}
	}

	// Final analysis
	result.EndTime = time.Now()
	result.TotalTicks = tick
	result.FinalScore = s.adapter.GetScore()
	result.Stats = s.observer.Stats()

	// Full anomaly detection
	if len(result.Anomalies) == 0 {
		result.Anomalies = s.detector.Analyze(s.observer.History())
	}

	return result
}

// generateReport creates the final QA report.
func (s *QASession) generateReport() QAReport {
	report := QAReport{
		GameName:    s.adapter.Name(),
		SessionTime: time.Now(),
		Config:      s.config,
		Runs:        s.runs,
	}

	// Aggregate stats
	totalScore := 0

	for _, run := range s.runs {
		report.TotalAnomalies += len(run.Anomalies)
		totalScore += run.FinalScore

		if run.FinalScore > report.BestScore {
			report.BestScore = run.FinalScore
		}

		if report.WorstScore == 0 || run.FinalScore < report.WorstScore {
			report.WorstScore = run.FinalScore
		}
	}

	if len(s.runs) > 0 {
		report.AvgScore = totalScore / len(s.runs)
	}

	// Determine conclusion
	if report.TotalAnomalies == 0 {
		report.Conclusion = "PASS - No anomalies detected"
	} else if report.TotalAnomalies <= 3 {
		report.Conclusion = "WARNING - Minor issues found"
	} else {
		report.Conclusion = "FAIL - Multiple anomalies detected"
	}

	return report
}

// GenerateMarkdown creates a markdown report.
func (r *QAReport) GenerateMarkdown() string {
	var sb strings.Builder

	sb.WriteString("# QA Test Report\n\n")
	sb.WriteString(fmt.Sprintf("**Game**: %s\n", r.GameName))
	sb.WriteString(fmt.Sprintf("**Date**: %s\n", r.SessionTime.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Runs**: %d × %d ticks\n\n", r.Config.Runs, r.Config.MaxTicks))

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total Anomalies | %d |\n", r.TotalAnomalies))
	sb.WriteString(fmt.Sprintf("| Best Score | %d |\n", r.BestScore))
	sb.WriteString(fmt.Sprintf("| Avg Score | %d |\n", r.AvgScore))
	sb.WriteString(fmt.Sprintf("| Worst Score | %d |\n", r.WorstScore))
	sb.WriteString(fmt.Sprintf("| **Conclusion** | **%s** |\n\n", r.Conclusion))

	// Per-run details
	sb.WriteString("## Run Details\n\n")

	for _, run := range r.Runs {
		sb.WriteString(fmt.Sprintf("### Run %d\n", run.RunIndex+1))
		sb.WriteString(fmt.Sprintf("- Ticks: %d | Score: %d | Game Over: %v\n",
			run.TotalTicks, run.FinalScore, run.GameOver))

		if len(run.Anomalies) > 0 {
			sb.WriteString(fmt.Sprintf("- Anomalies: %d\n", len(run.Anomalies)))

			for _, a := range run.Anomalies {
				sb.WriteString(fmt.Sprintf("  - [%s] %s @ tick %d\n",
					a.Severity.String(), a.Type, a.Tick))
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
