//go:build js && wasm

package web

import (
	"syscall/js"
)

// Storage handles local storage in the browser.
type Storage struct {
	localStorage js.Value
}

// NewStorage creates a new web storage handler.
func NewStorage() *Storage {
	return &Storage{
		localStorage: js.Global().Get("localStorage"),
	}
}

// Save saves a string to local storage.
func (s *Storage) Save(key, value string) error {
	s.localStorage.Call("setItem", key, value)
	return nil
}

// Load loads a string from local storage.
func (s *Storage) Load(key string) (string, error) {
	val := s.localStorage.Call("getItem", key)
	if val.IsNull() {
		return "", nil
	}
	return val.String(), nil
}

// Delete removes an item from local storage.
func (s *Storage) Delete(key string) error {
	s.localStorage.Call("removeItem", key)
	return nil
}

// Clear clears all local storage.
func (s *Storage) Clear() error {
	s.localStorage.Call("clear")
	return nil
}
