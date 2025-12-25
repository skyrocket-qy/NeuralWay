package security

import (
	"testing"
	"time"
)

func TestInputValidatorBounds(t *testing.T) {
	v := NewInputValidator()
	v.SetBounds(0, 0, 100, 100)

	// Valid position
	result := v.ValidatePosition(1, 50, 50)
	if !result.Valid {
		t.Error("Position should be valid")
	}

	// Out of bounds
	result = v.ValidatePosition(1, 200, 50)
	if result.Valid {
		t.Error("Position should be invalid (out of bounds)")
	}
}

func TestRateLimiter(t *testing.T) {
	r := NewRateLimiter()
	r.SetLimit("attack", 3, 1*time.Second)

	// First 3 should pass
	for i := range 3 {
		if !r.Allow(1, "attack") {
			t.Errorf("Attack %d should be allowed", i+1)
		}
	}

	// 4th should fail
	if r.Allow(1, "attack") {
		t.Error("4th attack should be rate limited")
	}
}

func TestSaveEncryptor(t *testing.T) {
	e := NewSaveEncryptor("test-password")

	original := []byte("secret game data")

	encrypted, err := e.Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}

	decrypted, err := e.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt error: %v", err)
	}

	if string(decrypted) != string(original) {
		t.Error("Decrypted data doesn't match original")
	}
}

func TestChecksum(t *testing.T) {
	data := []byte("game save data")
	checksum := Checksum(data)

	if !ValidateChecksum(data, checksum) {
		t.Error("Checksum validation should pass")
	}

	if ValidateChecksum([]byte("tampered"), checksum) {
		t.Error("Tampered data should fail checksum")
	}
}

func TestBotDetector(t *testing.T) {
	d := NewBotDetector()

	// Record some inputs
	for i := range 20 {
		d.RecordInput(1, "move", float64(i), float64(i))
		time.Sleep(10 * time.Millisecond)
	}

	analysis := d.Analyze(1)

	// Should have some confidence
	if analysis.Confidence == 0 {
		t.Error("Should have some confidence")
	}
}
