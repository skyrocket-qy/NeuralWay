package game

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// SaveData represents a game save.
type SaveData struct {
	Version   int            `json:"version"`
	Name      string         `json:"name"`
	Timestamp int64          `json:"timestamp"`
	PlayTime  float64        `json:"play_time"`
	Data      map[string]any `json:"data"`
	Checksum  string         `json:"checksum"`
}

// NewSaveData creates a new save data container.
func NewSaveData(name string) *SaveData {
	return &SaveData{
		Version:   1,
		Name:      name,
		Timestamp: time.Now().Unix(),
		Data:      make(map[string]any),
	}
}

// Set stores a value in the save data.
func (s *SaveData) Set(key string, value any) {
	s.Data[key] = value
}

// Get retrieves a value from the save data.
func (s *SaveData) Get(key string) (any, bool) {
	val, ok := s.Data[key]

	return val, ok
}

// GetInt retrieves an integer value.
func (s *SaveData) GetInt(key string, defaultVal int) int {
	val, ok := s.Data[key]
	if !ok {
		return defaultVal
	}

	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return defaultVal
	}
}

// GetFloat retrieves a float value.
func (s *SaveData) GetFloat(key string, defaultVal float64) float64 {
	val, ok := s.Data[key]
	if !ok {
		return defaultVal
	}

	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	default:
		return defaultVal
	}
}

// GetString retrieves a string value.
func (s *SaveData) GetString(key, defaultVal string) string {
	val, ok := s.Data[key]
	if !ok {
		return defaultVal
	}

	if str, ok := val.(string); ok {
		return str
	}

	return defaultVal
}

// GetBool retrieves a boolean value.
func (s *SaveData) GetBool(key string, defaultVal bool) bool {
	val, ok := s.Data[key]
	if !ok {
		return defaultVal
	}

	if b, ok := val.(bool); ok {
		return b
	}

	return defaultVal
}

// SaveManager handles save/load operations.
type SaveManager struct {
	SaveDir      string
	CurrentSave  *SaveData
	AutoSaveSlot string
	useChecksum  bool
}

// NewSaveManager creates a save manager.
func NewSaveManager(saveDir string) *SaveManager {
	return &SaveManager{
		SaveDir:      saveDir,
		AutoSaveSlot: "autosave",
		useChecksum:  true,
	}
}

// calculateChecksum generates a checksum for the save data.
func (sm *SaveManager) calculateChecksum(data map[string]any) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	hash := md5.Sum(jsonData)

	return hex.EncodeToString(hash[:])
}

// verifyChecksum verifies the save data integrity.
func (sm *SaveManager) verifyChecksum(save *SaveData) bool {
	if !sm.useChecksum {
		return true
	}

	expected := sm.calculateChecksum(save.Data)

	return save.Checksum == expected
}

// Save saves data to a slot.
func (sm *SaveManager) Save(slot string, save *SaveData) error {
	// Ensure save directory exists
	if err := os.MkdirAll(sm.SaveDir, 0o755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	// Update metadata
	save.Timestamp = time.Now().Unix()
	save.Checksum = sm.calculateChecksum(save.Data)

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(save, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal save data: %w", err)
	}

	// Write to file
	path := filepath.Join(sm.SaveDir, slot+".json")
	if err := os.WriteFile(path, jsonData, 0o644); err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	sm.CurrentSave = save

	return nil
}

// Load loads data from a slot.
func (sm *SaveManager) Load(slot string) (*SaveData, error) {
	path := filepath.Join(sm.SaveDir, slot+".json")

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open save file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	var save SaveData
	if err := json.Unmarshal(data, &save); err != nil {
		return nil, fmt.Errorf("failed to unmarshal save data: %w", err)
	}

	// Verify checksum
	if !sm.verifyChecksum(&save) {
		return nil, errors.New("save file corrupted: checksum mismatch")
	}

	sm.CurrentSave = &save

	return &save, nil
}

// Exists checks if a save slot exists.
func (sm *SaveManager) Exists(slot string) bool {
	path := filepath.Join(sm.SaveDir, slot+".json")
	_, err := os.Stat(path)

	return err == nil
}

// Delete removes a save slot.
func (sm *SaveManager) Delete(slot string) error {
	path := filepath.Join(sm.SaveDir, slot+".json")

	return os.Remove(path)
}

// ListSaves returns all available save slots.
func (sm *SaveManager) ListSaves() ([]string, error) {
	entries, err := os.ReadDir(sm.SaveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}

		return nil, err
	}

	var slots []string

	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			slots = append(slots, e.Name()[:len(e.Name())-5])
		}
	}

	return slots, nil
}

// GetSaveInfo returns metadata about a save without loading full data.
func (sm *SaveManager) GetSaveInfo(slot string) (name string, timestamp int64, playTime float64, err error) {
	save, err := sm.Load(slot)
	if err != nil {
		return "", 0, 0, err
	}

	return save.Name, save.Timestamp, save.PlayTime, nil
}

// AutoSave saves to the autosave slot.
func (sm *SaveManager) AutoSave(save *SaveData) error {
	return sm.Save(sm.AutoSaveSlot, save)
}

// LoadAutoSave loads from the autosave slot.
func (sm *SaveManager) LoadAutoSave() (*SaveData, error) {
	return sm.Load(sm.AutoSaveSlot)
}

// HasAutoSave returns true if an autosave exists.
func (sm *SaveManager) HasAutoSave() bool {
	return sm.Exists(sm.AutoSaveSlot)
}

// QuickSave saves to a numbered slot.
func (sm *SaveManager) QuickSave(save *SaveData, slotNum int) error {
	slot := fmt.Sprintf("quicksave_%d", slotNum)

	return sm.Save(slot, save)
}

// QuickLoad loads from a numbered slot.
func (sm *SaveManager) QuickLoad(slotNum int) (*SaveData, error) {
	slot := fmt.Sprintf("quicksave_%d", slotNum)

	return sm.Load(slot)
}
