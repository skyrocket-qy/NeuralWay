package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"time"
)

// InputValidator validates player inputs.
type InputValidator struct {
	maxSpeed       float64
	maxAccel       float64
	positionBounds [4]float64 // minX, minY, maxX, maxY
	lastPositions  map[uint32]positionRecord
	violations     map[uint32]int
}

type positionRecord struct {
	x, y      float64
	timestamp time.Time
}

// NewInputValidator creates a new input validator.
func NewInputValidator() *InputValidator {
	return &InputValidator{
		maxSpeed:       500,  // units per second
		maxAccel:       1000, // units per second²
		positionBounds: [4]float64{-10000, -10000, 10000, 10000},
		lastPositions:  make(map[uint32]positionRecord),
		violations:     make(map[uint32]int),
	}
}

// SetSpeedLimit sets the maximum allowed speed.
func (v *InputValidator) SetSpeedLimit(speed float64) {
	v.maxSpeed = speed
}

// SetBounds sets the allowed position bounds.
func (v *InputValidator) SetBounds(minX, minY, maxX, maxY float64) {
	v.positionBounds = [4]float64{minX, minY, maxX, maxY}
}

// ValidatePosition checks if a position update is valid.
func (v *InputValidator) ValidatePosition(clientID uint32, x, y float64) ValidationResult {
	result := ValidationResult{Valid: true}
	now := time.Now()

	// Check bounds
	if x < v.positionBounds[0] || x > v.positionBounds[2] ||
		y < v.positionBounds[1] || y > v.positionBounds[3] {
		result.Valid = false
		result.Reason = "position out of bounds"

		v.recordViolation(clientID)

		return result
	}

	// Check speed
	if last, ok := v.lastPositions[clientID]; ok {
		dt := now.Sub(last.timestamp).Seconds()
		if dt > 0 {
			dx := x - last.x
			dy := y - last.y

			speed := (dx*dx + dy*dy) / (dt * dt)
			if speed > v.maxSpeed*v.maxSpeed {
				result.Valid = false
				result.Reason = "speed violation"

				v.recordViolation(clientID)

				return result
			}
		}
	}

	// Update last position
	v.lastPositions[clientID] = positionRecord{x: x, y: y, timestamp: now}

	return result
}

// ValidationResult contains the result of validation.
type ValidationResult struct {
	Valid  bool
	Reason string
}

// recordViolation records a violation for a client.
func (v *InputValidator) recordViolation(clientID uint32) {
	v.violations[clientID]++
}

// GetViolationCount returns the violation count for a client.
func (v *InputValidator) GetViolationCount(clientID uint32) int {
	return v.violations[clientID]
}

// ResetViolations resets violations for a client.
func (v *InputValidator) ResetViolations(clientID uint32) {
	delete(v.violations, clientID)
}

// StateIntegrityChecker validates game state consistency.
type StateIntegrityChecker struct {
	rules []IntegrityRule
}

// IntegrityRule is a function that checks state validity.
type IntegrityRule func(state map[string]any) (valid bool, reason string)

// NewStateIntegrityChecker creates a new state checker.
func NewStateIntegrityChecker() *StateIntegrityChecker {
	return &StateIntegrityChecker{
		rules: make([]IntegrityRule, 0),
	}
}

// AddRule adds an integrity rule.
func (c *StateIntegrityChecker) AddRule(rule IntegrityRule) {
	c.rules = append(c.rules, rule)
}

// Check validates the game state against all rules.
func (c *StateIntegrityChecker) Check(state map[string]any) []string {
	violations := make([]string, 0)

	for _, rule := range c.rules {
		if valid, reason := rule(state); !valid {
			violations = append(violations, reason)
		}
	}

	return violations
}

// RateLimiter prevents action spam.
type RateLimiter struct {
	limits       map[string]rateLimit
	clientCounts map[uint32]map[string][]time.Time
}

type rateLimit struct {
	count  int
	window time.Duration
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits:       make(map[string]rateLimit),
		clientCounts: make(map[uint32]map[string][]time.Time),
	}
}

// SetLimit sets a rate limit for an action.
func (r *RateLimiter) SetLimit(action string, count int, window time.Duration) {
	r.limits[action] = rateLimit{count: count, window: window}
}

// Allow checks if an action is allowed.
func (r *RateLimiter) Allow(clientID uint32, action string) bool {
	limit, ok := r.limits[action]
	if !ok {
		return true // No limit set
	}

	now := time.Now()
	cutoff := now.Add(-limit.window)

	// Initialize client map if needed
	if r.clientCounts[clientID] == nil {
		r.clientCounts[clientID] = make(map[string][]time.Time)
	}

	// Clean old entries
	times := r.clientCounts[clientID][action]
	validTimes := make([]time.Time, 0)

	for _, t := range times {
		if t.After(cutoff) {
			validTimes = append(validTimes, t)
		}
	}

	// Check limit
	if len(validTimes) >= limit.count {
		return false
	}

	// Record this action
	validTimes = append(validTimes, now)
	r.clientCounts[clientID][action] = validTimes

	return true
}

// SaveEncryptor encrypts and decrypts save data.
type SaveEncryptor struct {
	key []byte
}

// NewSaveEncryptor creates a new save encryptor.
func NewSaveEncryptor(password string) *SaveEncryptor {
	// Derive key from password
	hash := sha256.Sum256([]byte(password))

	return &SaveEncryptor{key: hash[:]}
}

// Encrypt encrypts data.
func (e *SaveEncryptor) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// Decrypt decrypts data.
func (e *SaveEncryptor) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(data) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	return gcm.Open(nil, nonce, ciphertext, nil)
}

// EncryptJSON encrypts a struct to base64.
func (e *SaveEncryptor) EncryptJSON(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	encrypted, err := e.Encrypt(data)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptJSON decrypts base64 to a struct.
func (e *SaveEncryptor) DecryptJSON(encoded string, v any) error {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}

	decrypted, err := e.Decrypt(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(decrypted, v)
}

// Checksum computes SHA-256 checksum.
func Checksum(data []byte) string {
	hash := sha256.Sum256(data)

	return base64.StdEncoding.EncodeToString(hash[:])
}

// ValidateChecksum validates data against a checksum.
func ValidateChecksum(data []byte, expected string) bool {
	return Checksum(data) == expected
}
