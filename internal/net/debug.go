package net

import (
	"math/rand"
	"time"
)

// NetworkDebug provides utilities for testing network conditions.
type NetworkDebug struct {
	// LatencyMs adds artificial latency in milliseconds
	LatencyMs int

	// LatencyJitterMs adds random jitter to latency
	LatencyJitterMs int

	// PacketLoss is the probability (0.0-1.0) of dropping a packet
	PacketLoss float64

	// Enabled toggles all debug features
	Enabled bool
}

// NewNetworkDebug creates a new network debug utility.
func NewNetworkDebug() *NetworkDebug {
	return &NetworkDebug{
		Enabled: false,
	}
}

// Apply applies network debug conditions to a message send operation.
// Returns true if the message should be sent, false if dropped.
func (d *NetworkDebug) Apply() bool {
	if !d.Enabled {
		return true
	}

	// Simulate packet loss
	if d.PacketLoss > 0 && rand.Float64() < d.PacketLoss {
		return false
	}

	// Simulate latency
	if d.LatencyMs > 0 || d.LatencyJitterMs > 0 {
		latency := d.LatencyMs
		if d.LatencyJitterMs > 0 {
			latency += rand.Intn(d.LatencyJitterMs)
		}

		time.Sleep(time.Duration(latency) * time.Millisecond)
	}

	return true
}

// SetConditions sets multiple network conditions at once.
func (d *NetworkDebug) SetConditions(latencyMs, jitterMs int, packetLoss float64) {
	d.LatencyMs = latencyMs
	d.LatencyJitterMs = jitterMs
	d.PacketLoss = packetLoss
	d.Enabled = true
}

// PresetGood simulates good network conditions.
func (d *NetworkDebug) PresetGood() {
	d.SetConditions(20, 5, 0.0)
}

// PresetAverage simulates average network conditions.
func (d *NetworkDebug) PresetAverage() {
	d.SetConditions(80, 20, 0.01)
}

// PresetPoor simulates poor network conditions.
func (d *NetworkDebug) PresetPoor() {
	d.SetConditions(200, 100, 0.05)
}

// PresetTerrible simulates terrible network conditions.
func (d *NetworkDebug) PresetTerrible() {
	d.SetConditions(500, 200, 0.15)
}

// Disable turns off network debug.
func (d *NetworkDebug) Disable() {
	d.Enabled = false
}

// Stats tracks network statistics.
type Stats struct {
	MessagesSent     uint64
	MessagesReceived uint64
	BytesSent        uint64
	BytesReceived    uint64
	MessagesDropped  uint64
	AverageLatencyMs float64
	pingHistory      []int64
}

// NewStats creates a new stats tracker.
func NewStats() *Stats {
	return &Stats{
		pingHistory: make([]int64, 0, 100),
	}
}

// RecordSend records a sent message.
func (s *Stats) RecordSend(bytes int) {
	s.MessagesSent++
	s.BytesSent += uint64(bytes)
}

// RecordReceive records a received message.
func (s *Stats) RecordReceive(bytes int) {
	s.MessagesReceived++
	s.BytesReceived += uint64(bytes)
}

// RecordDrop records a dropped message.
func (s *Stats) RecordDrop() {
	s.MessagesDropped++
}

// RecordPing records a ping round-trip time.
func (s *Stats) RecordPing(latencyMs int64) {
	s.pingHistory = append(s.pingHistory, latencyMs)
	if len(s.pingHistory) > 100 {
		s.pingHistory = s.pingHistory[1:]
	}

	// Update average
	var sum int64
	for _, l := range s.pingHistory {
		sum += l
	}

	s.AverageLatencyMs = float64(sum) / float64(len(s.pingHistory))
}

// GetBandwidth returns approximate bytes per second (call periodically).
func (s *Stats) GetBandwidth(periodSeconds float64) (sendBps, recvBps float64) {
	sendBps = float64(s.BytesSent) / periodSeconds
	recvBps = float64(s.BytesReceived) / periodSeconds

	return sendBps, recvBps
}
