//go:build !js || !wasm

package web

import "errors"

// Storage functions for non-WASM platforms (mocks).
type Storage struct{}

func NewStorage() *Storage { return &Storage{} }

func (s *Storage) Save(key, value string) error {
	return errors.New("web storage not supported on this platform")
}

func (s *Storage) Load(key string) (string, error) {
	return "", errors.New("web storage not supported on this platform")
}

func (s *Storage) Delete(key string) error { return nil }
func (s *Storage) Clear() error            { return nil }
