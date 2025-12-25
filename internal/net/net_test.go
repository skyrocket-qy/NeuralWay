package net

import (
	"testing"
)

func TestMessageEncodeDecode(t *testing.T) {
	msg := &Message{
		Type:     MsgStateUpdate,
		Tick:     12345,
		ClientID: 42,
		Sequence: 100,
		Payload:  []byte("test payload"),
	}

	data := Encode(msg)

	decoded, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	if decoded.Type != msg.Type {
		t.Errorf("Type mismatch: got %d, want %d", decoded.Type, msg.Type)
	}

	if decoded.Tick != msg.Tick {
		t.Errorf("Tick mismatch: got %d, want %d", decoded.Tick, msg.Tick)
	}

	if decoded.ClientID != msg.ClientID {
		t.Errorf("ClientID mismatch: got %d, want %d", decoded.ClientID, msg.ClientID)
	}

	if string(decoded.Payload) != string(msg.Payload) {
		t.Errorf("Payload mismatch")
	}
}

func TestNewRPCMessage(t *testing.T) {
	msg := NewRPCMessage("spawn_enemy", []byte{1, 2, 3})

	method, args, err := ParseRPC(msg)
	if err != nil {
		t.Fatalf("ParseRPC error: %v", err)
	}

	if method != "spawn_enemy" {
		t.Errorf("Method mismatch: got %s", method)
	}

	if len(args) != 3 {
		t.Errorf("Args length mismatch: got %d", len(args))
	}
}

func TestNetworkDebugPresets(t *testing.T) {
	nd := NewNetworkDebug()

	nd.PresetGood()

	if nd.LatencyMs != 20 {
		t.Error("Good preset latency wrong")
	}

	nd.PresetPoor()

	if nd.LatencyMs != 200 {
		t.Error("Poor preset latency wrong")
	}
}

func TestStatsRecord(t *testing.T) {
	s := NewStats()

	s.RecordSend(100)
	s.RecordSend(50)
	s.RecordReceive(200)

	if s.MessagesSent != 2 {
		t.Errorf("MessagesSent should be 2, got %d", s.MessagesSent)
	}

	if s.BytesSent != 150 {
		t.Errorf("BytesSent should be 150, got %d", s.BytesSent)
	}

	if s.MessagesReceived != 1 {
		t.Errorf("MessagesReceived should be 1, got %d", s.MessagesReceived)
	}
}

func TestStatsPingAverage(t *testing.T) {
	s := NewStats()

	s.RecordPing(100)
	s.RecordPing(200)
	s.RecordPing(150)

	if s.AverageLatencyMs != 150 {
		t.Errorf("Average latency should be 150, got %.1f", s.AverageLatencyMs)
	}
}
