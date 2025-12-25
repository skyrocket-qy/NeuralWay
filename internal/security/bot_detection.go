package security

import (
	"math"
	"strings"
	"time"
)

// BotDetector analyzes player behavior to detect bots.
type BotDetector struct {
	inputHistory    map[uint32][]inputRecord
	timingHistory   map[uint32][]float64
	suspicionScores map[uint32]float64
	thresholds      BotDetectionThresholds
}

type inputRecord struct {
	action    string
	timestamp time.Time
	x, y      float64
}

// BotDetectionThresholds configures detection sensitivity.
type BotDetectionThresholds struct {
	MinReactionTime    float64 // Minimum human reaction time (ms)
	MaxInputRegularity float64 // Maximum regularity score (0-1)
	SuspicionThreshold float64 // Score threshold for flagging
	HistorySize        int     // Number of inputs to analyze
}

// DefaultThresholds returns default detection thresholds.
func DefaultThresholds() BotDetectionThresholds {
	return BotDetectionThresholds{
		MinReactionTime:    50,   // 50ms minimum
		MaxInputRegularity: 0.95, // 95% regularity is suspicious
		SuspicionThreshold: 0.7,
		HistorySize:        100,
	}
}

// NewBotDetector creates a new bot detector.
func NewBotDetector() *BotDetector {
	return &BotDetector{
		inputHistory:    make(map[uint32][]inputRecord),
		timingHistory:   make(map[uint32][]float64),
		suspicionScores: make(map[uint32]float64),
		thresholds:      DefaultThresholds(),
	}
}

// SetThresholds sets detection thresholds.
func (d *BotDetector) SetThresholds(t BotDetectionThresholds) {
	d.thresholds = t
}

// RecordInput records a player input for analysis.
func (d *BotDetector) RecordInput(clientID uint32, action string, x, y float64) {
	now := time.Now()

	// Record input
	history := d.inputHistory[clientID]
	history = append(history, inputRecord{action: action, timestamp: now, x: x, y: y})

	// Keep history bounded
	if len(history) > d.thresholds.HistorySize {
		history = history[1:]
	}

	d.inputHistory[clientID] = history

	// Record timing between inputs
	if len(history) >= 2 {
		prev := history[len(history)-2]
		dt := now.Sub(prev.timestamp).Seconds() * 1000 // ms

		timings := d.timingHistory[clientID]

		timings = append(timings, dt)
		if len(timings) > d.thresholds.HistorySize {
			timings = timings[1:]
		}

		d.timingHistory[clientID] = timings
	}
}

// Analyze analyzes a player's behavior.
func (d *BotDetector) Analyze(clientID uint32) BotAnalysis {
	analysis := BotAnalysis{
		ClientID: clientID,
	}

	timings := d.timingHistory[clientID]
	if len(timings) < 10 {
		analysis.Confidence = 0

		return analysis
	}

	// Check for inhuman reaction times
	minTiming := math.MaxFloat64
	for _, t := range timings {
		if t < minTiming {
			minTiming = t
		}
	}

	if minTiming < d.thresholds.MinReactionTime {
		analysis.Flags = append(analysis.Flags, "inhuman_reaction_time")
		analysis.SuspicionScore += 0.3
	}

	// Check for robotic regularity (low variance in timing)
	regularity := d.computeRegularity(timings)
	if regularity > d.thresholds.MaxInputRegularity {
		analysis.Flags = append(analysis.Flags, "robotic_timing")
		analysis.SuspicionScore += 0.4
	}

	analysis.TimingRegularity = regularity

	// Check for pattern repetition
	inputs := d.inputHistory[clientID]
	if patternScore := d.detectPatterns(inputs); patternScore > 0.8 {
		analysis.Flags = append(analysis.Flags, "repeated_patterns")
		analysis.SuspicionScore += 0.3
	}

	// Compute confidence based on sample size
	analysis.Confidence = math.Min(float64(len(timings))/100.0, 1.0)

	// Store suspicion score
	d.suspicionScores[clientID] = analysis.SuspicionScore

	// Flag if over threshold
	analysis.IsSuspicious = analysis.SuspicionScore >= d.thresholds.SuspicionThreshold

	return analysis
}

// BotAnalysis contains the results of bot analysis.
type BotAnalysis struct {
	ClientID         uint32
	IsSuspicious     bool
	SuspicionScore   float64 // 0-1
	Confidence       float64 // 0-1
	TimingRegularity float64
	Flags            []string
}

// computeRegularity calculates how regular the timing intervals are.
func (d *BotDetector) computeRegularity(timings []float64) float64 {
	if len(timings) < 2 {
		return 0
	}

	// Compute mean
	var sum float64
	for _, t := range timings {
		sum += t
	}

	mean := sum / float64(len(timings))

	// Compute standard deviation
	var variance float64

	for _, t := range timings {
		diff := t - mean
		variance += diff * diff
	}

	variance /= float64(len(timings))
	stddev := math.Sqrt(variance)

	// Coefficient of variation (inverted for regularity)
	if mean == 0 {
		return 0
	}

	cv := stddev / mean
	regularity := 1.0 - math.Min(cv, 1.0)

	return regularity
}

// detectPatterns detects repeated input patterns.
func (d *BotDetector) detectPatterns(inputs []inputRecord) float64 {
	if len(inputs) < 10 {
		return 0
	}

	// Simple pattern detection: count repeated action sequences
	patternLen := 3
	patterns := make(map[string]int)

	for i := 0; i <= len(inputs)-patternLen; i++ {
		pattern := ""

		var patternSb188 strings.Builder
		for j := range patternLen {
			patternSb188.WriteString(inputs[i+j].action + ",")
		}

		pattern += patternSb188.String()

		patterns[pattern]++
	}

	// Find most common pattern
	maxCount := 0
	for _, count := range patterns {
		if count > maxCount {
			maxCount = count
		}
	}

	// Return ratio of most common pattern
	total := len(inputs) - patternLen + 1
	if total <= 0 {
		return 0
	}

	return float64(maxCount) / float64(total)
}

// GetSuspicionScore returns the current suspicion score.
func (d *BotDetector) GetSuspicionScore(clientID uint32) float64 {
	return d.suspicionScores[clientID]
}

// ClearHistory clears history for a client.
func (d *BotDetector) ClearHistory(clientID uint32) {
	delete(d.inputHistory, clientID)
	delete(d.timingHistory, clientID)
	delete(d.suspicionScores, clientID)
}
