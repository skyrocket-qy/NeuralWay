package ai

import (
	"fmt"
	"math"
	"strings"
)

// AnomalyType categorizes detected issues.
type AnomalyType string

const (
	AnomalyStuck             AnomalyType = "stuck"              // Player not moving
	AnomalyDeathLoop         AnomalyType = "death_loop"         // Repeated deaths at same location
	AnomalyEntityLeak        AnomalyType = "entity_leak"        // Entity count growing unbounded
	AnomalyStateOscillation  AnomalyType = "state_oscillation"  // Rapid state changes
	AnomalyScoreRegression   AnomalyType = "score_regression"   // Score decreased
	AnomalyHealthDrain       AnomalyType = "health_drain"       // Rapid health loss
	AnomalyInfiniteLoop      AnomalyType = "infinite_loop"      // Same state repeating
	AnomalyBoundaryViolation AnomalyType = "boundary_violation" // Entity out of bounds
)

// Severity indicates how serious an anomaly is.
type Severity int

const (
	SeverityLow Severity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityLow:
		return "LOW"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityHigh:
		return "HIGH"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Anomaly represents a detected issue during QA testing.
type Anomaly struct {
	Type        AnomalyType   `json:"type"`
	Severity    Severity      `json:"severity"`
	Tick        int64         `json:"tick"`
	Description string        `json:"description"`
	Evidence    []Observation `json:"evidence,omitempty"`
}

// DetectionRule is a function that checks for anomalies.
type DetectionRule func(history []Observation) []Anomaly

// AnomalyDetector analyzes observation history to find bugs.
type AnomalyDetector struct {
	rules     []DetectionRule
	anomalies []Anomaly

	// Configurable thresholds
	StuckThreshold      int     // Ticks without movement to trigger stuck
	EntityLeakThreshold int     // Max entities before leak warning
	DeathLoopRadius     float64 // Distance to consider same death location
	HealthDrainRate     float64 // Health lost per tick to trigger warning
	BoundsWidth         float64 // Game world width
	BoundsHeight        float64 // Game world height
}

// NewAnomalyDetector creates a detector with default thresholds.
func NewAnomalyDetector() *AnomalyDetector {
	d := &AnomalyDetector{
		StuckThreshold:      120, // 2 seconds at 60fps
		EntityLeakThreshold: 500,
		DeathLoopRadius:     50,
		HealthDrainRate:     0.5,
		BoundsWidth:         1920,
		BoundsHeight:        1080,
	}

	// Add default detection rules
	d.rules = []DetectionRule{
		d.detectStuck,
		d.detectEntityLeak,
		d.detectScoreRegression,
		d.detectHealthDrain,
		d.detectBoundaryViolation,
	}

	return d
}

// AddRule adds a custom detection rule.
func (d *AnomalyDetector) AddRule(rule DetectionRule) {
	d.rules = append(d.rules, rule)
}

// Analyze runs all detection rules on the observation history.
func (d *AnomalyDetector) Analyze(history []Observation) []Anomaly {
	d.anomalies = make([]Anomaly, 0)

	for _, rule := range d.rules {
		found := rule(history)
		d.anomalies = append(d.anomalies, found...)
	}

	return d.anomalies
}

// GetAnomalies returns all detected anomalies.
func (d *AnomalyDetector) GetAnomalies() []Anomaly {
	return d.anomalies
}

// detectStuck checks if player hasn't moved for too long.
func (d *AnomalyDetector) detectStuck(history []Observation) []Anomaly {
	if len(history) < d.StuckThreshold {
		return nil
	}

	var anomalies []Anomaly

	stuckCount := 0

	var lastPos [2]float64

	for i, obs := range history {
		if i == 0 {
			lastPos = obs.State.PlayerPos

			continue
		}

		// Check if position changed
		dx := obs.State.PlayerPos[0] - lastPos[0]
		dy := obs.State.PlayerPos[1] - lastPos[1]
		moved := math.Sqrt(dx*dx+dy*dy) > 1.0

		if !moved {
			stuckCount++
			if stuckCount == d.StuckThreshold {
				anomalies = append(anomalies, Anomaly{
					Type:     AnomalyStuck,
					Severity: SeverityMedium,
					Tick:     obs.Tick,
					Description: fmt.Sprintf("Player stuck at (%.1f, %.1f) for %d ticks",
						obs.State.PlayerPos[0], obs.State.PlayerPos[1], d.StuckThreshold),
				})
			}
		} else {
			stuckCount = 0
			lastPos = obs.State.PlayerPos
		}
	}

	return anomalies
}

// detectEntityLeak checks for unbounded entity growth.
func (d *AnomalyDetector) detectEntityLeak(history []Observation) []Anomaly {
	if len(history) < 10 {
		return nil
	}

	var anomalies []Anomaly

	// Check if entity count exceeds threshold
	last := history[len(history)-1]
	if last.State.EntityCount > d.EntityLeakThreshold {
		anomalies = append(anomalies, Anomaly{
			Type:     AnomalyEntityLeak,
			Severity: SeverityHigh,
			Tick:     last.Tick,
			Description: fmt.Sprintf("Entity count (%d) exceeds threshold (%d)",
				last.State.EntityCount, d.EntityLeakThreshold),
		})
	}

	// Check for consistent growth trend
	if len(history) >= 100 {
		start := history[len(history)-100].State.EntityCount

		end := last.State.EntityCount
		if end > start*2 && end > 100 { // Doubled and significant
			anomalies = append(anomalies, Anomaly{
				Type:        AnomalyEntityLeak,
				Severity:    SeverityMedium,
				Tick:        last.Tick,
				Description: fmt.Sprintf("Entity count doubled in 100 ticks: %d → %d", start, end),
			})
		}
	}

	return anomalies
}

// detectScoreRegression checks if score decreased.
func (d *AnomalyDetector) detectScoreRegression(history []Observation) []Anomaly {
	if len(history) < 2 {
		return nil
	}

	var anomalies []Anomaly

	for i := 1; i < len(history); i++ {
		prev := history[i-1].State.Score
		curr := history[i].State.Score

		if curr < prev {
			anomalies = append(anomalies, Anomaly{
				Type:        AnomalyScoreRegression,
				Severity:    SeverityLow,
				Tick:        history[i].Tick,
				Description: fmt.Sprintf("Score decreased: %d → %d", prev, curr),
			})
		}
	}

	return anomalies
}

// detectHealthDrain checks for rapid health loss.
func (d *AnomalyDetector) detectHealthDrain(history []Observation) []Anomaly {
	if len(history) < 10 {
		return nil
	}

	var anomalies []Anomaly

	// Check last 60 ticks
	window := min(len(history), 60)

	recent := history[len(history)-window:]
	startHealth := recent[0].State.PlayerHealth[0]
	endHealth := recent[len(recent)-1].State.PlayerHealth[0]

	healthLost := float64(startHealth - endHealth)
	drainRate := healthLost / float64(window)

	if drainRate > d.HealthDrainRate {
		anomalies = append(anomalies, Anomaly{
			Type:     AnomalyHealthDrain,
			Severity: SeverityMedium,
			Tick:     recent[len(recent)-1].Tick,
			Description: fmt.Sprintf(
				"Rapid health drain: %.2f HP/tick (lost %d HP)",
				drainRate,
				int(healthLost),
			),
		})
	}

	return anomalies
}

// detectBoundaryViolation checks if player is out of bounds.
func (d *AnomalyDetector) detectBoundaryViolation(history []Observation) []Anomaly {
	var anomalies []Anomaly

	for _, obs := range history {
		pos := obs.State.PlayerPos

		if pos[0] < 0 || pos[0] > d.BoundsWidth ||
			pos[1] < 0 || pos[1] > d.BoundsHeight {
			anomalies = append(anomalies, Anomaly{
				Type:        AnomalyBoundaryViolation,
				Severity:    SeverityHigh,
				Tick:        obs.Tick,
				Description: fmt.Sprintf("Player out of bounds at (%.1f, %.1f)", pos[0], pos[1]),
			})
		}
	}

	return anomalies
}

// GenerateReport creates a markdown report of anomalies.
func (d *AnomalyDetector) GenerateReport() string {
	var sb strings.Builder
	sb.WriteString("# Anomaly Detection Report\n\n")

	if len(d.anomalies) == 0 {
		sb.WriteString("✅ **No anomalies detected**\n")

		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("⚠️ **Found %d anomalies**\n\n", len(d.anomalies)))
	sb.WriteString("| Tick | Type | Severity | Description |\n")
	sb.WriteString("|------|------|----------|-------------|\n")

	for _, a := range d.anomalies {
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
			a.Tick, a.Type, a.Severity.String(), a.Description))
	}

	return sb.String()
}
